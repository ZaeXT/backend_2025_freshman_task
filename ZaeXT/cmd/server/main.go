package main

import (
	"ai-qa-backend/internal/adapter/volcengine"
	"ai-qa-backend/internal/configs"
	"ai-qa-backend/internal/handler"
	"ai-qa-backend/internal/repository"
	"ai-qa-backend/internal/repository/db"
	"ai-qa-backend/internal/service"
	"ai-qa-backend/internal/tasks"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := configs.Init(); err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}
	gormDB, err := db.Init()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	repos := repository.NewRepository(gormDB)

	aiAdapter := volcengine.NewVolcengineAdapter()

	services := service.NewService(repos, aiAdapter)

	cronScheduler := tasks.StartCronJobs(services)

	router := handler.SetupRouter(services)

	addr := fmt.Sprintf(":%d", configs.Conf.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cronCtx := cronScheduler.Stop()
	<-cronCtx.Done()
	log.Println("Cron scheduler stopped.")

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err == nil {
		sqlDB.Close()
		log.Println("Database connection closed.")
	}

	log.Println("Server exited successfully.")
}
