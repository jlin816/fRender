package main

import (
    "fmt"
	"net"
    "os"
    "sync"
)

const (
    host = "localhost"
    port = "3333"
    address = host + ":" + port
    connType = "tcp"
)

type Master struct {
    mu      sync.Mutex
	friends map[FriendData]bool // friend info + statuses
}

type FriendData struct {
	id     int
	conn net.Conn
}

type StartJobReply struct {
	Friends []FriendData `json:"friends"`
}

func initMaster() (*Master) {
    mr := &Master{friends: make(map[FriendData]bool)}
    return mr
}

func (mr *Master) registerFriend(conn net.Conn) {
    mr.mu.Lock()
    defer mr.mu.Unlock()

    newFriend := FriendData{id: len(mr.friends), conn: conn}
    mr.friends[newFriend] = true
    fmt.Printf("Connected friend %d!\n", newFriend.id)
}

func (mr *Master) StartJob(numFriends int) *StartJobReply {
	// Returns active friends allocated to this job
    return &StartJobReply{}
}

func (mr *Master) heartbeat() {
}

func main() {
    fmt.Println("In main")
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
