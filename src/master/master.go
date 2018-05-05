package master

import (
    "fmt"
	"net"
    "os"
    "sync"
	"time"
	"encoding/gob"
	"bytes"
	"log"
	"net/rpc"
)

const (
    host = "localhost"
    port = "3333"
    address = host + ":" + port
    connType = "tcp"
)

const friendTimeout = 200 * time.Millisecond

type Master struct {
    mu      sync.Mutex
	friends []FriendData
}

type FriendData struct {
	id			int
	address		net.Addr
	conn		net.Conn
	writer		*bufio.Writer
	reader		*bufio.Reader
	lastActive	time.Time
	available	bool // alive and not busy
}

func initMaster() (*Master) {
    mr := &Master{friends: []FriendData{}}
    return mr
}

func (mr *Master) registerFriend(conn net.Conn) {
    mr.mu.Lock()
    defer mr.mu.Unlock()

    newFriend := FriendData{
		id: len(mr.friends),
		address: conn.RemoteAddr(),
		conn: conn,
	}
    mr.friends = append(mr.friends, newFriend)
    fmt.Printf("Connected friend %d!\n", newFriend.id)
	go mr.listen(newFriend)
}

func (mr *Master) listen(friend FriendData) {
	for {
		line, err := friend.reader.ReadString("\n")
		if err != nil {
			fmt.Println("Couldn't read from friend: ", err)
			break
		}
		mr.handleMessage(line, friend.conn)
	}

	friend.conn.Close()
	friend.available = false
}

func (mr *Master) handleMessage(message string, conn net.Conn) {
	switch {
	}
}

func (mr *Master) StartJob(numFriends int, requesterConn net.Conn) {
	// Returns active friends allocated to this job
	mr.mu.Lock()
	defer mr.mu.Unlock()

	friendCount := 0
	assignedFriends := []net.Addr{}
	for _, friend := range mr.friends {
		if !friend.available || (time.Since(friend.lastActive) > friendTimeout) {
			continue
		}

		assignedFriends[friendCount] = friend.address
		friendCount++

		if friendCount == numFriends {
			var replyBuffer bytes.Buffer
			enc := gob.NewEncoder(&replyBuffer)
			err := enc.Encode(assignedFriends)
			if err != nil {
				log.Fatal("Encode error:", err)
			}

			break
		}
	}

	requesterConn.Write(replyBuffer.Bytes())
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
