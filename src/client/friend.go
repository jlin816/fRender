package client

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//const masterAddress = "hello_world"
const blenderPath = "/Applications/Blender/blender.app/Contents/MacOS/blender"

type RenderFramesArgs struct {
	StartFrame int
	EndFrame   int
	Filename   string
}

type Friend struct {
	me            int
	username      string
	port          int
	masterConn    net.Conn
	requesterConn net.Conn
	server        net.Listener
	rpcServer     net.Listener
	Busy          bool
}

func initFriend(username string, port int) *Friend {
	friend := Friend{port: port, username: username}

	//make local folder
	path := friend.getLocalFilename("")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	friend.registerWithMaster()
	rpc.Register(&friend)

	// Hacky stuff from https://github.com/golang/go/issues/13395
	// oldMux := http.DefaultServeMux
	// mux := http.NewServeMux()
	// http.DefaultServeMux = mux

	rpc.HandleHTTP()

	// http.DefaultServeMux = oldMux

	rpcserver, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port+1))
	if err != nil {
		os.Exit(1)
	}
	friend.rpcServer = rpcserver
	go http.Serve(friend.rpcServer, nil)

	server, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		os.Exit(1)
	}
	friend.server = server

	go friend.listenMaster()
	go friend.sendHeartbeatsToMaster()
	go friend.listenServer()

	return &friend
}

func (fr *Friend) listenMaster() {
	for {
		message := make([]byte, 4096)
		length, err := fr.masterConn.Read(message)
		if err != nil {
			fr.masterConn.Close()
			fmt.Printf("error")
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))
		}
	}
}

func (fr *Friend) listenServer() {
	for {
		conn, err := fr.server.Accept()
		if err != nil {
			os.Exit(1)
		}
		fr.requesterConn = conn
		fr.receiveFile(conn)
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
	filename = fr.getLocalFilename(filename)
	file, err := os.Open(filename)
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

func (fr *Friend) receiveFile(connection net.Conn) { // maybe want port as argument
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)
	fmt.Printf("file received\n")

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")
	fileName = fr.getLocalFilename(fileName)
	newFile, err := os.Create(fileName)

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
	connection, err := net.Dial("tcp", "localhost:3333") // TODO: Update port
	if err != nil {
		panic(err)
	}
	fr.masterConn = connection
	fmt.Printf("friend registered w/master\n")
}

func (fr *Friend) renderFrames(file string, start_frame int, end_frame int) string {
	// blender -b bob_lamp_update_export.blend -s 0 -e 100 -o render_files/frame_##### -a
	relativeFolder := fr.getLocalFilename(fmt.Sprintf("%v_frames_%v", file, fr.username))
	outputFolder, _ := filepath.Abs(relativeFolder)
	outputFiles := outputFolder + "/frame_#####"
	absoluteFilepath, _ := filepath.Abs(fr.getLocalFilename(file))

	args := []string{
		"-b",
		absoluteFilepath,
		"-F",
		"PNG",
		"-s",
		fmt.Sprint(start_frame),
		"-e",
		fmt.Sprint(end_frame),
		"-o",
		outputFiles,
		"-a",
	}

	blenderCmd := exec.Command(blenderPath, args...)
	_, err := blenderCmd.Output()
	if err != nil {
		panic(err)
	}

	zipCmd := exec.Command("zip", "-rj", relativeFolder+".zip", relativeFolder)
	_, err1 := zipCmd.Output()
	if err1 != nil {
		panic(err1)
	}
	return fmt.Sprintf("%v_frames_%v", file, fr.username) + ".zip"
}

func (fr *Friend) sendHeartbeatsToMaster() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for _ = range ticker.C {
		heartbeatMessage := []byte(fmt.Sprintf("%v", fr.Busy))
		fr.masterConn.Write(heartbeatMessage)
	}
}

func (fr *Friend) RenderFrames(args RenderFramesArgs, reply *string) error {
	fmt.Printf("rendering frames\n")
	file := fr.renderFrames(args.Filename, args.StartFrame, args.EndFrame)
	fr.sendFile(fr.requesterConn, file)
	fmt.Println(file)
	*reply = file
	return nil
}

func (fr *Friend) getLocalFilename(filename string) string {
	return "files/" + fr.username + "_friend/" + filename
}
