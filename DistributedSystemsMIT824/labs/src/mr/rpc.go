package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.

// Request for a new task
type RequestTaskArgs struct{}

type RequestTaskReply struct {
	Type       int
	MapTask    MapTaskReply
	ReduceTask ReduceTaskReply
}

const (
	ReqTerm    = iota
	MapTask    = iota
	ReduceTask = iota
)

type MapTaskReply struct {
	TaskId    int
	Filename  string
	NumReduce int
}

type ReduceTaskReply struct {
	TaskId int
	NumMap int
}

// Signal task completion
type TaskCompleteArgs struct {
	Type   int
	TaskId int
}

type TaskCompleteReply struct{}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
