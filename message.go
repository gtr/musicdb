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

// RaftMessage represents a raft message (relating to the communication between
// client and nodes in the cluser).
type RaftMessage struct {
}
