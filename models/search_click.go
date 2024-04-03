package models

import "time"

// SearchClick represents a click event.
type SearchClick struct {
	SearchClickId  int       `json:"-" gorm:"column:search_click_id;type:int;AUTO_INCREMENT"`
	SearchID       string    `json:"search_id" gorm:"column:search_id"`
	ResultType     string    `json:"result_type" gorm:"column:result_type"`
	ResultID       int       `json:"result_id" gorm:"column:result_id"`
	ResultPosition int       `json:"result_position" gorm:"column:result_position"`
	Timestamp      time.Time `json:"-" gorm:"column:timestamp;type:datetime;default:CURRENT_TIMESTAMP;"`
}
