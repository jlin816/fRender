package client

import (
	"bytes"
	. "common"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const BUFFERSIZE = 1024

type FriendData struct {
	id   int // currently unused
	conn net.Conn
	rpc  *rpc.Client
}

type Requester struct {
	me               int
	username         string
	friends          []FriendData
	masterAddr       net.Addr
	masterHttpClient *rpc.Client
	mu               sync.Mutex
}

type Tasks struct {
	available    []int
	completed    int
	wg           *sync.WaitGroup
	mu           sync.Mutex
	registerChan chan FriendData
}

func initRequester(username string, masterAddr string) *Requester {
	addr, err := net.ResolveTCPAddr("tcp", masterAddr)
	if err != nil {
		fmt.Printf("Invalid master addr %s", masterAddr)
		panic(err)
	}
	requester := Requester{username: username, masterAddr: addr}
	requester.registerWithMaster()

	path := requester.getLocalFilename("")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
	// go requester.listenOnSocket()
	// requester.startJob()
	fmt.Printf("requester initialised %v\n", username)
	rand.Seed(time.Now().Unix())

	return &requester
}

func (req *Requester) sendFile(connection net.Conn, filename string) {
	// from http://www.mrwaggel.be/post/golang-transfer-a-file-over-a-tcp-socket/
	// defer connection.Close()
	filename = req.getLocalFilename(filename)
	file, err := os.Open(filename)
	fmt.Printf("sending %v\n", filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	// Sending filename and filesize
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := fillString(fileInfo.Name(), 64)
	connection.Write([]byte(fileSize))
	connection.Write([]byte(fileName))
	sendBuffer := make([]byte, BUFFERSIZE)
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	return
}

func (req *Requester) receiveFile(connection net.Conn) {
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")
	fileName = req.getLocalFilename(fileName)
	newFile, err := os.Create(fileName)
	fmt.Printf("received file! %v\n", fileName)

	if err != nil {
		panic(err)
	}
	defer newFile.Close()
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, connection, (fileSize - receivedBytes))
			connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			break
		}
		io.CopyN(newFile, connection, BUFFERSIZE)
		receivedBytes += BUFFERSIZE
	}
}

func (req *Requester) registerWithMaster() {
	httpClient, err := rpc.DialHTTP("tcp", req.masterAddr.String())
	if err != nil {
		fmt.Println("Couldn't connect requester to master")
		panic(err)
	}

	args := RegisterRequesterArgs{Username: req.username}
	reply := RegisterFriendReply{}
	err = httpClient.Call("Master.RegisterRequester", args, &reply)
	if err != nil {
		fmt.Printf("Error registering requester: %v", err)
		panic(err)
	}

	req.masterHttpClient = httpClient
	fmt.Printf("Requester registered w/master!!\n")
}

func (req *Requester) connectToFriends(friendAddresses []string) {
	req.friends = make([]FriendData, 0)
	for _, frAddress := range friendAddresses {

		// connect to TCP file server
		connection, err := net.Dial("tcp", frAddress)
		if err != nil {
			panic(err)
		}
		fmt.Printf("tcp connected to %v\n", frAddress)

		// connect to RPC server (by def'n, port+1 of file server)
		addrParts := strings.Split(frAddress, ":")
		port, _ := strconv.ParseInt(addrParts[1], 0, 64)
		rpcAddr := fmt.Sprintf("%v:%v", addrParts[0], port+1)
		rpcconn, err := rpc.Dial("tcp", rpcAddr)
		if err != nil {
			panic(err)
		}
		fmt.Printf("rpc connected to %v\n", frAddress)

		// mark friend as unavailable
		args := 0
		reply := 0
		err = rpcconn.Call("Friend.MarkAsUnavailable", args, &reply)
		if err != nil {
			panic(err)
		}

		req.friends = append(req.friends, FriendData{conn: connection, rpc: rpcconn})
		fmt.Printf("connected to %v\n", frAddress)
	}
}

func basicSplitFrames(numFrames int, numFriends int) (frameSplit [][]int, verificationFrames [][2]int) {
	framesPerFriend := (numFrames + numFriends - 1) / numFriends
	frameSplit = make([][]int, numFriends)
	verificationFrames = make([][2]int, numFriends)

	friend := -1
	for i := 0; i <= numFrames; i++ {
		if i%framesPerFriend == 0 && friend < (numFriends-1) {
			friend += 1
		}
		frameSplit[friend] = append(frameSplit[friend], i)
	}

	for i := 0; i < numFriends; i++ {
		frame := frameSplit[i][rand.Intn(len(frameSplit[i]))] // pick a random frame from i's work
		verificationFrames[i][1] = frame
		if i == (numFriends - 1) {
			verificationFrames[0][0] = frame
		} else {
			verificationFrames[i+1][0] = frame
		}
	}

	for i := 0; i < numFriends; i++ {
		frameSplit[i] = append(frameSplit[i], verificationFrames[i][0])
	}
	return frameSplit, verificationFrames
}

