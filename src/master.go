package main

import (
    "fmt"
	"net"
    "os"
)

const (
    host = "localhost"
    port = "3333"
    address = host + ":" + port
    connType = "tcp"
)

type Master struct {
	friends map[Friend]net.Conn
}

type FriendData struct {
	id     int
	socket net.Conn
}

type StartJobReply struct {
	Friends []FriendData `json:"friends"`
}

func initMaster() (mr *Master) {
    return &Master{}
}

func (mr *Master) registerFriend(conn net.Conn) {
	// Register a new user for the first time.
    fmt.Println("Connected a new friend!")
}

func (mr *Master) StartJob(numFriends int) *StartJobReply {
	// Returns active friends allocated to this job
    return &StartJobReply{}
}

func (mr *Master) heartbeat() {
}

func main() {
    mr := initMaster()

    // Start listening for new friends
    listener, err := net.Listen(connType, address)
    if err != nil {
        fmt.Printf("Error listening: %v", err)
        os.Exit(1)
    }

    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Error accepting: %v\n", err)
            os.Exit(1)
        }

        go mr.registerFriend(conn)
    }
}
