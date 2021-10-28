package main

import (
	"flag"

	"github.com/kataras/iris/v12"
)

var portFlag = flag.String("listen", "8080", "The port which Iris will listen to.")

func main() {
	// Initialize the album "database" with hardcoded albums.
	InitializeHardcodedAlbums()

	// Parse the flag.
	flag.Parse()
	port := *portFlag

	// Initialize an Iris app.
	app := iris.Default()

	// Register a folder for HTML templates.
	app.RegisterView(iris.HTML("./views", ".html"))

	// Show the homepage of the app.
	app.Get("/", ShowHomePage)

	// Handle the add album route.
	app.Post("/add", HandleAddAlbumRoute)

	// Show the add album page.
	app.Get("/add", ShowAddPage)

	// Show the album page for a particular album.
	app.Get("/album/{id:uint64}", ShowAlbumPage)

	// Handle the delete album route.
	app.Post("/delete/{id:uint64}", HandleDeleteAlbumRoute)

	// Handle the edit album page for a particular album.
	app.Post("/edit/{id:uint64}", HandleEditAlbumRoute)

	// Set Iris to listen on a specified port.
	app.Listen(":" + port)
}
