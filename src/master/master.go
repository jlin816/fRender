package master

import (
    "fmt"
	"net"
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
    server  *http.Server
	friends []FriendData
    requesters []RequesterData
}

type FriendData struct {
	id			int
    username    string
	address		net.Addr
//	writer		*bufio.Writer
//	reader		*bufio.Reader
	lastActive	time.Time
	available	bool // alive and not busy
}

type RequesterData struct {
    id          int
    username    string
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
        username: args.Username,
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
        username: args.Username,
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

// ======= PUBLIC METHODS =========

func NewMaster() *Master {
    mr := &Master{friends: []FriendData{}}

    rpc.Register(mr)

    // Workaround from https://github.com/golang/go/issues/13395
    oldMux := http.DefaultServeMux
    mux := http.NewServeMux()
    http.DefaultServeMux = mux

	rpc.HandleHTTP()

    http.DefaultServeMux = oldMux

    // Start listening for new friends
    go http.ListenAndServe(address, mux)
    fmt.Println("Started master server")
    return mr
}

func (mr *Master) GetAllRequesters() ([]RequesterData) {
    return mr.requesters
}

// ====== PRIVATE METHODS =========
