package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	// Your definitions here.
	endCond *sync.Cond
	done    bool

	taskCh     chan RequestTaskReply
	completeCh chan *TaskCompleteArgs

	files   []string
	nReduce int

	l net.Listener
}

// Your code here -- RPC handlers for the worker to call.

func (c *Coordinator) RequestTask(args *RequestTaskArgs, reply *RequestTaskReply) error {
	stopCh := c.getStopSignal()

	select {
	case task := <-c.taskCh:
		// Pull task from the queue
		*reply = task
	case <-stopCh:
		// Stop the worker
		*reply = RequestTaskReply{
			Type: ReqTerm,
		}
	}

	return nil
}

func (c *Coordinator) CompleteTask(args *TaskCompleteArgs, reply *TaskCompleteReply) error {
	stopCh := c.getStopSignal()

	select {
	// signal task completion
	case c.completeCh <- args:
	// server stopped
	case <-stopCh:
	}

	*reply = TaskCompleteReply{}

	return nil
}

func (c *Coordinator) getStopSignal() chan struct{} {
	// Check if the server stopped/stopping
	stopCh := make(chan struct{})
	go func() {
		c.endCond.L.Lock()
		defer c.endCond.L.Unlock()
		for !c.done {
			c.endCond.Wait()
		}

		// Signal that all tasks are complete
		stopCh <- struct{}{}
	}()

	return stopCh
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	// Close the listener on termination
	c.l = l
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	// ret := false

	// Your code here.

	// Generate and execute map tasks
	mapTasks := []RequestTaskReply{}
	for i, file := range c.files {
		mapTasks = append(mapTasks, RequestTaskReply{
			Type: MapTask,
			MapTask: MapTaskReply{
				TaskId:    i,
				Filename:  file,
				NumReduce: c.nReduce,
			},
		})
	}

	c.execTasks(MapTask, mapTasks)

	// Generate and execute reduce tasks
	reduceTasks := []RequestTaskReply{}
	nMap := len(mapTasks)
	for i := 0; i < c.nReduce; i++ {
		reduceTasks = append(reduceTasks, RequestTaskReply{
			Type: ReduceTask,
			ReduceTask: ReduceTaskReply{
				TaskId: i,
				NumMap: nMap,
			},
		})
	}

	c.execTasks(ReduceTask, reduceTasks)

	c.l.Close() // not strictly required

	// Notify every handler that all required work is complete
	c.endCond.L.Lock()
	c.done = true
	c.endCond.Broadcast()
	c.endCond.L.Unlock()

	// return ret
	return true
}

const (
	TaskTimeout = 10 * time.Second
)

func (c *Coordinator) execTasks(taskType int, tasks []RequestTaskReply) {
	// All map tasks are pending initially
	pendingTaskIdQ := []int{}
	nTasks := len(tasks)
	for i := 0; i < nTasks; i++ {
		pendingTaskIdQ = append(pendingTaskIdQ, i)
	}

	// Wait for all map tasks to complete
	completedTasks := map[int]struct{}{}

	markComplete := func(args *TaskCompleteArgs) {
		if args.Type == taskType {
			completedTasks[args.TaskId] = struct{}{}
		}
	}

	reassignTask := func(backupTaskId int) {
		if _, exists := completedTasks[backupTaskId]; !exists {
			// Consider task for re assignment (backup and failures)
			pendingTaskIdQ = append(pendingTaskIdQ, backupTaskId)
		}
	}

	// Wait until all tasks are complete
	backupTaskCh := make(chan int)
	for len(completedTasks) < nTasks {
		if len(pendingTaskIdQ) == 0 {
			// No pending tasks remaining
			select {
			case completedTaskInfo := <-c.completeCh:
				markComplete(completedTaskInfo)
			case backupTaskId := <-backupTaskCh:
				reassignTask(backupTaskId)
			}
		} else {
			select {
			case c.taskCh <- tasks[pendingTaskIdQ[0]]:
				// Dequeue assigned task
				queuedTaskId := pendingTaskIdQ[0]
				pendingTaskIdQ = pendingTaskIdQ[1:]
				// TODO: Cancel goroutines after all tasks are done
				time.AfterFunc(TaskTimeout, func() {
					// Wait for TaskTimeout to reassign the task to another worker
					backupTaskCh <- queuedTaskId
				})
			case completedTaskInfo := <-c.completeCh:
				markComplete(completedTaskInfo)
			case backupTaskId := <-backupTaskCh:
				reassignTask(backupTaskId)
			}
		}
	}
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	c.endCond = sync.NewCond(&sync.Mutex{})
	c.done = false
	c.taskCh = make(chan RequestTaskReply)
	c.completeCh = make(chan *TaskCompleteArgs)
	c.files = files
	c.nReduce = nReduce

	c.server()
	return &c
}
