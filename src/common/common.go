package common

type StartJobArgs struct {
  Username    string
	NumFriends	int
}

type StartJobReply struct {
	Friends		[]string
}

type PointsArgs struct {
  PointDist map[string]int
}

type PointsReply struct {
	Success bool
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
    LastJobCompleted int
}

type HeartbeatReply struct {
}
