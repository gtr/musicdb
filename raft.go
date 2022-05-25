package main

// The raft.go file closely follows the raft paper "In Search of an
// Understandable Consensus Algorithm" by Diego Ongaro and John Ousterhout.

import (
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// ================================ NODE STATE ================================

// NodeState represents one of 4 states for a node: (0) Follower, (1) Candidate
// (2) Leader, and (3) Dead.
type NodeState int

const (
	FOLLOWER  NodeState = 0
	CANDIDATE NodeState = 1
	LEADER    NodeState = 2
	DEAD      NodeState = 3
)

// ============================= CONSENSUS MODULE =============================

// ConsensusModule represents an instance of a node in the raft algorithm.
type ConsensusModule struct {
	// Persistent state on all nodes:
	id          int        // ID of the current node
	currentTerm int        // Latest term node has seen
	votedFor    int        // Candidate that recieve vote in current term
	log         []LogEntry // Log entries

	// Volatile state on all nodes:
	state       NodeState // The current state of the node
	commitIndex int       // Index of highest log entry known to be committed
	lastApplied int       // Index of highest log entry applied to state machine
	votes       int       // The number of votes a node has (used for elections)

	// Volatile state on leaders (reinitialized after election):
	nextIndex  []int // For each server, index of the next log entry to send to that server
	matchIndex []int // For each server, index of highest log entry known to be replicated on server

	// Election and peers
	leader  string              // Who the node thinks the leader is
	peerIds []int               // A list of all other node peers in the cluser
	peers   map[int]*rpc.Client // A list of all other node peers RPC clients

	// Concurrency and timing
	mu                 sync.Mutex           // A mutex to protect node data
	electionResetEvent time.Time            // Time of last election
	commitChannel      chan<- EntryToCommit // The channel that the node will pass committed log entries
}

// ======================= COMMUNICATION TO OTHER PEERS =======================

func (node *ConsensusModule) GetPeer(peer int) *rpc.Client {
	node.mu.Lock()
	defer node.mu.Unlock()

	return node.peers[peer]
}

func (node *ConsensusModule) SetPeer(peer int, client *rpc.Client) {
	node.mu.Lock()
	defer node.mu.Unlock()

	node.peers[peer] = client
}

/*
 * ConnectToPeer connects to a peer given its ID and network address.
 */
func (node *ConsensusModule) ConnectToPeer(peer int, addr net.Addr) error {
	if node.GetPeer(peer) == nil {
		client, err := rpc.Dial(addr.Network(), addr.String())
		if err != nil {
			return err
		}

		node.SetPeer(peer, client)
	}

	return nil
}

/*
 * DisconnectFromPeer disconnects from a peer given its ID.
 */
func (node *ConsensusModule) DisconnectFromPeer(peer int) error {
	if node.GetPeer(peer) != nil {
		err := node.peers[peer].Close()
		node.SetPeer(peer, nil)
		return err
	}
	return nil
}

/*
 * DoRPC performs an RPC to a peer.
 */
func (node *ConsensusModule) DoRPC(peer int, method string, args, reply interface{}) error {
	client := node.GetPeer(peer)

	if client == nil {
		return fmt.Errorf("Client is nil.")
	}

	return client.Call(method, args, reply)
}

// ================================= LOG INFO =================================

/*
 * lastLogTerm returns the last log term of the node.
 */
func (node *ConsensusModule) lastLogTerm() int {
	if len(node.log) == 0 {
		return -1
	}
	return node.log[len(node.log)-1].Term
}

/*
 * lastLogIndex returns the last log index of the node.
 */
func (node *ConsensusModule) lastLogIndex() int {
	return len(node.log) - 1
}

/*
 * UpdatePeerIndicies updates nextIndex and matchIndex for all the peers.
 */
func (node *ConsensusModule) UpdatePeerIndicies() {
	for _, peer := range node.peerIds {
		node.nextIndex[peer] = len(node.log)
		node.matchIndex[peer] = -1
	}
}

// ============================= ELECTION PROCESS =============================

/*
 * getElectionTimeout returns a random election timeout between 100-200ms.
 */
func (node *ConsensusModule) getElectionTimeout() time.Duration {
	return time.Millisecond * time.Duration(100+rand.Intn(100))
}

/*
 * StartElectionTimer starts a countdown timer in which the node will attempt
 * to become the leader. Each invocation of the function generates a random
 * timeout so that it is unlikely that two nodes attempt to become the leader.
 */
func (node *ConsensusModule) StartElectionTimer() {
	duration := node.getElectionTimeout()
	node.mu.Lock()
	term := node.currentTerm
	node.mu.Unlock()

	// Make a new ticker for 10 milliseconds.
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		// Blocks until we receive a message in this ticker channel.
		<-ticker.C
		node.mu.Lock()
		defer node.mu.Unlock()

		// In followers, this loop should run forever. There are two ways in
		// which the loop is broken...

		// (1) if the current term is not the term we started with (new leader)
		if node.currentTerm != term {
			return
		}

		// (2) if we haven't received any heartbeats from the leader within our
		// timeout duration, in which case we start a new election process
		last := time.Since(node.electionResetEvent)
		if last >= duration {
			node.StartElectionProcess()
			return
		}
	}
}

/*
 * prepareRequestVoteForPeer concurrently prepares and sends RequestVote RPCs
 * for a peer.
 */
