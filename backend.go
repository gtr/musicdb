package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
)

// ============================== BACKEND SERVER ==============================

// BackendServer represents a backend TCP BackendServer.
type BackendServer struct {
	// Backend fields
	Host string   // The hostname of the backend server
	Port string   // The port number of the backend server
	DB   *AlbumDB // A pointer to the in-memory album database

	consensus *ConsensusModule // The Consesus module
}

/*
 * NewBackendServer initializes a new backend BackendServer.
 */
func NewBackendServer(host, port string, endpoints []string) *BackendServer {
	return &BackendServer{
		Host: host,
		Port: port,
		DB:   NewAlbumDB(),
	}
}

/*
 * Start starts running the backend server; continously listens for incoming
 * requests from the frontend server(s).
 */
func (srv *BackendServer) Start() {
	log.Println("[BackendServer] Starting backend BackendServer on " + srv.Host + srv.Port)

	listener, err := net.Listen("tcp4", srv.Port)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Continously listen for requests.
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Start", err)
			return
		}
		go srv.HandleClientConn(conn)
	}
}

/*
 * HandleClientConn handles an incoming client connection; reads message.
 */
func (srv *BackendServer) HandleClientConn(conn net.Conn) {
	log.Println("[BackendServer] Handling " + conn.RemoteAddr().String())

	for {
		msg := srv.ReadClientMessage(conn)
		srv.HandleClientRequest(conn, msg)
	}
}

// ============================ READ/WRITE MESSAGES ===========================

/*
 * ReadClientMessage reads a client message from a TCP connection.
 */
func (srv *BackendServer) ReadClientMessage(conn net.Conn) *DataMessage {
	log.Print("[BackendServer] Reading message")

	for {
		msg := &DataMessage{}

		decoder := gob.NewDecoder(conn)
		if err := decoder.Decode(msg); err != nil {
			panic(err)
		}
		log.Println(msg)
		return msg
	}
}

/*
 * WriteClientMessage writes a message to a client over a TCP connection.
 */
func (srv *BackendServer) WriteClientMessage(conn net.Conn, msg *DataMessage) {
	log.Println("[BackendServer] Sending message", msg)

	encoder := gob.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		panic(err)
	}
}

// ======================= COMMUNICATION TO OTHER NODES =======================

// Returns the address of the current node which is just the hostname and port.
func (node *BackendServer) GetAddress() string {
	return node.Host + node.Port
}

// ============================== CLIENT REQUESTS =============================

/*
 * HandleClientRequest handles the client's request and performs the
 * appropriate data store operations for the given request.
 */
func (srv *BackendServer) HandleClientRequest(conn net.Conn, request *DataMessage) {
	switch request.Method {
	case "GetAllAlbums":
		srv.handleGetAllAlbums(conn)
	case "GetAlbum":
		srv.handleGetAlbum(conn, request)
	case "AddAlbum":
		srv.handleAddAlbum(conn, request)
	case "EditAlbum":
		srv.handleEditAlbum(conn, request)
	case "DeleteAlbum":
		srv.handleDeleteAlbum(conn, request)
	default:
		log.Println("[BackendServer] Invalid method", request.Method)
		os.Exit(1)
	}
}

/*
 * handleGetAllAlbums gets all albums from the in-memory databse.
 */
func (srv *BackendServer) handleGetAllAlbums(conn net.Conn) {
	response := &DataMessage{
		Method:     "GetAllAlbums",
		AlbumArray: srv.DB.GetAllAlbums(),
		Status:     true,
	}

	srv.WriteClientMessage(conn, response)
}

/*
 * handleGetAlbum gets an album from the in-memory database.
 */
func (srv *BackendServer) handleGetAlbum(conn net.Conn, request *DataMessage) {
	album, err := srv.DB.GetAlbum(request.Index)
	if err != nil {
		srv.WriteClientMessage(conn, &DataMessage{
			Status: false,
		})
	}

	response := &DataMessage{
		Method:     "GetAlbum",
		AlbumArray: []*Album{album},
		Status:     true,
	}

	srv.WriteClientMessage(conn, response)
}

/*
 * handleAddAlbum adds an album to the in-memory database.
 */
func (srv *BackendServer) handleAddAlbum(conn net.Conn, request *DataMessage) {
	album := request.AlbumArray[0]
	srv.DB.AddAlbum(album.Title, album.Artist, album.URL, album.Year)

	response := &DataMessage{
		Status: true,
	}

	srv.WriteClientMessage(conn, response)
}

/*
 * handleEditAlbum edits an album in the in-memory database.
 */
func (srv *BackendServer) handleEditAlbum(conn net.Conn, request *DataMessage) {
	log.Println("[BackendServer] handleEditAlbum", request)
	album := request.AlbumArray[0]
	err := srv.DB.EditAlbum(request.Index, album.Title, album.Artist, album.URL, album.Year)

	if err != nil {
		log.Println("[BackendServer]", err)
	}
	response := &DataMessage{
		Status: err == nil,
	}

	srv.WriteClientMessage(conn, response)
}

/*
 * handleDeleteAlbum deletes an album from the in-memory database.
 */
func (srv *BackendServer) handleDeleteAlbum(conn net.Conn, request *DataMessage) {
	fmt.Println("handleDeleteAlbum " + request.Index)
	err := srv.DB.RemoveAlbum(request.Index)

	response := &DataMessage{
		Status: err == nil,
	}

	srv.WriteClientMessage(conn, response)
}

// ========================= MAIN & PARSING FUNCTIONS =========================

func ParseBackendendCommandLineArgs() (string, []string) {
	args := os.Args
	endPoints := []string{}
	httpPort := ":8090"
	i := 1
	for i < len(args) {
		if args[i] == "--listen" {
			httpPort = ParseListenFlag(args, i)
			i += 2
		} else if args[i] == "--backend" {
			endPoints = ParseBackendEndpointsFlag(args, i)
			i += 2
		} else {
			fmt.Println("Incorrect usage")
			os.Exit(1)
		}
	}
	return httpPort, endPoints
}

func main() {

	httpPort, endpoints := ParseBackendendCommandLineArgs()

	srv := NewBackendServer("localhost", httpPort, endpoints)
	srv.Start()
}
