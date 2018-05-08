package main

import (
    "client"
    "master"
    "testing"
)

func TestRegisterRequester(t *testing.T) {
    // Test that a requester can register itself on the master.
    client.NewClient()
    master.NewMaster()
}

func TestRegisterFriend(t *testing.T) {
    // Test that a friend can register itself on the master.
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
