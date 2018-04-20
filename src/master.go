package main

import (
    "net"
    "encoding/json"
)

type Master struct {
    friends     map[Friend]bool
}

type Friend struct {
    id          int
    socket      net.Conn
}

type StartJobReply struct {
    Friends     `json:"friends"`
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
