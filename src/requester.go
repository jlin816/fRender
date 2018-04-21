package main

// const masterAddress = "hello_world"

type Requester struct {
	me int
}

func initRequester() (*Requester) {
	requester := Requester{}
	requester.registerWithMaster()
	go requester.listenOnSocket()

	return &requester
}

func (req *Requester) listenOnSocket() {
	// call receiveJob here somewhere??
}

func (req *Requester) registerWithMaster() {

}

func (req *Requester) startJob() {

}

func (req *Requester) getProgress() {

}

func (req *Requester) cancelJob() {

}
