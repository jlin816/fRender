package common

type StartJobArgs struct {
	NumFriends	int
}

type StartJobReply struct {
	Friends		[]string
}

type RegisterFriendArgs struct {
	Address		string
    Username    string
}

type RegisterFriendReply struct {
	Success		bool
}

type RegisterRequesterArgs struct {
	Address		string
    Username    string
}

type RegisterRequesterReply struct {
	Success		bool
}

type HeartbeatArgs struct {
    Username    string
    Available   bool
}

type HeartbeatReply struct {
}
