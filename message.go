package main

// DataMessage represents a data message (relating to the data store) sent to
// the backend server over a TCP connection containing the method being called
// with optional index and an optional albumArray holding the album(s)
// requested
type DataMessage struct {
	Method     string   // The method being called
	Index      string   // The index of the album in the in-memory database
	AlbumArray []*Album // The album(s)
	Status     bool     // Boolean to determine if the request was successful
}

// NodeMessage represents a raft message (relating to the communication between
// client and nodes in the cluser).
type NodeMessage struct {
	Method string // The method being called
	ID     string // The ID of the node
	Term   int    // The current term
}

// ============================= REQUEST VOTE RPC =============================

// RequestVoteArgs represents the arguments passed to the RequestVote RPC. It's
// invoked by candidates to gather votes.
type RequestVoteArgs struct {
	term         int // Candidate's term
	candidateID  int // Candidate requesting vote
	lastLogIndex int // Index of candidate's last log entry
	lastLogTerm  int // Term of candidate's last log entry
}

// RequestVoteReply represents the reply to the RequestVote RPC.
type RequestVoteReply struct {
	term        int  // currentTerm, for the candidate to update itself
	voteGranted bool // True means the candidate received a vote
}

// ============================ APPEND ENTRIES RPC ============================

// AppendEntriesArgs represents the arguments to the AppendEntries RPC. It's
// invoked by the leader t replicate log entries; also used as a heartbeat.
type AppendEntriesArgs struct {
	term         int        // The leader's term
	leaderId     int        // So the follower can redirect clients
	prevLogIndex int        // Index of log entry immediately preceding new ones
	prevLogTerm  int        // Term of prevLogIndex entry
	entries      []LogEntry // Log entries to store (empty for heartbeat)
	leaderCommit int        // Leader's commitIndex
}

// AppendEntriesReply represents the reply to the AppendEntries RPC.
type AppendEntriesReply struct {
	term    int  // currentTerm, for the leader to update itself
	success bool // True if follower contained entry matching prevLogIndex and preLogTerm
}
