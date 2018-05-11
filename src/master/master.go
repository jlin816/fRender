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
    username    string
	address		net.Addr
	lastActive	time.Time
	available	bool // alive and not busy
	lastJob int // only increments
	points  int
}

type RequesterData struct {
    username    string
	  points			int
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
        username: args.Username,
		address: addr,
		available: true,
		lastActive: time.Now(),
		lastJob: 0,
		points: 0,
	}
    mr.friends = append(mr.friends, newFriend)
    fmt.Printf("Connected friend %s!\n", newFriend.username)
	reply.Success = true
	return nil
}

func (mr *Master) RegisterRequester(args RegisterRequesterArgs, reply *RegisterRequesterReply) error {
	mr.mu.Lock()
	defer mr.mu.Unlock()

    newRequester := RequesterData {
        username: args.Username,
				points: 0,
    }
    mr.requesters = append(mr.requesters, newRequester)
    fmt.Printf("Connected new requester %s!\n", newRequester.username)
    reply.Success = true
	return nil
}

func (mr *Master) StartJob(args StartJobArgs, reply *StartJobReply) error {
	fmt.Println("StartJob called")
	mr.mu.Lock()
	defer mr.mu.Unlock()

	friendCount := 0
	assignedFriends := make([]string, args.NumFriends)
	for _, req := range mr.requesters { // spend points to start job
		if req.username == args.Username {
			req.points -= 1
			break
		}
	}

	for _, friend := range mr.friends {
		if !friend.available || (time.Since(friend.lastActive) > friendTimeout) {
			continue
		}

		assignedFriends[friendCount] = friend.address.String()
		friend.available = false // mark this friend as unavailable
		friend.lastJob++
		friendCount++

		if friendCount == args.NumFriends {
			reply.Friends = assignedFriends
			return nil
		}
	}
	return errors.New("Not enough active friends")
}

func (mr *Master) Heartbeat(args HeartbeatArgs, reply *HeartbeatReply) error { mr.mu.Lock()
    defer mr.mu.Unlock()

    for _, friend := range mr.friends {
        if friend.username == args.Username {
            friend.lastActive = time.Now()
						if args.LastJobCompleted == friend.lastJob && args.Available {
							friend.available = true
							friend.points += 1    // this is when a friend has successfully finished a job
						}
            return nil
        }
    }
    return errors.New("Friend not found??")
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

func (mr *Master) GetAllFriends() ([]FriendData) {
    return mr.friends
}

// ====== PRIVATE METHODS =========
