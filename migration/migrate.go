package migration

import (
	"github.com/elastic/go-elasticsearch/v7"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	_ "net/http/pprof"
	"sync"
	"time"
)

func Migrate(db *gorm.DB, esClient *elasticsearch.Client) {
	var wg sync.WaitGroup
	wg.Add(2)

	started := time.Now()
	go func() {
		defer wg.Done()
		migrateMovies(db, esClient)
	}()

	go func() {
		defer wg.Done()
		migrateBooks(db, esClient)
	}()

	wg.Wait()
	ended := time.Now()
	logrus.Infof("job of importing movies and books data from CSV files to DB is done in %v minutes.", ended.Sub(started).Minutes())
}
