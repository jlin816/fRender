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
	lastActive	time.Time
	busy		bool
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
}

func (mr *Master) StartJob(numFriends int, requesterConn net.Conn) {
	// Returns active friends allocated to this job
	mr.mu.Lock()
	defer mr.mu.Unlock()

	friendCount := 0
	assignedFriends := []net.Addr{}
	for _, friend := range mr.friends {
		if friend.busy || (time.Since(friend.lastActive) > friendTimeout) {
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

			requesterConn.Write(replyBuffer.Bytes())
			return
		}
	}
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
