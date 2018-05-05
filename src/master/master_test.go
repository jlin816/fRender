package master

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestReceiveMessage(t *testing.T) {
	go main()
	conn, err := net.Dial("tcp", "localhost:3333")
	if err != nil {
		t.Errorf("Error sending: %v", err)
		return
	}
	defer conn.Close()
}

func TestMultipleReceiveMessage(t *testing.T) {
    killChan := make(chan int)
	go startMaster()
	go connectClient(killChan)
    go connectClient(killChan)
	time.Sleep(5000 * time.Millisecond)
    killChan <- 1
    killChan <- 1
}

func TestStartJob(t *testing.T) {
    killChan := make(chan int)
    go connectClient(killChan)
    go connectClient(killChan)
    // Test that if there are enough friends our request is fulfilled

    // TODO Test that if there are not enough friends 

    // Test that inactive friends are not returned
    killChan <- 1
}

func connectClient(ch chan int) {
	conn, err := net.Dial("tcp", "localhost:3333")
	if err != nil {
		fmt.Printf("Error sending: %v", err)
		return
	}
    <-ch
	conn.Close()
}

func startMaster() {
    main()
}
