package main

import (
	"net"
)

type Master struct {
	friends map[Friend]bool
}

type FriendData struct {
	id     int
	socket net.Conn
}

type StartJobReply struct {
	Friends []FriendData `json:"friends"`
}

func initMaster() (mr *Master) {
}

func (mr *Master) RegisterNewFriend() {
	// Register a new user for the first time.
}

func (mr *Master) StartJob(numFriends int) StartJobReply {
	// Returns active friends allocated to this job
}

func (mr *Master) heartbeat() {
}
