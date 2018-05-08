package master

import (
    "fmt"
	"net"
    "os"
    "sync"
	"time"
	"net/rpc"
	"errors"
	"net/http"
    . "common"
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
    requesters []RequesterData
}

type FriendData struct {
	id			int
	address		net.Addr
//	writer		*bufio.Writer
//	reader		*bufio.Reader
	lastActive	time.Time
	available	bool // alive and not busy
}

type RequesterData struct {
    id          int
}

// ====== RPC METHODS ===========

func (mr *Master) RegisterFriend(args RegisterFriendArgs, reply *RegisterFriendReply) error {
    mr.mu.Lock()
    defer mr.mu.Unlock()

	addr, err := net.ResolveTCPAddr("tcp", args.Address)
	if err != nil {
		return errors.New("Can't resolve friend's TCP addr")
	}

    newFriend := FriendData {
		id: len(mr.friends),
		address: addr,
		available: true,
		lastActive: time.Now(),
	}
    mr.friends = append(mr.friends, newFriend)
    fmt.Printf("Connected friend %d!\n", newFriend.id)
	reply.Success = true
	return nil
}

func (mr *Master) RegisterRequester(args RegisterRequesterArgs, reply *RegisterRequesterReply) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

    newRequester := RequesterData {
        id: len(mr.requesters),
    }
    mr.requesters = append(mr.requesters, newRequester)
    fmt.Printf("Connected new requester %d!\n", newRequester.id)
    reply.Success = true
	return nil
}

func (mr *Master) StartJob(args StartJobArgs, reply *StartJobReply) error {
	fmt.Println("StartJob called")
	mr.mu.Lock()
	defer mr.mu.Unlock()

	friendCount := 0
	assignedFriends := make([]string, args.NumFriends)
	for _, friend := range mr.friends {
		if !friend.available || (time.Since(friend.lastActive) > friendTimeout) {
			continue
		}

		assignedFriends[friendCount] = friend.address.String()
		friendCount++

		if friendCount == args.NumFriends {
			reply.Friends = assignedFriends
			return nil
		}
	}
	return errors.New("Not enough active friends")
}

// ====== PRIVATE METHODS =========

func initMaster() (*Master) {
    mr := &Master{friends: []FriendData{}}
    return mr
}

func main() {
    mr := initMaster()

	rpc.Register(mr)
	rpc.HandleHTTP()


    // Start listening for new friends
    listener, err := net.Listen(connType, address)
    if err != nil {
        fmt.Printf("Error listening: %v", err)
        os.Exit(1)
    }

	fmt.Println("serving")
	http.Serve(listener, nil)
	fmt.Println("past serving")

    defer listener.Close()
	defer fmt.Println("closed")
}
