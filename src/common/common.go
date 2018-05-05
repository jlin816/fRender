package common

type StartJobArgs struct {
	NumFriends	int
}

type StartJobReply struct {
	Friends		[]string
}

type RegisterFriendArgs struct {
	Address		string
}

type RegisterFriendReply struct {
	Success		bool
}

type RegisterRequesterArgs struct {
	Address		string
}

type RegisterRequesterReply struct {
	Success		bool
}
