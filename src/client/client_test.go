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
		if err != nil {
			fmt.Printf("Error accepting: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("connection accepted: #+v", conn)
		counter += 1
		if counter == 2 {
			break
		}
	}
}

func TestSendFile(t *testing.T) {

}
