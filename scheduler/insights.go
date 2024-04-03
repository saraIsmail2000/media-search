package scheduler

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"log"
	"media-search/constants"
	"os"
	"sync"
	"time"
)

// ClickedItem represents a clicked item
type ClickedItem struct {
	ID     int    `json:"result_id" gorm:"result_id"`
	Type   string `json:"result_type" gorm:"result_type"`
	Clicks int    `json:"clicks" gorm:"clicks"`
}

// Insights represents the daily insights
type Insights struct {
	TopClickedItems  []ClickedItem `json:"top_clicked_items"`
	AvgClickPosition float64       `json:"avg_click_position"`
	TotalSearches    int           `json:"total_searches"`
	TotalClicks      int           `json:"total_clicks"`
	CTR              float64       `json:"ctr"`
}

func saveInsightsData(db *gorm.DB) {
	var topClickedItems []ClickedItem
	var averageClickPositions float64
	var searches, clicks int

	// fetch all dta concurrently
	var wg sync.WaitGroup
	wg.Add(4)

	// Fetch top 10 clicked items
	go func() {
		defer wg.Done()
		if err := db.Raw(`
			SELECT result_id as id, result_type as type, COUNT(*) as clicks
			FROM search_clicks
			WHERE timestamp >= DATE_SUB(NOW(), INTERVAL 24 HOUR)
			GROUP BY result_id, result_type
			ORDER BY clicks DESC
			LIMIT 10;
		`).Scan(&topClickedItems).Error; err != nil {
			return
		}
	}()

	// Fetch average click position per day
	go func() {
		defer wg.Done()
		if err := db.Raw(`
			SELECT AVG(result_position) AS avg_position
			FROM search_clicks
			WHERE timestamp >= DATE_SUB(NOW(), INTERVAL 24 HOUR);
		`).Scan(&averageClickPositions).Error; err != nil {
			return
		}
	}()

	// Fetch total searches per day
	go func() {
		defer wg.Done()
		if err := db.Raw(`
			SELECT COUNT(*) AS searches
			FROM search_events
			WHERE timestamp >= DATE_SUB(NOW(), INTERVAL 24 HOUR);
		`).Scan(&searches).Error; err != nil {
			return
		}
	}()

	// Fetch total clicks per day
	go func() {
		defer wg.Done()
		if err := db.Raw(`
			SELECT COUNT(*) AS clicks
			FROM search_clicks
			WHERE timestamp >= DATE_SUB(NOW(), INTERVAL 24 HOUR)
		`).Scan(&clicks).Error; err != nil {
			log.Printf("Error executing total searches and clicks query: %v", err)
			return
		}
	}()

	wg.Wait()
	var ctr = float64(clicks) / float64(searches) * 100

	insightsData := Insights{
		TopClickedItems:  topClickedItems,
		AvgClickPosition: averageClickPositions,
		TotalSearches:    searches,
		TotalClicks:      clicks,
		CTR:              ctr,
	}

	// Marshal insights data to JSON
	jsonData, err := json.MarshalIndent(insightsData, "", "    ")
	if err != nil {
		log.Fatalf("Error marshalling insights data to JSON: %v", err)
	}

	currentDate := time.Now().Format("2006-01-02")

	// Create the directory if it does not exist
	if _, err := os.Stat(constants.JsonDir); os.IsNotExist(err) {
		if err := os.MkdirAll(constants.JsonDir, 0755); err != nil {
			log.Fatalf("Error creating JSON directory: %v", err)
		}
	}

	// Create a new JSON file with the current date in the filename
	jsonFile, err := os.Create(fmt.Sprintf("%s%s.json", constants.JsonDir, currentDate))
	if err != nil {
		log.Fatalf("Error creating JSON file: %v", err)
	}
	defer jsonFile.Close()

	// Write JSON data to the file
	if _, err := jsonFile.Write(jsonData); err != nil {
		log.Fatalf("Error writing JSON data to file: %v", err)
	}

	logrus.Info("Insights data saved to %s%s.json\n", constants.JsonDir, currentDate)
}
