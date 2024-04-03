package models

import "time"

// Book represents a book record.
type Book struct {
	ID               int       `json:"book_id" csv:"bookID" gorm:"column:book_id;primaryKey"`
	Title            string    `json:"title" csv:"title" gorm:"column:title"`
	Authors          string    `json:"authors" csv:"authors" gorm:"column:authors"`
	AverageRating    float64   `json:"average_rating" csv:"average_rating" gorm:"column:average_rating"`
	ISBN             string    `json:"isbn" csv:"isbn" gorm:"column:isbn"`
	ISBN13           string    `json:"isbn13" csv:"isbn13" gorm:"column:isbn13"`
	LanguageCode     string    `json:"language_code" csv:"language_code" gorm:"column:language_code"`
	NumPages         int       `json:"num_pages" csv:"num_pages" gorm:"column:num_pages"`
	RatingsCount     int       `json:"ratings_count" csv:"ratings_count" gorm:"column:ratings_count"`
	TextReviewsCount int       `json:"text_reviews_count" csv:"text_reviews_count" gorm:"column:text_reviews_count"`
	PublicationDate  time.Time `json:"publication_date" csv:"publication_date" gorm:"column:publication_date"`
	Publisher        string    `json:"publisher" csv:"publisher" gorm:"column:publisher"`
}
