# Notes

**Source:** http://nil.csail.mit.edu/6.824/2021/schedule.html

## Map reduce

### Paper

**Source:** http://nil.csail.mit.edu/6.824/2021/papers/mapreduce.pdf

* Introduces a functional style of map+reduce which is highly parellelizable
* Run-time system -> partitions input data, scheudle program execution across machines, handle machine failures
* Output equivalent to the sequential execution of *deterministic programs*
* In general, the output of each reduce task is equivalent to some sequential execution of the map task
* Backup tasks to handle **stragglers**
* Custom partition key functions. Ex: hostname of URL, instead of the URL itself

### Questions

* Tradeoffs of having a streaming file connection instead of the reduce job requesting only after map task is complete
* What are some alternate mechanisms to abstract handling fault tolerance and distributed nature? Tradeoffs
* Determine that the map reduce job failed (list failure senarios exhaustively)
* How would map-reduce be implemented on a NUMA architecture as opposed to a cluster of commodity machines?
* Can multiple implementations of map reduce be tier-ed?
* Why does master ping worker and not worker sends heartebats to master?
* Explain a secenario where R1 and R2 receive files produces by different executions of M
* How do the masters datastructures change when the backup tasks are introduced to avoid stragglers
* Can we use protobuf format to store key/value pairs instead of text files?
* Are reduce jobs scheduled close to all the intermediate files?

### Lecture

* Tradeoff between fault tolerance, consistency, performance
* Fault tolerance: availability through replication, recoverability through logging/txns
* Consistency: does get return the last put
* Performance: throughput and [tail] latency (harder to achieve)

## Google File System

### Paper

* Component failures are the norm
* Huge files
* Append heavy
* Snapshot and record append operations
* Usecases: multiway merge results and producer-consumer queue
* Files divided into chunks and each identified by 64-bit handle
* Default 3 replicas
* Record append has at least once semantics
* Chunks have leases
* Chunks may replicate record appends inconsistently. However, there is at least one consistent atomic append.
* Chunk placement: affects reliability and availability, maximize bandwith usage
* Chunk information dissemination also helps in garbage collection of headless chunks

### Questions

* Why is there no journaling (transaction log) step in GFS?
* How does the design cater to high throughput and not low latency?
* Do we ask data from the chunkserver that has the most number of chunks required?
* Design distributed metadata communication (potentially across clients)
* Are chunks cached by the kernel for use across multiple clients on the same machine/OS?
* Pros and cons of a longer heartbeat interval to the master process
* Why is caching chunk locations useful if the reads and writes are sequential (do they repeat)?
* Is the lease information cached at the client-side?
* Why doesnt the primary wait for all secondary writes to be acknowledged before writing its own?
* Tradeoff between master instructing copy on write rather than the chunk servers deciding for themselves
* Can any locks be omitted if the master process serializes all requests or is event-driven?
* Is proximity to client important for replica placement?

### Lecture

* Distributed storage systems are important since you can design apps as stateless and backed by fault-tolerant storage
* Challenges: high performance

## Fault-tolerant virtual machines

### Paper

* Deterministic replay
* Split-brain problem
* Primary - logging mode. Backup - replay mode

### Questions

* How are faults replicated? Are they ever non-deterministic? Ex: Page faults or spurious interrupts
* How is split-brain problem solved if the primary doesn't know that the secondary wants to go live?
* Are acknowledgments batched? Isnt there are a latency tradeoff for batching acknowledgments?
* What guest OSes are supported? Is it only linux or can any OS run?
* Is the storage system distributed?
* Why do we need to send the non-deterministic instruction? Isnt the next non-deterministic instruction determined?
* Can we even modify the next instruction at which the interrupt needs to be triggered for trapping?

## Raft

### Paper

* Safety: correct result in the presence of faults and duplication
* Available
* Independence on clocks for consensus
* Properties
  * Election safety - only one leader elected
  * Leader append only - leader logs are never overwritten or missing
  * Log matching - (index, term) uniquely determines state operation across logs
  * Leader completeness - committed entries always present in future leaders
* Raft protocol
  * Leader election - RequestVote RPC - randomized timeouts
  * Log replication - AppendRPCs - can go backwards when a new leader is selected
  * Election restriction - up-to-date committed entries for candidate to leader

### Questions

* What are the adjustments and corner cases when servers are ranked for leader election?
* Leader changes and the follower is not updated on the append entries from the previous leader yet
* Why not start nextIndex at committed index? Does a new leader even know what entries are committed?
* Does up-to-date only include committed entries?
* Why does the T1 > T2 argument for log completeness not work when T1 == T2?
* Why is lastAppliedIndex volatile?
* Why is matchIndex different from nextIndex?
