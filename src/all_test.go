package main

import (
    "client"
    "master"
    "testing"
    "fmt"
)

func assert(t *testing.T, condition bool, message string) {
    if !condition {
        t.Error("Failed: ", message)
    }
    fmt.Println("Success: ", message)
}

func TestRegisterClient(t *testing.T) {
    mr := master.NewMaster()
    requesters := mr.GetAllRequesters()
    friends := mr.GetAllFriends()
    // Make sure master initially knows no requesters / friends.
    assert(t, len(requesters) == 0, "Master initially knows no requesters")
    assert(t, len(friends) == 0, "Master initially knows no friends")

    client.NewClient("client1")
    // Test that a client can register itself as a requester on the master.
    requesters = mr.GetAllRequesters()
    assert(t, len(requesters) == 1, "Master knows one requester")
//    assert(t, requesters[0].username == "client1", "Master has registered tthe requester")

    // Test that a client can register itself as a friend on the master.
    friends = mr.GetAllFriends()
    assert(t, len(friends) == 1, "Master knows one friend")
}

func TestStartJobSuccess(t *testing.T) {
    // Test that a requester can get back n friends if there are n friends available.
}

func TestStartJobRetry(t *testing.T) {
    // Test that a requester gets <n friends on first try but eventually gets n friends.
}

func TestFriendStatusUpdates(t *testing.T) {
    // Test that when a friend is busy, the master knows and doesn't assign it to a new job.
    // Test that when a friend goes down, the master finds out.
    // Test that when a friend becomes available again after {busy, down},
    //  master updates and assigns it to a new job.
}

func TestRequesterFriendCommunication(t *testing.T) {
    // Test that when a requester gets a friend address, it can communicate with it
    //  and friend eventually gets the frames.
}

func TestReceiveFrames(t *testing.T) {
    // Test that when a friend is assigned to a job, it eventually gets the frames.
}

func TestRenderFrames(t *testing.T) {
    // Full integration: test when a job is requested, the requester gets back rendered frames.
}
