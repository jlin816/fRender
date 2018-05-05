package main

import (
	"fmt"
	"net"
	"os"
	"testing"
)

func TestConnectToMaster(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:3333")
	if err != nil {
		fmt.Printf("Error listening: %v", err)
		os.Exit(1)
	}

	client := initClient()

	_ = client

	defer listener.Close()
	counter := 0
	for {
		conn, err := listener.Accept()
		send(conn)
		go listen(conn)
		if err != nil {
			fmt.Printf("Error accepting: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("connection accepted: %+v\n", conn)
		counter += 1
		// if counter == 2 {
		// 	break
		// }
	}
}

func TestSendFile(t *testing.T) {

}

func listen(conn net.Conn) {
	for {
		message := make([]byte, 4096)
		length, err := conn.Read(message)
		if err != nil {
			conn.Close()
			fmt.Printf("error")
			break
		}
		if length > 0 {
			fmt.Println("MASTER RECEIVED: " + string(message))
		}
	}
	// FAILURE CODE GOES HERE??
}

func send(conn net.Conn) {
	conn.Write([]byte("hello world"))

}
