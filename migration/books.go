package migration

import (
	"context"
	"encoding/csv"
	"github.com/elastic/go-elasticsearch/v7"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io"
	"media-search/constants"
	"media-search/middleware"
	"media-search/models"
	"os"
	"strconv"
	"sync"
	"time"
)

// fetchBooksFromDB read all books existing in Db
func fetchBooksFromDB(db *gorm.DB) map[int]models.Book {
	dataMap := make(map[int]models.Book)
	var books []models.Book

	err := db.Find(&books).Error
	if err != nil {
		logrus.Infof("Error querying database: %v", err)
	}

	for _, record := range books {
		dataMap[record.ID] = record
	}

	return dataMap
}

// readBooksFile reads data from CSV file and sends records over channel
func readBooksFile(filename string, channel chan<- models.Book) error {
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
		if record[0] == "bookID" {
			continue
		}
		id, _ := strconv.Atoi(record[0])
		rating, _ := strconv.ParseFloat(record[3], 32)
		numPages, _ := strconv.Atoi(record[7])
		ratingsCount, _ := strconv.Atoi(record[8])
		textReviewsCount, _ := strconv.Atoi(record[9])
		publishDate, _ := time.Parse("1/2/2006", record[10])
		var book = models.Book{
			ID:               id,
			Title:            record[1],
			Authors:          record[2],
			AverageRating:    rating,
			ISBN:             record[4],
			ISBN13:           record[5],
			LanguageCode:     record[6],
			NumPages:         numPages,
			RatingsCount:     ratingsCount,
			TextReviewsCount: textReviewsCount,
			PublicationDate:  publishDate,
			Publisher:        record[11],
		}
		// Send book over channel
		channel <- book
	}

	return nil
}

// processBookRecord processes each record received from channel and performs database operation
func processBookRecord(db *gorm.DB, record models.Book, booksMap map[int]models.Book, processedBooks *[]int, esClient *elasticsearch.Client) {
	// Check if the movie exists in the map
	existingBook, exists := booksMap[record.ID]

	if !exists {
		// If the movie doesn't exist in the map, create a new record
		if err := db.Create(&record).Error; err != nil {
			logrus.Infof("Error creating movie %v ", record.ID)
		}
	} else {
		// If the movie exists, update its fields
		if err := db.Model(&existingBook).Updates(record).Error; err != nil {
			logrus.Infof("Error updating movie %v ", record.ID)
		}
	}
	// insert into elastic search index
	middleware.AddMovieOrBookToIndex(context.Background(), esClient, middleware.MapBookToIndexedData(record), constants.ElasticSearchBooksIndex)
	logrus.Infof("book %v is processed", record.ID)
	*processedBooks = append(*processedBooks, record.ID)
}

func deleteMissingBooksFromDB(db *gorm.DB, esClient *elasticsearch.Client, processedBooks []int) {
	var toDelete []int
	db.Model(models.Book{}).Select("book_id").Where("book_id NOT IN (?)", processedBooks).Find(&toDelete)

	if len(toDelete) > 0 {
		// deleting extra books from database
		db.Where("book_id IN (?)", toDelete).Delete(&models.Book{})
		// deleting extra books from elastic
		middleware.DeleteDocumentsFromIndex(esClient, constants.ElasticSearchBooksIndex, toDelete)
	}
}

func migrateBooks(db *gorm.DB, esClient *elasticsearch.Client) {
	var wg sync.WaitGroup

	booksMap := fetchBooksFromDB(db)
	processedBooks := make([]int, 0)

	// Create buffered channels for communication
	booksFile := make(chan models.Book, 1000)
	booksToProcess := make(chan models.Book, 1000)

	// Semaphore channel to limit concurrency
	semaphore := make(chan struct{}, 100) // Limit to 100 concurrent goroutines

	go func() {
		err := readBooksFile("books.csv", booksFile)
		if err != nil {
			panic(err)
		}
		close(booksFile)
	}()

	// Start pool of goroutines to handle database operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for record := range booksToProcess {
				semaphore <- struct{}{} // Acquire semaphore (decrement)
				processBookRecord(db, record, booksMap, &processedBooks, esClient)
				<-semaphore // Release semaphore (increment)
			}
		}()
	}

	// Process records received from file channels and send to db channel
	go func() {
		for record := range booksFile {
			booksToProcess <- record
		}
		close(booksToProcess)
	}()

	wg.Wait()
	close(semaphore)
	logrus.Info("books imported")

	// delete the books remaining in database and not found in the csv file
	deleteMissingBooksFromDB(db, esClient, processedBooks)
}
