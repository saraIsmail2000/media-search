package scheduler

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Scheduler(db *gorm.DB) {
	c := cron.New()

	// run insights job daily -> every 24H
	_, err := c.AddFunc("0 0 * * *", func() {
		saveInsightsData(db)
	})
	if err != nil {
		logrus.Infof("Error adding cron job %v", err)
		return
	}
	logrus.Info("Start Scheduler and execute insights every 24h")
	c.Start()
}
