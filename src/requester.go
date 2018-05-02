package main

//const masterAddress = "hello_world"

type Requester struct {
	me      int
	friends []FriendData
}

func initRequester() *Requester {
	Requester := Requester{}
	Requester.registerWithMaster()
	go Requester.listenOnSocket()

	return &Requester //help
}

func (req *Requester) listenOnSocket() {
	// call receiveJob here somewhere??
}

func (req *Requester) registerWithMaster() {

}

func (req *Requester) connectToFriends() {

}

func (req *Requester) startJob() {

}

func (req *Requester) getProgress() {

}

func (req *Requester) cancelJob() {

}
