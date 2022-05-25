package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"

	"github.com/kataras/iris/v12"
)

// ============================== FRONTEND SERVER ==============================

// FrontendServer represents the frontend server
type FrontendServer struct {
	HTTPPort  string       // Port to listen to HTTP requests
	Endpoints []string     // Endpoints to the backend servers
	Conn      *net.TCPConn // TCP connection to backend server (leader)
}

/*
 * NewFrontendServer initializes a new frontend server.
 */
func NewFrontendServer(httpPort string, endpoints []string) *FrontendServer {
	return &FrontendServer{
		HTTPPort:  httpPort,
		Endpoints: endpoints,
	}
}

/*
 * Start starts running the frontend server; initializes an iris app, and
 * handles GET and POST requests from client(s).
 */
func (srv *FrontendServer) Start() {

	// Initialize an Iris app.
	app := iris.Default()

	// Connect to the backend server via TCP
	srv.ConnectToBackend(srv.PickRandom())

	// Register a folder for HTML templates.
	app.RegisterView(iris.HTML("./views", ".html"))

	// Show the homepage of the app.
	app.Get("/", srv.ShowHomePage)

	// Handle the add album route.
	app.Post("/add", srv.HandleAddAlbumRoute)

	// Show the add album page.
	app.Get("/add", srv.ShowAddPage)

	// Show the album page for a particular album.
	app.Get("/album/{id:uint64}", srv.ShowAlbumPage)

	// Handle the delete album route.
	app.Post("/delete/{id:uint64}", srv.HandleDeleteAlbumRoute)

	// Handle the edit album page for a particular album.
	app.Post("/edit/{id:uint64}", srv.HandleEditAlbumRoute)

	// Set Iris to listen on a specified port.
	app.Listen(srv.HTTPPort)

}

// ================================ GET ROUTES ================================

/*
 * ShowHomePage handles a GET request for the "/" route. This page is shown
 * when the user first starts up the application.
 *
 * It sets the view to "home.html".
 */
func (srv *FrontendServer) ShowHomePage(ctx iris.Context) {
	log.Println("GET:		/")

	albums := srv.GetAllAlbums()
	ctx.View("home.html", iris.Map{
		"AlbumDB": albums,
	})
}

/*
 * GetAllAlbums returns all the albums in the key-value store
 */
func (srv *FrontendServer) GetAllAlbums() []*Album {
	request := &DataMessage{
		Method: "GetAllAlbums",
	}

	srv.WriteMessage(request)
	response := srv.ReadMessage()

	return response.AlbumArray
}

/*
 * ShowAlbumPage handles a GET request for the "/album/{id}" route. This page
 * is shown when the user requests to view a specific album.
 *
 * It sets the view to "album.html".
 */
func (srv *FrontendServer) ShowAlbumPage(ctx iris.Context) {
	albumID, _ := ctx.Params().GetUint64("id")
	albumIDString := strconv.Itoa(int(albumID))
	log.Print("GET:		/album/" + albumIDString)

	// Retrieve the album.
	request := &DataMessage{
		Method: "GetAlbum",
		Index:  albumIDString,
	}
	response := srv.WriteAndReadMessage(request)
	album := response.AlbumArray[0]

	// Set the HTML elements equal to the values in the album struct.
	ctx.ViewData("Title", album.Title)
	ctx.ViewData("Artist", album.Artist)
	ctx.ViewData("Year", album.Year)
	ctx.ViewData("Url", album.URL)
	ctx.ViewData("Id", album.Id)

	// Set the view.
	ctx.View("album.html")
}

/*
 * ShowAddPage handles a GET request for the "/add" route. This page is shown
 * when the user wants to add a new album.
 *
 * It sets the view to "add.html".
 */
func (srv *FrontendServer) ShowAddPage(ctx iris.Context) {
	log.Println("GET:		/add")
	ctx.View("add.html")
}

// ================================ POST ROUTES ===============================

/*
 * HandleAddAlbumRoute handles a POST request for the "/add" route.
 *
 * It retrieves values from the form and then makes a AddAlbum request to the
 * backend server
 */
func (srv *FrontendServer) HandleAddAlbumRoute(ctx iris.Context) {
	log.Print("POST:	/add")

	// Retrieve the values from the HTML form.
	title := ctx.PostValue("title")
	artist := ctx.PostValue("artist")
	url := ctx.PostValue("url")
	year := ctx.PostValue("year")

	// Call the AddAlbum function to add the album.
	album := &Album{
		Title:  title,
		Artist: artist,
		URL:    url,
		Year:   year,
	}

	request := &DataMessage{
		Method:     "AddAlbum",
		AlbumArray: []*Album{album},
	}

	response := srv.WriteAndReadMessage(request)
	if !response.Status {
		os.Exit(1)
	}

	// Return to the homepage.
	ctx.Redirect("/")
}

