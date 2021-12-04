First, you may already heard about Finite State Machine (FSM) in Raft. It is the place where you can process the data in the local node.

First, you may notice (and must aware) that all operation in Raft Consensus is must be through the Leader. So, we need to ensure that all request must be sent to Leader node. Leader node then will send the command to all the follower, and wait to majority the server process the command. How many numbers of “majority” is symbolized by N number called “quorum”. Each follower then will do these:

    1. After receiving the command operation, data will be save in Write Ahead Log mode as a log entries.
    2. After successfully writing log entries, the data will be sent to FSM where you must process your data in a deterministic way.
    3. After you successfully process the data in FSM, it will return the data. Then the leader will be notified that this node already 
      success. Once the Leader thinks that there is enough follower telling that they already successfully process the data, then the
      Leader will tell the client that data already “distributed” to N quorum.
