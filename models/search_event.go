package models

import "time"

// SearchEvent represents a search event.
type SearchEvent struct {
	SearchID  string    `json:"search_id" gorm:"column:search_id"`
	Query     string    `json:"search_query" gorm:"column:search_query"`
	Timestamp time.Time `json:"-" gorm:"column:timestamp;type:datetime;default:CURRENT_TIMESTAMP;"`
}
