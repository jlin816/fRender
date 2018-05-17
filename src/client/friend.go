package client

import (
	. "common"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const blenderPath = "/Applications/Blender/blender.app/Contents/MacOS/blender"

type RenderFramesArgs struct {
	StartFrame int
	EndFrame   int
	Filename   string
	Frames     []int
}

type Friend struct {
	me               int
	username         string
	port             int
	masterAddr       net.Addr
	requesterConn    net.Conn
	server           net.Listener
	available        bool
	httpClient       *rpc.Client
	rpcServer        net.Listener
	lastJobCompleted int
	Bad              bool
	logger           *log.Logger
}

func initFriend(username string, port int, masterAddr string) *Friend {
	// set up logging

	addr, err := net.ResolveTCPAddr("tcp", masterAddr)
	if err != nil {
		fmt.Printf("Invalid master addr %s", masterAddr)
		panic(err)
	}
	friend := Friend{username: username, port: port, masterAddr: addr}
	f, _ := os.OpenFile(fmt.Sprintf("logs/%v-friend.log", username), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	friend.logger = log.New(f, "", 0)
	friend.registerWithMaster()

	//make local folder
	path := friend.getLocalFilename("")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	// Friends receive RPCs as well, init as a server
	rpc.Register(&friend)
	handler := rpc.NewServer()
	handler.Register(&friend)
	myIP, _ := externalIP()
	ln, err := net.Listen("tcp", fmt.Sprintf("%v:%d", myIP, port+1))
	// friend.logger.Printf("rpc server listening on %v", ln.Addr())
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			cxn, err := ln.Accept()
			if err != nil {
				panic(err)
			}
			// friend.logger.Printf("Server %s accepted connection to %s from %s\n", friend.username, cxn.LocalAddr(), cxn.RemoteAddr())
			go handler.ServeConn(cxn)
		}
	}()

	server, err := net.Listen("tcp", fmt.Sprintf("%v:%d", myIP, port))
	if err != nil {
		panic(err)
	}

	friend.server = server
	friend.lastJobCompleted = 0
	go friend.sendHeartbeatsToMaster()
	go friend.listenServer()

	friend.debug("", "~Initialised friend %v~\n\n", username)

	return &friend
}

func (fr *Friend) debug(color string, message string, args ...interface{}) {
	colorCode := "\033[97m"
	switch color {
	case "default":
		colorCode = "\033[39m"
	case "":
		colorCode = "\033[39m"
	case "red":
		colorCode = "\033[31m"
	case "green":
		colorCode = "\033[32m"
	case "yellow":
		colorCode = "\033[33m"
	case "blue":
		colorCode = "\033[34m"
	case "magenta":
		colorCode = "\033[35m"
	case "cyan":
		colorCode = "\033[36m"
	case "gray":
		colorCode = "\033[90m"
	case "lgray":
		colorCode = "\033[37m"
	case "lred":
		colorCode = "\033[91m"
	}
	msg := fmt.Sprintf(colorCode+message+"\033[0m\n", args...)
	fr.logger.Print(msg)
}

func (fr *Friend) logImg(filename string) {

	if filename == "" {
		filename = "../figure1.png"
	} else {
		filename = fr.getLocalFilename(filename)
	}
	fr.debug("cyan", "\nRendered file %v", filename)
	content, err := ioutil.ReadFile(filename)
	str := base64.StdEncoding.EncodeToString(content)
	if err != nil {
		panic(err)
	}
	start := "\033]1337;File=inline=1;width=25%;preserveAspectRatio=0:"
	fr.logger.Print(start + str + "\a\n\n")
}

func (fr *Friend) listenServer() {
	for {
		conn, err := fr.server.Accept()
		if err != nil {
			os.Exit(1)
		}
		fr.requesterConn = conn
		// fr.receiveFile(conn)
		// do something
	}
}

func fillString(returnString string, toLength int) string {
	// from http://www.mrwaggel.be/post/golang-transfer-a-file-over-a-tcp-socket/
	for {
		lengthString := len(returnString)
		if lengthString < toLength {
			returnString = returnString + ":"
			continue
		}
		break
	}
	return returnString
}

