package client

const masterAddress = "locahost:3333"

type Client struct {
	me  int
    username string
	fr  *Friend
	req *Requester
}

func initClient(username string) *Client {
    client := Client{username: username}
	friend := initFriend(username, 19997)
	requester := initRequester(username, 19998)

	client.fr = friend
	client.req = requester

	return &client
}

func NewClient(username string) *Client {
    return initClient(username)
}

func (cl *Client) requestJob() {

}
