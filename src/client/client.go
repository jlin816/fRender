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
	friend := initFriend(username, port, masterAddress)

	client.fr = friend
	client.req = requester

	return &client
}

func NewClient(username string, port int) *Client {
	cl := initClient(username, port)
	return cl
}

func NewBadClient(username string, port int) *Client {
	cl := initClient(username, port)
	cl.fr.Bad = true
	return cl
}

func (cl *Client) StartJob(filename string, numFrames int, numFriends int) {
	cl.req.StartJob(filename, numFrames, numFriends)
}
