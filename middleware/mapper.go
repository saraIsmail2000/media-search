package middleware

import "media-search/models"

// IndexedData this represents the structure of books and movies data model as saved in elastic search indexes
type IndexedData struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Type     string  `json:"type"`
	Authors  *string `json:"authors,omitempty"`
	Director *string `json:"director,omitempty"`
	Writers  *string `json:"writers,omitempty"`
	Cast     *string `json:"cast,omitempty"`
}

// MapMovieToIndexedData maps Movie struct to IndexedData struct.
func MapMovieToIndexedData(movie models.Movie) IndexedData {
	var authors, director, writers, cast *string

	// If writers, director, or cast are empty strings, set them to nil pointers.
	if movie.Director != "" {
		director = &movie.Director
	}
	if movie.Writers != "" {
		writers = &movie.Writers
	}
	if movie.Cast != "" {
		cast = &movie.Cast
	}

	// Split writers by comma if there are multiple writers.
	if movie.Writers != "" {
		writers = &movie.Writers
	}

	return IndexedData{
		ID:       movie.ID,
		Title:    movie.Title,
		Type:     "movie",
		Authors:  authors,
		Director: director,
		Writers:  writers,
		Cast:     cast,
	}
}

// MapBookToIndexedData maps Book struct to IndexedData struct.
func MapBookToIndexedData(book models.Book) IndexedData {
	var authors *string

	// If any field is empty, set its pointer to nil.
	if book.Authors != "" {
		authors = &book.Authors
	}

	return IndexedData{
		ID:      book.ID,
		Title:   book.Title,
		Type:    "book",
		Authors: authors,
	}
}