func (fr *Friend) sendFile(connection net.Conn, filename string) {
	// from http://www.mrwaggel.be/post/golang-transfer-a-file-over-a-tcp-socket/
	fr.debug("red", "Sending file %v\n", filename)
	filename = fr.getLocalFilename(filename)
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
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
	fr.debug("red", "Sent file %v\n", filename)
	return
}

func (fr *Friend) receiveFile(connection net.Conn) { // maybe want port as argument
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")
	fileName = fr.getLocalFilename(fileName)
	newFile, err := os.Create(fileName)
	fr.debug("green", "File received from requester: %v\n\n", fileName)

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

func (fr *Friend) registerWithMaster() {
	httpClient, err := rpc.DialHTTP("tcp", fr.masterAddr.String())
	if err != nil {
		fmt.Println("Couldn't connect friend to master")
		panic(err)
	}

	myIP, err := externalIP()
	if err != nil {
		fmt.Println("Couldn't get IP addr for friend")
		panic(err)
	}
	args := RegisterFriendArgs{Address: fmt.Sprintf("%s:%d", myIP, fr.port), Username: fr.username}
	reply := RegisterFriendReply{}
	err = httpClient.Call("Master.RegisterFriend", args, &reply)
	if err != nil {
		fmt.Printf("Error registering friend: %v", err)
		panic(err)
	}

	fr.httpClient = httpClient

	fr.debug("lgray", "%v registered with master\n", fr.username)
}

func (fr *Friend) renderFrames(file string, frames []int) string {
	relativeFolder := fr.getLocalFilename(fmt.Sprintf("%v_frames_%v", file, fr.username))
	outputFolder, _ := filepath.Abs(relativeFolder)
	outputFiles := outputFolder + "/frame_#####"
	absoluteFilepath, _ := filepath.Abs(fr.getLocalFilename(file))

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

	for _, frame := range frames {
		fr.logImg(fmt.Sprintf("%v_frames_%v/frame_%05d.png", file, fr.username, frame))
	}

	zipCmd := exec.Command("zip", "-rj", relativeFolder+".zip", relativeFolder)
	err1 := zipCmd.Run()
	if err1 != nil {
		panic(err1)
	}
	os.RemoveAll(relativeFolder)
	return fmt.Sprintf("%v_frames_%v", file, fr.username) + ".zip"
}

func (fr *Friend) sendHeartbeatsToMaster() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for _ = range ticker.C {
		args := HeartbeatArgs{Available: fr.available, Username: fr.username, LastJobCompleted: fr.lastJobCompleted}
		reply := HeartbeatReply{}
		err := fr.httpClient.Call("Master.Heartbeat", args, &reply)
		if err != nil {
			fmt.Printf("Error sending heartbeats to master: %v", err)
			panic(err)
		}
	}
}

func (fr *Friend) getLocalFilename(filename string) string {
	return "files/" + fr.username + "_friend/" + filename
}

func arrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

// From https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

////// PUBLIC METHODS ///////

func (fr *Friend) RenderFrames(args RenderFramesArgs, reply *string) error {
	fr.debug("cyan", "Begin rendering frames! \nFrames to render: %v\n", args.Frames)
	var file string
	if fr.Bad {
		file = fr.badRenderFrames(args.Filename, args.Frames)
	} else {
		file = fr.renderFrames(args.Filename, args.Frames)
	}
	fr.debug("cyan", "Frames rendered \n")
	fr.sendFile(fr.requesterConn, file)
	os.RemoveAll(fr.getLocalFilename(file))
	fr.lastJobCompleted++
	*reply = file
	return nil
}

func (fr *Friend) MarkAsUnavailable(args int, reply *string) error {
	fr.available = false
	*reply = fr.username
	return nil
}

func (fr *Friend) MarkAsAvailable(args int, reply *int) error {
	fr.available = true
	return nil
}
func (fr *Friend) ReceiveFile(args int, reply *int) error {
	fr.receiveFile(fr.requesterConn)
	return nil
}
