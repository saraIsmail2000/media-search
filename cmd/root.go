package cmd

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"media-search/middleware"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "media-search",
	Short: "Media Search Application",
	Long:  `Media Search Application`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initViper() *viper.Viper {
	var config = viper.New()
	config.AutomaticEnv()
	config.SetConfigFile(".env")
	if err := config.ReadInConfig(); err != nil {
		panic(err)
	}
	return config
}

func initDBConnection(config *viper.Viper) *gorm.DB {
	// Open database connection
	user := config.GetString("MYSQL_USER")
	pass := config.GetString("MYSQL_PASSWORD")
	dbName := config.GetString("MYSQL_DATABASE")
	dbPort := config.GetString("MYSQL_PORT")
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, dbPort, dbName)
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		logrus.Infof("Error connecting to database %v", err)
	}
	logrus.Info("Database connection established")
	return db
}

func initRedisCache(config *viper.Viper) *middleware.RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: config.GetString("REDIS_SERVER"),
	})
	return &middleware.RedisCache{Client: client}

}

func initElasticSearchClient() *elasticsearch.Client {
	// Initialize Elasticsearch Client
	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		logrus.Infof("failed to connect to elastic search client: %v", err.Error())
	}
	return client
}
