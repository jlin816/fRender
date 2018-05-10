package main

import (
	"client"
	"fmt"
	"io/ioutil"
	"master"
	"os"
	"os/exec"
	"testing"
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

func TestRegisterClients(t *testing.T) {
	mr := master.NewMaster()
	client.NewClient("client1", 19997)
	client.NewClient("client2", 19995)
	client.NewClient("client3", 19993)
	// Test that a client can register itself as a requester on the master.
	requesters := mr.GetAllRequesters()
	assert(t, len(requesters) == 3, "Master knows three requesters")
	//    assert(t, requesters[0].username == "client1", "Master has registered tthe requester")

	// Test that a client can register itself as a friend on the master.
	// fmt.Println("TEST REGISTER CLIENT")
}

func TestStartJobSuccessOneFriend(t *testing.T) {
	startup("TestStartJobSuccessOneFriend")

	_ = master.NewMaster()
	cl := client.NewClient("client1", 19997)
	cl.StartJob("file.blend", 15)

	// timer1 := time.NewTimer(10 * time.Second)
	// <-timer1.C
	// Test that a requester can get back n friends if there are n friends available.
}

func TestStartJobSuccessManyFriends(t *testing.T) {
	startup("TestStartJobSuccessManyFriends")
	numFrames := 6

	_ = master.NewMaster()
	cl1 := client.NewClient("client1", 19997)
	_ = client.NewClient("client2", 19995)
	_ = client.NewClient("client3", 19993)
	cl1.StartJob("file.blend", numFrames)

	files, _ := ioutil.ReadDir("files/client1_requester/file.blend_frames")
	assert(t, len(files) == (numFrames+1), "Rendered all the frames!")

	// timer1 := time.NewTimer(10 * time.Second)
	// <-timer1.C
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

func TestCleanup(t *testing.T) {
	// handy function to cleanup files before pushing code! :)
	os.RemoveAll("files/")
	os.Mkdir("files/", os.ModePerm)
}

func startup(test string) {
	os.RemoveAll("files/")
	os.Mkdir("files/", os.ModePerm)

	cpCmd := exec.Command("cp", "-a", fmt.Sprintf("testFiles/%v/.", test), "files/")
	_, err := cpCmd.Output()
	if err != nil {
		panic(err)
	}
}
