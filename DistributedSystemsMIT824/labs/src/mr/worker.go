package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"net/rpc"
	"os"
	"sort"
	"time"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	for doNextTask(mapf, reducef) {
		// Loop until error or termination request from the server
		time.Sleep(time.Second) // sleep to be safe
	}

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()

}

// Return false to terminate the loop (no more tasks will be executed)
func doNextTask(mapf func(string, string) []KeyValue, reducef func(string, []string) string) bool {
	reqReply, ok := callRequestTask()
	if !ok {
		return false
	}

	switch reqReply.Type {
	case ReqTerm:
		fmt.Println("Terminated by the coordinator")
		return false
	case MapTask:
		return doMapTask(reqReply.MapTask, mapf)
	case ReduceTask:
		return doReduceTask(reqReply.ReduceTask, reducef)
	default:
		fmt.Println("Unknown request response type!")
		return false
	}
}

func callRequestTask() (*RequestTaskReply, bool) {
	args := &RequestTaskArgs{}
	reply := &RequestTaskReply{}
	ok := call("Coordinator.RequestTask", args, reply)
	return reply, ok
}

func doMapTask(taskInfo MapTaskReply, mapf func(string, string) []KeyValue) bool {
	taskId := taskInfo.TaskId
	filename := taskInfo.Filename
	numReduce := taskInfo.NumReduce

	// Open file for reading its contents for map reduce
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("cannot open %v\n", filename)
		return false
	}
	defer file.Close()

	// read the file
	content, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("cannot read %v\n", filename)
		return false
	}

	// run map function of map-reduce
	kva := mapf(filename, string(content))

	// write key value pairs to file system
	if committed := commitKva(taskId, numReduce, kva); !committed {
		return false
	}

	return signalCompletion(MapTask, taskId)
}

func doReduceTask(taskInfo ReduceTaskReply, reducef func(string, []string) string) bool {
	taskId := taskInfo.TaskId
	numMap := taskInfo.NumMap

	// collect all key value pairs from each file and then sort
	intermediate := []KeyValue{}
	for i := 0; i < numMap; i++ {
		kva, extracted := extractKva(i, taskId)
		if !extracted {
			return false
		}
		intermediate = append(intermediate, kva...)
	}

	// reduce and commit output
	if committed := commitOutput(taskId, intermediate, reducef); !committed {
		return false
	}

	return signalCompletion(ReduceTask, taskId)
}

func signalCompletion(taskType int, taskId int) bool {
	args := &TaskCompleteArgs{
		Type:   taskType,
		TaskId: taskId,
	}
	reply := &TaskCompleteReply{}
	return call("Coordinator.CompleteTask", args, reply)
}

// Creates and commits key value pairs to numReduce files of the name mr-<map taskId>-<reduce taskId> in json format
func commitKva(taskId int, numReduce int, kva []KeyValue) bool {
	// Create temp files from 0 - numReduce
	// Get json encoders
	encoders := []*json.Encoder{}
	files := []*os.File{}
	for i := 0; i < numReduce; i++ {
		// for ith reduce job
		oname := fmt.Sprintf("mr-%v-%v", taskId, i)
		ofile, oerr := ioutil.TempFile("/tmp", oname)
		if oerr != nil {
			fmt.Printf("Failed to open file with name %v\n", ofile.Name())
			return false
		}
		// Remove temporary file
		// removing the file is optional since /tmp files are removed periodically anyway
		defer os.Remove(ofile.Name())
		files = append(files, ofile)

		// store encoder for future writes
		encoders = append(encoders, json.NewEncoder(ofile))
	}

	// Append each kv to the appropriate file
	for _, kv := range kva {
		// Find the reducer to whom the key should go to
		reduceId := ihash(kv.Key) % numReduce
		// Encode the kv as json
		if encErr := encoders[reduceId].Encode(&kv); encErr != nil {
			fmt.Printf("Cannot encode kv %v\n", encErr)
			return false
		}
	}

	// Atomically commit temp files of reduce jobs [0, numReduce)
	for i := 0; i < numReduce; i++ {
		oname := fmt.Sprintf("mr-%v-%v", taskId, i)
		filename := files[i].Name()
		files[i].Close()
		oerr := os.Rename(filename, oname)
		if oerr != nil {
			fmt.Printf("Rename failed with err: %v\n", oerr)
			return false
		}
	}

	return true
}

func extractKva(mapId int, reduceId int) ([]KeyValue, bool) {
	// Read contents from mr-<mapId>-<reduceId>
	iname := fmt.Sprintf("mr-%v-%v", mapId, reduceId)
	ifile, ierr := os.Open(iname)
	if ierr != nil {
		fmt.Printf("Couldn't open file with err %v\n", ierr)
		return nil, false
	}
	decoder := json.NewDecoder(ifile)
	kva := []KeyValue{}
	for {
		var kv KeyValue
		if decErr := decoder.Decode(&kv); decErr != nil {
			if decErr == io.EOF {
				// Reached end of file
				return kva, true
			}

			// Decode error
			fmt.Printf("Failed to decode keyvalue pair with err: %v\n", decErr)
			return nil, false
		}
		kva = append(kva, kv)
	}
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

func commitOutput(taskId int, intermediate []KeyValue, reducef func(string, []string) string) bool {
	// Create tempfile for appending key value pairs
	oname := fmt.Sprintf("mr-out-%v", taskId)
	tfile, terr := ioutil.TempFile("/tmp", oname)
	if terr != nil {
		fmt.Printf("File open failed with err: %v\n", terr)
		return false
	}
	defer os.Remove(tfile.Name())

	// Sort by Key to group values of the same key
	sort.Sort(ByKey(intermediate))

	i := 0
	for i < len(intermediate) {
		// Reduce values of the same key
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}

		// Collect from i to j (have the same key)
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}

		// Reduce values of the same key
		output := reducef(intermediate[i].Key, values)

		// Append kv to file
		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(tfile, "%v %v\n", intermediate[i].Key, output)

		i = j
	}

	// Atomic rename
	filename := tfile.Name()
	tfile.Close()
	rerr := os.Rename(filename, oname)
	if rerr != nil {
		fmt.Printf("Rename failed with err: %v\n", rerr)
		return false
	}

	return true
}

//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		// TODO: dial only once and detect pipe closes later to catch genuine closes
		// log.Fatal("dialing:", err)
		return false
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
