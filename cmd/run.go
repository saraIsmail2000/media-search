package cmd

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"
	"media-search/docs"
	"media-search/middleware"
	"media-search/scheduler"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "run",
	Short: "run the echo server",
	Long:  "run the echo server",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	config := initViper()
	db := initDBConnection(config)

	e := echo.New()
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())

	docs.SwaggerInfo.Host = config.GetString("SWAGGER_HOST")

	// init redis client connection
	redisCache := initRedisCache(config)
	elasticSearchClient := initElasticSearchClient()

	// add middleware and routes
	middleware.SetupRouters(e, db, redisCache, elasticSearchClient)

	// run scheduler for cron jobs
	go scheduler.Scheduler(db)

	//run application
	go func() {
		if err := e.Start("localhost:8081"); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	return nil
}
