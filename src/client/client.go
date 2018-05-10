package client

const masterAddress = "localhost:3333"

type Client struct {
	me       int
	username string
	fr       *Friend
	req      *Requester
}

func initClient(username string, port int) *Client {
	client := Client{username: username}

	requester := initRequester(username, masterAddress)
	friend := initFriend(username, port)

	client.fr = friend
	client.req = requester

	return &client
}

func NewClient(username string, port int) *Client {
	return initClient(username, port)
}

func (cl *Client) StartJob(filename string) {
	cl.req.StartJob(filename)
}
