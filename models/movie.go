package models

// Movie represents a movie record.
type Movie struct {
	ID             int     `json:"movie_id" csv:"movieID" gorm:"column:movie_id"`
	Title          string  `json:"title" csv:"title" gorm:"column:title"`
	PublishYear    int     `json:"publish_year" csv:"year" gorm:"column:publish_year"`
	Summary        string  `json:"summary" csv:"summary" gorm:"column:summary"`
	ShortSummary   string  `json:"short_summary" csv:"short_summary" gorm:"column:short_summary"`
	IMDBID         string  `json:"imdb_id" csv:"imdb_id" gorm:"column:imdb_id"`
	Runtime        int     `json:"runtime" csv:"runtime" gorm:"column:runtime"`
	YouTubeTrailer string  `json:"youtube_trailer" csv:"youtube_trailer" gorm:"column:youtube_trailer"`
	Rating         float64 `json:"rating" csv:"rating" gorm:"column:rating"`
	MoviePoster    string  `json:"movie_poster" csv:"movie_poster" gorm:"column:movie_poster"`
	Director       string  `json:"director" csv:"director" gorm:"column:director"`
	Writers        string  `json:"writers" csv:"writers" gorm:"column:writers"`
	Cast           string  `json:"cast" csv:"cast" gorm:"column:cast"`
}
