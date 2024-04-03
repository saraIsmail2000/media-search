package middleware

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"math/rand"
	"media-search/models"
	"net/http"
	"time"
)

type SearchAPI struct {
	db                  *gorm.DB
	redisClient         *RedisCache
	elasticSearchClient *elasticsearch.Client
}

func NewSearchAPI(db *gorm.DB, redisCache *RedisCache, elasticSearchClient *elasticsearch.Client) *SearchAPI {
	return &SearchAPI{
		db:                  db,
		redisClient:         redisCache,
		elasticSearchClient: elasticSearchClient,
	}
}

// SearchResult represents the combined search result of books and movies.
type SearchResult struct {
	SearchID string      `json:"search_id"`
	Cache    bool        `json:"cache"`
	Results  interface{} `json:"results"`
}

// SaveSearchEvent Report search event
// @Description Report that a search has been done
// @ID save-search
// @Accept json
// @Produce json
// @Param event body models.SearchEvent true "Search Event"
// @Success 200 {string} string "Search event reported successfully"
// @Router /save-search [post]
func (s *SearchAPI) SaveSearchEvent(c echo.Context) error {
	var event models.SearchEvent
	if err := c.Bind(&event); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	// Insert search event into database
	err := s.db.Create(event).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to save search event")
	}

	return c.JSON(http.StatusOK, "Search event reported successfully")
}

// SaveClickEvent Report click event
// @Description Report a click from the search results
// @ID report-click
// @Accept json
// @Produce json
// @Param event body models.SearchClick true "Click Event"
// @Success 200 {string} string "Click event reported successfully"
// @Router /save-click [post]
func (s *SearchAPI) SaveClickEvent(c echo.Context) error {
	var event models.SearchClick
	if err := c.Bind(&event); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	// Insert click event into database
	err := s.db.Create(event).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to save click event")
	}

	return c.JSON(http.StatusOK, "Click event reported successfully")
}

func ToPtr[T any](v T) *T {
	return &v
}

// HandleSearch Search function
// @Description Search through movies and books
// @ID search
// @Accept json
// @Produce json
// @Param search_query query string true "Search Query"
// @Success 200 {string} string "Click event reported successfully"
// @Router /search [get]
func (s *SearchAPI) HandleSearch(c echo.Context) error {
	var ctx = context.Background()
	var searchEvent models.SearchEvent

	searchQuery := c.QueryParam("search_query")
	// find search event with same search query with limit one
	err := s.db.Where("search_query = ?", searchQuery).Order("timestamp desc").First(&searchEvent).Error

	if err == nil {
		//Check if the search result is in the redisClient by checking search_id from searchEvent as the key
		cachedResult, found := s.redisClient.getFromCache(ctx, searchEvent.SearchID)
		if found {
			return c.JSON(200, SearchResult{
				SearchID: searchEvent.SearchID,
				Cache:    true,
				Results:  cachedResult,
			})
		}
	}

	//Perform search query against Elasticsearch
	result, err := applyElasticSearch(c.Request().Context(), s.elasticSearchClient, searchQuery)
	if err != nil {
		return err
	}

	searchID := generateSearchID()
	// Cache the search result for 30 seconds
	s.redisClient.saveToCache(ctx, searchID, result, 30*time.Second)

	return c.JSON(200, SearchResult{
		SearchID: searchID,
		Cache:    false,
		Results:  result,
	})
}

func generateSearchID() string {
	return fmt.Sprintf("%v", rand.Int())
}
