package main

import (
	"errors"
	"strconv"
)

// A global variable keeping track of the current ID (primary key) of the
// database. This has to be changed later on since this is a single point of
// failure and we need a better way to create unique primary keys.
var currID = 0

// A global variable which is the in-memory database implemented as a map from
// integers to an album pointer.
var AlbumDB map[int]*Album

// Album is a struct representing an album.
type Album struct {
	Id     string
	Title  string
	Artist string
	URL    string
	Year   string
}

// hardcodedAlbums is a 2D slice of strings where each individual slice is an
// album's metadata; this is used to store a hardcoded list of albums.
var hardcodedAlbums = [][]string{
	{"Disintigration", "The Cure", "https://lastfm.freetls.fastly.net/i/u/770x0/a0f446f0184f425e52fcdb32b9cf82e5.jpg", "1989"},
	{"Kids See Ghosts", "Kids See Ghosts", "https://lastfm.freetls.fastly.net/i/u/770x0/e9dd5c8d3294ca0a0f58cbf7ad5fd6a6.jpg", "2018"},
	{"Devotion", "Tirzah", "https://lastfm.freetls.fastly.net/i/u/770x0/1961645688c754bd7a26bd540b9f7a7d.jpg", "2018"},
	{"Untouched", "Secret Shine", "https://lastfm.freetls.fastly.net/i/u/770x0/9ff3fccfcb6cc587dca7e9bcbd551024.jpg", "1993"},
	{"Purple Haze", "Cam'ron", "https://lastfm.freetls.fastly.net/i/u/770x0/3025393c10b6cc84bf85cba203bdb7f6.jpg", "2004"},
}

/*
 * InitializeHardcodedAlbums initializes the AlbumDB with hardcoded albums.
 */
func InitializeHardcodedAlbums() {
	AlbumDB = make(map[int]*Album)
	for _, album := range hardcodedAlbums {
		AddAlbum(album[0], album[1], album[2], album[3])
	}
}

/*
 * AddAlbum adds a new album struct to our in-memory database.
 */
func AddAlbum(title, artist, url, year string) {
	AlbumDB[currID] = &Album{
		Id:     strconv.Itoa(currID),
		Title:  title,
		Artist: artist,
		URL:    url,
		Year:   year,
	}

	// Increment the ID by 1 for the next AddAlbum call.
	currID += 1
}

/*
 * RemoveAlbum removes an album struct from our in-memory database.
 *
 * Returns an error if the ID is not valid or if there isn't an album
 * associated with the given ID.
 */
func RemoveAlbum(id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	if _, ok := AlbumDB[idInt]; ok {
		delete(AlbumDB, idInt)
	} else {
		return errors.New("Album does not exist")
	}
	return nil
}

/*
 * EditAlbum retrieves an album using its ID and then edits that album's fields
 * to be updated with the given album fields if they are non-empty. If they are
 * empty, the fields are not modified.
 *
 * Returns an error if the ID is not valid or if there isn't an album
 * associated with the given ID.
 */
func EditAlbum(id, title, artist, url, year string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	if _, ok := AlbumDB[idInt]; ok {
		// Retrieve the album using the ID.
		a := AlbumDB[idInt]

		// For each field, if the given value is mon-empty, update the fields
		// using the new value; otherwise, leave the fields as is.
		if title != "" {
			a.Title = title
		}
		if artist != "" {
			a.Artist = artist
		}
		if url != "" {
			a.URL = url
		}
		if year != "" {
			a.Year = year
		}
	} else {
		return errors.New("Album does not exist")
	}
	return nil
}

/*
 * GetAlbum retrieves an album using its ID.
 *
 * Also returns an error if the ID is not valid or if there isn't an album
 * associated with the given ID.
 */
func GetAlbum(id string) (*Album, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	if _, ok := AlbumDB[idInt]; ok {
		a := AlbumDB[idInt]
		return a, nil
	} else {
		return nil, errors.New("Album does not exist")
	}
}
