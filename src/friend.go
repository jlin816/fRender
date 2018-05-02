package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

//const masterAddress = "hello_world"
const blenderPath = "blender"

type Friend struct {
	me int
}

func initFriend() (*Friend) {
	friend := Friend{}
	friend.registerWithMaster()
	go friend.listenOnSocket()

	return &friend //help
}

func (fr *Friend) listenOnSocket() {
	// call receiveJob here somewhere??
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

func (fr *Friend) receiveFile() { // maybe want port as argument
  connection, err := net.Dial("tcp", "localhost:27001") // TODO: Update port
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

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

}

func (fr *Friend) receiveJob() {

}

func (fr *Friend) renderFrames(file string, start_frame int, end_frame int) {
	// blender -b bob_lamp_update_export.blend -s 0 -e 100 -o render_files/frame_##### -a

	binary, lookErr := exec.LookPath(blenderPath)
	if lookErr != nil {
		panic(lookErr)
	}

	output_folder := fmt.Sprintf("%v_frames/frame_#####", file)

	args := []string{
		blenderPath,
		fmt.Sprintf("-b %v", file),
		"-F PNG",
		fmt.Sprintf("-s %v", start_frame),
		fmt.Sprintf("-e %v", end_frame),
		fmt.Sprintf("-o %v", output_folder),
		"-a",
	}

	env := os.Environ()

	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}
}
