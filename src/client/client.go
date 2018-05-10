package client

const masterAddress = "localhost:3333"

type Client struct {
	me  int
    username string
	fr  *Friend
	req *Requester
}

func initClient(username string) *Client {
    client := Client{username: username}

	requester := initRequester(username, 19998, masterAddress)
	friend := initFriend(username, 19997, masterAddress)

	client.fr = friend
	client.req = requester

	return &client
}

func NewClient(username string) *Client {
    return initClient(username)
}

func (cl *Client) requestJob() {

}
