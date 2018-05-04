package main

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
	masterConn    net.Conn
	requesterConn net.Conn
	server        net.Listener
	Busy          bool
}

func initFriend() *Friend {
	friend := Friend{}
	friend.registerWithMaster()
	rpc.Register(&friend)
	rpc.HandleHTTP()
	fmt.Printf("xyz")
	server, err := net.Listen("tcp", "localhost:19997") // TODO: update with actual port
	if err != nil {
		os.Exit(1)
	}
	fmt.Printf("abc")
	friend.server = server
	go http.Serve(server, nil)

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
	defer connection.Close()
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
	// connection, err := net.Dial("tcp", "localhost:27001") // TODO: Update port
	// if err != nil {
	// 	panic(err)
	// }
	// defer connection.Close()
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)
	fmt.Printf("file received\n")

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")
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

func (fr *Friend) receiveJob() {

}

func (fr *Friend) renderFrames(file string, start_frame int, end_frame int) {
	// blender -b bob_lamp_update_export.blend -s 0 -e 100 -o render_files/frame_##### -a

	output_folder, _ := filepath.Abs(fmt.Sprintf("%v_frames/frame_#####", file))
	absoluteFilepath, _ := filepath.Abs(file)

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
		output_folder,
		"-a",
	}

	blenderCmd := exec.Command(blenderPath, args...)
	_, err := blenderCmd.Output()
	if err != nil {
		panic(err)
	}

}

func (fr *Friend) sendHeartbeatsToMaster() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for _ = range ticker.C {
		heartbeatMessage := []byte(fmt.Sprintf("%v", fr.Busy))
		fr.masterConn.Write(heartbeatMessage)
	}
}

func (fr *Friend) PrintHello(args int, reply *int) error {
	fmt.Printf("HIYA\n")
	*reply = 100
	return nil
}

func (fr *Friend) RenderFrames(args RenderFramesArgs, reply *int) error {
	fmt.Printf("rendering frames\n")
	fr.renderFrames(args.Filename, args.StartFrame, args.EndFrame)
	return nil
}
