package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const BUFFERSIZE = 1024

type FriendData struct {
	id   int
	conn net.Conn
}

type Requester struct {
	me         int
	friends    []FriendData
	masterConn net.Conn
}

func initRequester() *Requester {
	requester := Requester{}
	requester.registerWithMaster()
	// go requester.listenOnSocket()

	return &requester
}

func (req *Requester) listenOnSocket() {
	// call receiveJob here somewhere??
}

func (req *Requester) sendFile(connection net.Conn, filename string) {
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

func (req *Requester) receiveFile() { // maybe want port as argument
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

func (req *Requester) registerWithMaster() {
	connection, err := net.Dial("tcp", "localhost:3333") // TODO: Update port
	if err != nil {
		panic(err)
	}
	req.masterConn = connection
	fmt.Printf("friend registered w/master")
}

func (req *Requester) connectToFriends() {

}

func (req *Requester) startJob() {
	// for p in peers:
	server, err := net.Listen("tcp", "localhost:27001") // TODO: update with actual port
	if err != nil {
		os.Exit(1)
	}
	defer server.Close()
	for {
		conn, err := server.Accept()
		if err != nil {
			os.Exit(1)
		}
		go req.sendFile(conn, "filename") // TODO: update filename
	}

}

func (req *Requester) getProgress() {

}

func (req *Requester) cancelJob() {

}