/*
 * HandleDeleteAlbumRoute handles a POST request for the "/delete/{id}" route.
 *
 * It makes a DeleteAlbum request to the backend server with the ID of the
 * album that will be deleted.
 */
func (srv *FrontendServer) HandleDeleteAlbumRoute(ctx iris.Context) {
	// Log the route.
	albumID, _ := ctx.Params().GetUint64("id")
	albumIDString := strconv.Itoa(int(albumID))
	log.Print("POST:	/delete/" + albumIDString)

	request := &DataMessage{
		Method: "DeleteAlbum",
		Index:  albumIDString,
	}

	response := srv.WriteAndReadMessage(request)
	if !response.Status {
		log.Fatal("HandleDeleteAlbum")
	}

	ctx.Redirect("/")
}

/*
 * HandleEditAlbumRoute handles a POST request for the "/edit/{id}" route.
 *
 * It retrieves values from the form and then makes an album struct and makes a
 * EditAlbum request to the backend server with the album struct.
 */
func (srv *FrontendServer) HandleEditAlbumRoute(ctx iris.Context) {
	// Log the route.
	albumID, _ := ctx.Params().GetUint64("id")
	albumIDString := strconv.Itoa(int(albumID))
	log.Print("POST:	/edit/" + albumIDString)

	// Get the values of the form.
	title := ctx.PostValue("title")
	artist := ctx.PostValue("artist")
	url := ctx.PostValue("url")
	year := ctx.PostValue("year")

	album := &Album{
		Title:  title,
		Artist: artist,
		URL:    url,
		Year:   year,
	}

	// Send a request to edit album.
	request := &DataMessage{
		Method:     "EditAlbum",
		Index:      albumIDString,
		AlbumArray: []*Album{album},
	}
	response := srv.WriteAndReadMessage(request)
	if !response.Status {
		log.Fatalln("Error editing album")
	}

	// Return to the homepage.
	ctx.Redirect("/")
}

// ============================ READ/WRITE MESSAGES ===========================

/*
 * ReadMessage receives a message from the backend server by decoding the bytes
 * sent over a TCP connection.
 */
func (srv *FrontendServer) ReadMessage() *DataMessage {
	msg := &DataMessage{}

	decoder := gob.NewDecoder(srv.Conn)
	if err := decoder.Decode(msg); err != nil {
		panic(err)
	}

	fmt.Println("[FrontendServer] received", msg)
	return msg
}

/*
 * WriteMessage sends a message to the backend server by encoding a DataMessage
 * struct into bytes and sending it over a TCP connection.
 */
func (srv *FrontendServer) WriteMessage(msg *DataMessage) {
	log.Println("[FrontendServer] sending", msg)

	encoder := gob.NewEncoder(srv.Conn)
	if err := encoder.Encode(msg); err != nil {
		panic(err)
	}
}

// ====================== FRONTEND/BACKEND COMMUNICATION ======================

/*
 * WriteAndReadMessage is a wrapper function to send a request to and recieve a
 * response from the backend server.
 */
func (srv *FrontendServer) WriteAndReadMessage(request *DataMessage) *DataMessage {
	srv.WriteMessage(request)
	return srv.ReadMessage()
}

func (srv *FrontendServer) PickRandom() string {
	return srv.Endpoints[rand.Intn(len(srv.Endpoints))]
}

/*
 * ConnectToBackend connects the frontend server to the backend server by
 * dialing a TCP connection.
 */
func (srv *FrontendServer) ConnectToBackend(address string) {
	tcp, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)

	}

	srv.Conn = conn
}

func (srv *FrontendServer) AskForLeader() {

}

func (srv *FrontendServer) FindLeader() {
	// Pick a random backend to connect to.
	curr := srv.PickRandom()

	//
	srv.ConnectToBackend(curr)

}

// ========================= MAIN & PARSING FUNCTIONS =========================

/*
 * ParseFrontendCommandLineArgs parses the command line flags used to invoike
 * the program and returns the HTTP port and the TCP endpoints.
 */
func ParseFrontendCommandLineArgs() (string, []string) {
	args := os.Args
	endPoints := []string{}
	httpPort := ":8080"
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
	httpPort, endpoints := ParseFrontendCommandLineArgs()

	srv := NewFrontendServer(httpPort, endpoints)
	srv.Start()
}