func (node *ConsensusModule) prepareRequestVoteForPeer(peer, currTerm int) {
	node.mu.Lock()
	lastLogIndex := node.lastLogIndex()
	lastLogTerm := node.lastLogTerm()
	node.mu.Unlock()

	// Create a RequestVoteArgs message.
	requestVoteArgs := RequestVoteArgs{
		term:         currTerm,
		candidateID:  node.id,
		lastLogIndex: lastLogIndex,
		lastLogTerm:  lastLogTerm,
	}

	var requestVoteReply RequestVoteReply

	err := node.DoRPC(peer, "RequestVote", requestVoteArgs, &requestVoteReply)
	if err == nil {
		node.mu.Lock()
		defer node.mu.Unlock()

		// If the reply's term is greater tham ours, stop being the candidate
		// and become a follower again.
		if requestVoteReply.term > currTerm {
			node.BecomeFollower(requestVoteReply.term)
		}

		// Continuing on from the last if statement, if we are no longer a
		// candidate, just return.
		if node.state != CANDIDATE {
			return
		}

		// If the reply's term matches our term and they voted for us, increase
		// the vote count and check if we have a quorum.
		if requestVoteReply.term == currTerm && requestVoteReply.voteGranted {
			node.votes += 1
			if (node.votes * 2) > len(node.peers) {
				node.BecomeLeader()
				return
			}
		}
	}

}

/*
 * StartElectionProcess starts a new election process for the node.
 */
func (node *ConsensusModule) StartElectionProcess() {
	node.mu.Lock()
	defer node.mu.Unlock()

	// 1. Change the state of the current node to become a candidate.
	node.state = CANDIDATE

	// 2. Vote for yourself :)
	node.votedFor = node.id
	node.votes = 1

	// 3. Note the current term and the time.
	node.currentTerm += 1
	term := node.currentTerm
	node.electionResetEvent = time.Now()

	// 4. For each peer, send them for a request vote message.
	for _, peer := range node.peerIds {
		go node.prepareRequestVoteForPeer(peer, term)
	}

	go node.StartElectionTimer()
}

// ============================ LEADER OPERATIONS =============================

/*
 * true if the number of votes constitutes a quorum (majority)
 */
func (node *ConsensusModule) hasQuorum(votes int) bool {
	return (votes*2 > len(node.peerIds)+1)
}

func (node *ConsensusModule) checkIfStillLeader() bool {
	node.mu.Lock()
	defer node.mu.Unlock()

	return node.state == LEADER
}

/*
 * LeaderLoop will run as long as the node is the leader.
 */
func (node *ConsensusModule) LeaderLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		node.SendHeartbeats()
		<-ticker.C

		if !node.checkIfStillLeader() {
			return
		}
	}

}

func (node *ConsensusModule) prepareAppendEntriesForPeer(peer, term int) {
	node.mu.Lock()
	next := node.nextIndex[peer]
	prev := next - 1
	prevLogTerm := -1
	if prev >= 0 {
		prevLogTerm = node.log[prev].Term
	}

	entries := node.log[next:]

	appendEntriesArgs := AppendEntriesArgs{
		term:         term,
		leaderId:     node.id,
		prevLogIndex: prev,
		prevLogTerm:  prevLogTerm,
		entries:      entries,
		leaderCommit: node.commitIndex,
	}

	node.mu.Unlock()

	var reply AppendEntriesReply
	err := node.DoRPC(peer, "AppendEntries", appendEntriesArgs, &reply)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		node.mu.Lock()
		defer node.mu.Unlock()

		// If the reply's term is greater than our saved term, that
		// means that the leader is out of sync and is thus no longer
		// the leader.
		if reply.term > term {
			node.BecomeFollower(reply.term)
			return
		}

		if node.state == LEADER && term == reply.term {
			// If the AppendEntries request was not successful, return the
			// nextIndex pointer to next - 1.
			if !reply.success {
				node.nextIndex[peer] = next - 1
				return
			}
			node.updateEntries(peer, next, entries)
		}
	}
}

/*
 * SendHeartbeats sends one heartbeat per peer concurrently.
 */
func (node *ConsensusModule) SendHeartbeats() {
	node.mu.Lock()
	currTerm := node.currentTerm
	node.mu.Unlock()

	// Concurrently prepare to send AppendEntries messages to our peers.
	for _, peer := range node.peerIds {
		go node.prepareAppendEntriesForPeer(peer, currTerm)
	}

}

// updateEntries updated our node's commit index to match that of our peers'
func (node *ConsensusModule) updateEntries(peer, next int, entries []LogEntry) {
	node.nextIndex[peer] = next + len(entries)
	node.matchIndex[peer] = node.nextIndex[peer] - 1

	commitIndex := node.commitIndex + 1

	for commitIndex < len(node.log) {
		commitIndex++
		if node.log[commitIndex].Term == node.currentTerm {
			count := 1

			// Go through all our peer's indicies to check which are greater than
			for _, currPeer := range node.peerIds {
				if node.matchIndex[currPeer] >= commitIndex {
					count += 1
				}
			}
			// If we have a quorum, then we can update the commit index of our
			// log :).
			if node.hasQuorum(count) {
				node.commitIndex = commitIndex
			}
		}
	}
}

// ============================ NODE STATE CHANGES ============================

/*
 * BecomeLeader changes a node to the LEADER state and then sends heartbeats to
 * other peers to establish its authority and prevent new elections.
 */
func (node *ConsensusModule) BecomeLeader() {
	// Change the node state to LEADER
	node.state = LEADER

	// Update the indicies for all peers.
	node.UpdatePeerIndicies()

	// Run the leader loop, concurrently.
	go node.LeaderLoop()
}

/*
 * BecomeFollower changes a node to the FOLLOWER state.
 */
func (node *ConsensusModule) BecomeFollower(term int) {
	// Reset fields back to follower defaults.
	node.state = FOLLOWER
	node.votedFor = -1
	node.currentTerm = term
	node.electionResetEvent = time.Now()

	// Start the periodic election timer.
	go node.StartElectionTimer()
}
