package migration

import (
	"context"
	"encoding/csv"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io"
	"media-search/constants"
	"media-search/middleware"
	"media-search/models"
	"os"
	"strconv"
	"sync"
)

// fetchMoviesFromDB read all movies existing in DB
func fetchMoviesFromDB(db *gorm.DB) map[int]models.Movie {
	dataMap := make(map[int]models.Movie)
	var movies []models.Movie

	err := db.Find(&movies).Error
	if err != nil {
		panic(err)
	}

	for _, record := range movies {
		dataMap[record.ID] = record
	}

	return dataMap
}

// readMoviesFile reads data from CSV file and sends records over channel
func readMoviesFile(filename string, channel chan<- models.Movie) error {
	// Open CSV file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)

	// Parse CSV records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Skip header row
		if record[0] == "movieID" {
			continue
		}

		id, _ := strconv.Atoi(record[0])
		year, _ := strconv.Atoi(record[2])
		runtime, _ := strconv.Atoi(record[6])
		rating, _ := strconv.ParseFloat(record[8], 32)
		var movie = models.Movie{
			ID:             id,
			Title:          record[1],
			PublishYear:    year,
			Summary:        record[3],
			ShortSummary:   record[4],
			IMDBID:         record[5],
			Runtime:        runtime,
			YouTubeTrailer: record[7],
			Rating:         rating,
			MoviePoster:    record[9],
			Director:       record[10],
			Writers:        record[11],
			Cast:           record[12],
		}
		// Send movie over channel
		channel <- movie
	}
	return nil
}

// processMovieRecord processes each record received from channel and performs database operation
func processMovieRecord(db *gorm.DB, record models.Movie, existingMovies map[int]models.Movie, processedMovies *[]int, esClient *elasticsearch.Client) {
	// Check if the movie exists in the map
	existingMovie, exists := existingMovies[record.ID]

	if !exists {
		// If the movie doesn't exist in the map, create a new record
		if err := db.Create(&record).Error; err != nil {
			logrus.Infof("Error creating movie %v ", record.ID)
		}
	} else {
		// If the movie exists, update its fields
		if err := db.Model(&existingMovie).Updates(record).Error; err != nil {
			logrus.Infof("Error updating movie %v ", record.ID)
		}
	}
	// insert into elastic search index
	middleware.AddMovieOrBookToIndex(context.Background(), esClient, middleware.MapMovieToIndexedData(record), constants.ElasticSearchMoviesIndex)
	logrus.Infof("movie %v is processed", record.ID)
	*processedMovies = append(*processedMovies, record.ID)

}

func deleteMissingMoviesFromDB(db *gorm.DB, esClient *elasticsearch.Client, processedMovies []int) {
	var toDelete []int
	db.Model(models.Movie{}).Select("movie_id").Where("movie_id NOT IN (?)", processedMovies).Find(&toDelete)

	if len(toDelete) > 0 {
		// deleting extra movies from database
		db.Where("movie_id IN (?)", toDelete).Delete(&models.Movie{})
		// deleting extra movies from elastic
		middleware.DeleteDocumentsFromIndex(esClient, constants.ElasticSearchMoviesIndex, toDelete)
	}
}

func migrateMovies(db *gorm.DB, esClient *elasticsearch.Client) {
	var wg sync.WaitGroup

	existingMovies := fetchMoviesFromDB(db)
	processedMovies := make([]int, 0)

	// Create buffered channels for communication
	moviesFile := make(chan models.Movie, 1000)
	moviesToProcess := make(chan models.Movie, 1000)

	// Semaphore channel to limit concurrency
	semaphore := make(chan struct{}, 100) // Limit to 100 concurrent goroutines

	// Start goroutine to read CSV files concurrently and close moviesFile channel when done
	go func() {
		err := readMoviesFile("movies.csv", moviesFile)
		if err != nil {
			panic(err)
		}
		close(moviesFile)
	}()

	// Start pool of goroutines to handle database operations with semaphore limiting
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for record := range moviesToProcess {
				semaphore <- struct{}{} // Acquire semaphore (decrement)
				processMovieRecord(db, record, existingMovies, &processedMovies, esClient)
				<-semaphore // Release semaphore (increment)
			}
		}()
	}

	// Process records received from file channels and send to db channel
	go func() {
		for record := range moviesFile {
			moviesToProcess <- record
		}
		close(moviesToProcess)
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	close(semaphore)
	logrus.Info("movies imported")

	// delete the movies remaining in database and not found in the csv file
	deleteMissingMoviesFromDB(db, esClient, processedMovies)
}
