package main

import (
	"log"
	"strconv"

	"github.com/kataras/iris/v12"
)

// ================================ POST ROUTES ===============================

/*
 * HandleAddAlbumRoute handles a POST request for the "/add" route.
 *
 * It retrieves values from the form and then adds the album struct to the
 * in-memory database.
 */
func HandleAddAlbumRoute(ctx iris.Context) {
	// Log the route.
	log.Print("POST:	/add")

	// Retrieve the values from the HTML form.
	title := ctx.PostValue("title")
	artist := ctx.PostValue("artist")
	url := ctx.PostValue("url")
	year := ctx.PostValue("year")

	// Call the AddAlbum function to add the album.
	AddAlbum(title, artist, url, year)

	// Return to the homepage.
	ctx.Redirect("/")
}

/*
 * HandleDeleteAlbumRoute handles a POST request for the "/delete/{id}" route.
 *
 * It calls the RemoveAlbum function.
 */
func HandleDeleteAlbumRoute(ctx iris.Context) {
	// Log the route.
	albumID, _ := ctx.Params().GetUint64("id")
	albumIDString := strconv.Itoa(int(albumID))
	log.Print("POST:	/delete/" + albumIDString)

	// Call the RemoveAlbum function to remove the album.
	if err := RemoveAlbum(albumIDString); err != nil {
		log.Fatal("HandleDeleteAlbum: ", err)
	}

	// Return to the homepage.
	ctx.Redirect("/")
}

/*
 * HandleEditAlbumRoute handles a POST request for the "/edit/{id}" route.
 *
 * It retrieves values from the form and then adds the album struct to the
 * in-memory database.
 */
func HandleEditAlbumRoute(ctx iris.Context) {
	// Log the route.
	albumID, _ := ctx.Params().GetUint64("id")
	albumIDString := strconv.Itoa(int(albumID))
	log.Print("POST:	/edit/" + albumIDString)

	// Get the values of the form.
	title := ctx.PostValue("title")
	artist := ctx.PostValue("artist")
	url := ctx.PostValue("url")
	year := ctx.PostValue("year")

	// Call the EditAlbum function to edit the album.
	EditAlbum(albumIDString, title, artist, url, year)

	// Return to the homepage.
	ctx.Redirect("/")
}

// ================================ GET ROUTES ================================

/*
 * ShowHomePage handles a GET request for the "/" route. This page is shown
 * when the user first starts up the application.
 *
 * It sets the view to "home.html".
 */
func ShowHomePage(ctx iris.Context) {
	// Log the route.
	log.Print("GET:		/")

	// Set the view.
	ctx.View("home.html", iris.Map{
		"AlbumDB": AlbumDB,
	})

}

/*
 * ShowAddPage handles a GET request for the "/add" route. This page is shown
 * when the user wants to add a new album.
 *
 * It sets the view to "add.html".
 */
func ShowAddPage(ctx iris.Context) {
	// Log the route.
	log.Print("GET:		/add")

	// Set the view.
	ctx.View("add.html")
}

/*
 * ShowAlbumPage handles a GET request for the "/album/{id}" route. This page
 * is shown when the user requests to view a specific album.
 *
 * It sets the view to "album.html".
 */
func ShowAlbumPage(ctx iris.Context) {
	// Log the route.
	albumID, _ := ctx.Params().GetUint64("id")
	albumIDString := strconv.Itoa(int(albumID))
	log.Print("GET:		/album/" + albumIDString)

	// Retrieve the album.
	album := AlbumDB[int(albumID)]

	// Set the HTML elements equal to the values in the album struct.
	ctx.ViewData("Title", album.Title)
	ctx.ViewData("Artist", album.Artist)
	ctx.ViewData("Year", album.Year)
	ctx.ViewData("Url", album.URL)
	ctx.ViewData("Id", album.Id)

	// Set the view.
	ctx.View("album.html")
}
