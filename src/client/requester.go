package client

import (
	. "common"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const BUFFERSIZE = 1024

type FriendData struct {
	id   int // currently unused
	conn net.Conn
	rpc  *rpc.Client
}

type Requester struct {
	me         int
	username   string
	port       int
	friends    []FriendData
	masterAddr net.Addr
}

func initRequester(username string, port int, masterAddr string) *Requester {
	addr, err := net.ResolveTCPAddr("tcp", masterAddr)
	if err != nil {
		fmt.Printf("Invalid master addr %s", masterAddr)
		panic(err)
	}
	requester := Requester{username: username, port: port, masterAddr: addr}
	requester.registerWithMaster()

	path := requester.getLocalFilename("")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
	// go requester.listenOnSocket()
	// requester.startJob()

	return &requester
}

func (req *Requester) listenOnSocket() {
	// call receiveJob here somewhere??
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
		fmt.Println("Couldn't register requester with master")
		panic(err)
	}

	args := RegisterRequesterArgs{Username: req.username}
	reply := RegisterFriendReply{}
	err = httpClient.Call("Master.RegisterRequester", args, &reply)
	if err != nil {
		fmt.Printf("Error registering: %v", err)
		panic(err)
	}
	fmt.Printf("Requester registered w/master!!\n")
}

func (req *Requester) connectToFriends(friendAddresses []string) {
	req.friends = make([]FriendData, 0)
	for _, frAddress := range friendAddresses {
		connection, err := net.Dial("tcp", frAddress+":19997") // TODO: Update port
		if err != nil {
			panic(err)
		}
		rpcconn, err := rpc.DialHTTP("tcp", frAddress+":19996")
		if err != nil {
			panic(err)
		}
		req.friends = append(req.friends, FriendData{conn: connection, rpc: rpcconn})
	}
}

// struct Range {
// 	int start
// 	int end
// }
//
// func basicSplitFrames(int numFrames, int numFriends) {
// 	framesPerFriend = math.Ceil(numFrames/numFriends)
// 	frameSplit = make([]Range, numFriends)
// 	for i := 0; i < numFriends-1; i++ {
// 		frameSplit[i] = Range{start: i * framesPerFriend, end: (i+1) * framesPerFriend}
// 	}
// 	frameSplit[numFriends-1] = Range{start: numFriends * framesPerFriend, end: numFrames}
// 	return frameSplit
// }

func (req *Requester) StartJob(filename string) {
	// create folder for output
	outputFolder := req.getLocalFilename(fmt.Sprintf("%v_frames", filename))
	if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
		os.Mkdir(outputFolder, os.ModePerm)
	}

	// get list of friends
	friendAddresses := req.getFriendsFromMaster(1)

	//  connectToFriends
	req.connectToFriends(friendAddresses)

	// determine frame split
	// numFrames := 150 // TODO
	// numFriends := len(req.friends)
	// frameSplit = basicSplitFrames(numFrames, numFriends)

	// send file to each friend
	for _, friend := range req.friends {
		req.sendFile(friend.conn, filename)

		args := RenderFramesArgs{StartFrame: 0, EndFrame: 1, Filename: filename}
		var reply string
		err := friend.rpc.Call("Friend.RenderFrames", args, &reply)
		if err != nil {
			log.Fatal("rpc error:", err)
		}
		fmt.Printf("reply: %v\n", reply)
		req.receiveFile(friend.conn)

		zipCmd := exec.Command("unzip", req.getLocalFilename(reply), "-d", outputFolder)
		fmt.Printf("%v %v %v %v", "unzip", req.getLocalFilename(reply), "-d", outputFolder)
		_, err1 := zipCmd.Output()
		if err1 != nil {
			panic(err1)
		}
	}
	fmt.Println("all frames received...")

}

func (req *Requester) getProgress() {

}

func (req *Requester) cancelJob() {

}

func (req *Requester) getFriendsFromMaster(n int) []string {
	// TODO

	list := make([]string, 0)
	list = append(list, "localhost")

	return list
}

func (req *Requester) getLocalFilename(filename string) string {
	return "files/" + req.username + "_requester/" + filename
}
