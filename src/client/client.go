package main

const masterAddress = "hello_world"

type Client struct {
	me  int
	fr  *Friend
	req *Requester
}

func initClient() *Client {
	client := Client{}
	friend := initFriend()
	requester := initRequester()

	client.fr = friend
	client.req = requester

	return &client
}

func (cl *Client) requestJob() {

}