func (req *Requester) StartJob(filename string, numFrames int, numFriends int) bool {
	fmt.Println("start job...")
	// create folder for output
	outputFolder := req.getLocalFilename(fmt.Sprintf("%v_frames", filename))
	if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
		os.Mkdir(outputFolder, os.ModePerm)
	}

	// create folder for verification
	verFolder := req.getLocalFilename(fmt.Sprintf("%v_frames_%v", filename, "req"))
	if _, err := os.Stat(verFolder); os.IsNotExist(err) {
		os.Mkdir(verFolder, os.ModePerm)
	}

	// get list of friends
	friendAddresses := req.getFriendsFromMaster(numFriends)

	//  connectToFriends
	req.connectToFriends(friendAddresses)

	// build task manager
	var tasks Tasks
	var wg sync.WaitGroup
	tasks.wg = &wg
	tasks.registerChan = make(chan FriendData, len(friendAddresses))
	go func() {
		for _, friend := range req.friends {
			tasks.registerChan <- friend
		}
	}()

	// determine frame split
	frameSplit, verificationFrames := basicSplitFrames(numFrames, numFriends)
	fmt.Println(verificationFrames)
	for i := 0; i < len(frameSplit); i++ {
		tasks.available = append(tasks.available, i)
	}

	// assign tasks to available friends...
	for friend := range tasks.registerChan {
		tasks.mu.Lock()
		if len(tasks.available) > 0 {
			taskNum := tasks.available[0]
			tasks.available = tasks.available[1:]
			tasks.wg.Add(1)
			tasks.mu.Unlock()
			go req.renderFramesOnFriend(filename, friend, frameSplit[taskNum], &tasks, taskNum)
		} else {
			tasks.mu.Unlock()
			fmt.Println("all tasks allocated, waiting...")
			tasks.wg.Wait() //wait for all pending tasks to complete

			if tasks.completed >= len(frameSplit) { // check all tasks succeeded

				//run the verification procedure. If there are bad tasks, then add back to the task manager
				success := req.verifyAllFrames(filename, verificationFrames, &tasks)
				if success {
					fmt.Println("verification complete...")
					break
				} else {
					fmt.Println("verification failed, reassigning tasks...")
					fmt.Println(tasks.available)
				}
			}
		}
	}
	wg.Wait()
	fmt.Println("all frames received...")

	// merge frames :)
	req.mergeFrames(filename, len(frameSplit))

	// code to kill hanging threads, and close up connections
	req.closeConnections()

	return true

}

func (req *Requester) mergeFrames(filename string, nTasks int) {
	outputFolder := req.getLocalFilename(fmt.Sprintf("%v_frames", filename))

	for i := 0; i < nTasks; i++ {
		framesFolder := req.getLocalFilename(fmt.Sprintf("%v_frames_%v/.", filename, i))
		cpCmd := exec.Command("/bin/cp", "-rf", framesFolder, outputFolder)
		err1 := cpCmd.Run()
		if err1 != nil {
			panic(err1)
		}
	}

	// remove all the temp folders we created
	for i := 0; i < nTasks; i++ {
		os.RemoveAll(req.getLocalFilename(fmt.Sprintf("%v_frames_%v", filename, i)))
	}
	os.RemoveAll(req.getLocalFilename(fmt.Sprintf("%v_frames_%v", filename, "req")))
}

func (req *Requester) verifyAllFrames(filename string, verificationFrames [][2]int, tasks *Tasks) bool {
	nTasks := len(verificationFrames)
	badTasks := make(map[int]bool, nTasks)
	success := true

	for i := 0; i < nTasks; i++ {
		if !badTasks[i] { // don't check bad tasks

			// take neighbours on the circle, and open respective frames
			outputFolder1 := req.getLocalFilename(fmt.Sprintf("%v_frames_%v", filename, i))
			j := i + 1
			if j == nTasks {
				j = 0
			}
			outputFolder2 := req.getLocalFilename(fmt.Sprintf("%v_frames_%v", filename, j))
			frame := verificationFrames[i][1]
			pathToFile1 := fmt.Sprintf("%v/frame_%05d.png", outputFolder1, frame)
			pathToFile2 := fmt.Sprintf("%v/frame_%05d.png", outputFolder2, frame)
			fmt.Printf("compare %v and %v\n", pathToFile1, pathToFile2)

			if !verifyFrames(pathToFile1, pathToFile2) {
				success = false

				// render conflict frame ourselves, to test against other friends
				req.renderFrame(filename, frame)
				verFolder := req.getLocalFilename(fmt.Sprintf("%v_frames_%v", filename, "req"))
				pathToFileVer := fmt.Sprintf("%v/frame_%05d.png", verFolder, frame)

				if !verifyFrames(pathToFile1, pathToFileVer) {
					tasks.mu.Lock()
					tasks.available = append(tasks.available, i)
					tasks.completed--
					tasks.mu.Unlock()
					os.RemoveAll(outputFolder1)
					badTasks[i] = true
				}

				if !verifyFrames(pathToFile2, pathToFileVer) {
					tasks.mu.Lock()
					tasks.available = append(tasks.available, j)
					tasks.completed--
					tasks.mu.Unlock()
					os.RemoveAll(outputFolder2)
					badTasks[j] = true
				}
			}
		}
	}
	return success
}

