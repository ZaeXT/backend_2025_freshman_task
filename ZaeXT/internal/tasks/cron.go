package tasks

import (
	"ai-qa-backend/internal/service"
	"log"

	"github.com/robfig/cron/v3"
)

func StartCronJobs(services *service.Service) *cron.Cron {
	c := cron.New(cron.WithSeconds())

	_, err := c.AddFunc("0 0 4 * * *", func() {
		log.Println("Cron Job [CleanupRecycleBin] started...")
		deletedCount, err := services.RecycleBin.CleanupExpired()
		if err != nil {
			log.Printf("Cron Job [CleanupRecycleBin] ERROR: %v", err)
		} else {
			log.Printf("Cron Job [CleanupRecycleBin] finished. Permanently deleted %d conversations.", deletedCount)
		}
	})
	if err != nil {
		log.Fatalf("Failed to add cron job [CleanupRecycleBin]: %v", err)
	}

	go c.Start()
	log.Println("Cron job scheduler started.")
	return c
}
