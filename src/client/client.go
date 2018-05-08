package client

const masterAddress = "hello_world"

type Client struct {
	me  int
	fr  *Friend
	req *Requester
}

func initClient() *Client {
	client := Client{}
	friend := initFriend(19997)
	requester := initRequester(19998)

	client.fr = friend
	client.req = requester

	return &client
}

func NewClient() *Client {
    return initClient()
}

func (cl *Client) requestJob() {

}
