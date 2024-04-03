package middleware

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/gorm"
)

func SetupRouters(e *echo.Echo, db *gorm.DB, redisCache *RedisCache, elasticSearchClient *elasticsearch.Client) {
	context := "/api/v1"
	searchAPI := NewSearchAPI(db, redisCache, elasticSearchClient)

	e.GET(context+"/swagger/*", echoSwagger.WrapHandler)

	e.GET("/search", searchAPI.HandleSearch)
	e.POST("/save-search", searchAPI.SaveSearchEvent)
	e.POST("/save-click", searchAPI.SaveClickEvent)
	e.POST("/integrate", searchAPI.SaveClickEvent)

}
