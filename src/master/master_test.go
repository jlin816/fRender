package master

import (
	"fmt"
	"net"
	"testing"
	"time"
    "net/rpc"
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
	go startMaster()
	go connectHTTPClient()
    go connectHTTPClient()
	time.Sleep(5000 * time.Millisecond)
}

func TestStartJob(t *testing.T) {
    go main()
    c1 := connectHTTPClient()
    // Test StartJob RPC can be called
    args := StartJobArgs{NumFriends: 1}
    reply := StartJobReply{}
    fmt.Println("here")
    err := c1.Call("Master.StartJob", args, &reply)
    if err != nil {
        fmt.Printf("Error calling StartJob: %v", err)
    }
    fmt.Printf("Success")

    // Test that if there are enough friends our request is fulfilled

    // TODO Test that if there are not enough friends 

    // Test that inactive friends are not returned
}

func connectHTTPClient() *rpc.Client {
    client, err := rpc.DialHTTP("tcp", "localhost:3333")
    fmt.Println("done")
	if err != nil {
		fmt.Printf("Error connecting: %v", err)
		return nil
	}

    // Register yourself on Master
    args := RegisterFriendArgs{Address: "127.0.0.1:3001"}
    reply := RegisterFriendReply{}
    err = client.Call("Master.RegisterFriend", args, &reply)
    if err != nil {
		fmt.Printf("Error registering: %v", err)
		return nil
	}
    return client
}

func startMaster() {
    main()
}