func (req *Requester) closeConnections() {
	var args int
	var reply int
	for _, friend := range req.friends {
		friend.rpc.Call("Friend.MarkAsAvailable", args, &reply)
		friend.conn.Close()
		friend.rpc.Close()
	}
}

func (req *Requester) receiveFileOnFriend(friend *FriendData) {
	var args int
	var reply int
	friend.rpc.Call("Friend.ReceiveFile", args, &reply)
}

func (req *Requester) receiveFileWithWait(connection net.Conn, recvChannel chan bool) {
	req.receiveFile(connection)
	recvChannel <- true
}

func verifyFrames(filepath1 string, filepath2 string) bool {
	chunkSize := 64000
	// from https://stackoverflow.com/questions/29505089/how-can-i-compare-two-files-in-golang

	f1, err := os.Open(filepath1)
	if err != nil {
		log.Fatal(err)
	}

	f2, err := os.Open(filepath2)
	if err != nil {
		log.Fatal(err)
	}

	fi1, _ := f1.Stat()
	fi2, _ := f2.Stat()
	if fi1.Size() != fi2.Size() {
		return false
	}

	for {
		b1 := make([]byte, chunkSize)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true
			} else if err1 == io.EOF || err2 == io.EOF {
				return false
			} else {
				log.Fatal(err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			return false
		}
	}
}

func (req *Requester) renderFramesOnFriend(filename string, friend FriendData, frames []int, tasks *Tasks, taskNum int) {
	success := true

	// create output folder
	outputFolder := req.getLocalFilename(fmt.Sprintf("%v_frames_%v", filename, taskNum))
	if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
		os.Mkdir(outputFolder, os.ModePerm)
	}
	go req.receiveFileOnFriend(&friend)
	req.sendFile(friend.conn, filename)

	args := RenderFramesArgs{Filename: filename}
	args.Frames = frames

	fmt.Println(args)
	var reply string
	receiveChannel := make(chan bool)
	go req.receiveFileWithWait(friend.conn, receiveChannel)
	err := friend.rpc.Call("Friend.RenderFrames", args, &reply)
	if err != nil {
		success = false
		log.Fatal("rpc error:", err)
	}
	fmt.Printf("reply: %v\n", reply)
	_ = <-receiveChannel
	req.mu.Lock()
	zipCmd := exec.Command("unzip", "-n", req.getLocalFilename(reply), "-d", outputFolder)
	fmt.Printf("%v %v %v %v %v", "unzip", "-n", req.getLocalFilename(reply), "-d", outputFolder)
	err1 := zipCmd.Run()
	if err1 != nil {
		panic(err1)
	}
	os.RemoveAll(req.getLocalFilename(reply))
	req.mu.Unlock()

	tasks.mu.Lock()
	if success {
		tasks.completed = tasks.completed + 1 //lock on write
		tasks.mu.Unlock()

	} else {
		tasks.available = append(tasks.available, taskNum)
		tasks.mu.Unlock()
		time.Sleep(100 * time.Millisecond) //stall failed worker
	}
	tasks.wg.Done()
	tasks.registerChan <- friend
}

func (req *Requester) getProgress() {

}

func (req *Requester) cancelJob() {

}

func (req *Requester) getFriendsFromMaster(n int) []string {
	args := StartJobArgs{NumFriends: n, Username: req.username}
	reply := StartJobReply{}

	err := req.masterHttpClient.Call("Master.StartJob", args, &reply)
	if err != nil {
		fmt.Printf("Error calling StartJob to get friends from master: %v", err)
	}
	fmt.Printf("Got friends from master: %v", reply.Friends)

	return reply.Friends
}

func (req *Requester) getLocalFilename(filename string) string {
	return "files/" + req.username + "_requester/" + filename
}

func (req *Requester) renderFrames(file string, frames []int) string {
	relativeFolder := req.getLocalFilename(fmt.Sprintf("%v_frames_%v", file, "req"))
	outputFolder, _ := filepath.Abs(relativeFolder)
	outputFiles := outputFolder + "/frame_#####"
	absoluteFilepath, _ := filepath.Abs(req.getLocalFilename(file))

	args := []string{
		"-b",
		absoluteFilepath,
		"-F",
		"PNG",
		"-o",
		outputFiles,
		"-f",
		arrayToString(frames, ","),
	}

	blenderCmd := exec.Command(blenderPath, args...)
	err := blenderCmd.Run()
	if err != nil {
		panic(err)
	}
	return relativeFolder
}

func (req *Requester) renderFrame(file string, frame int) string {
	frames := make([]int, 0)
	frames = append(frames, frame)
	folder := req.renderFrames(file, frames)

	return fmt.Sprintf("%v/frame_%05d.png", folder, frame)
}
