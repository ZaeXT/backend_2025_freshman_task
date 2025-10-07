package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"AIBackend/internal/db"
	"AIBackend/internal/httpserver"
	"AIBackend/internal/provider"
)

func main() {
	// Load .env if present (dev convenience)
	_ = godotenv.Load()

	// Initialize DB
	pgURL := os.Getenv("DATABASE_URL")
	if pgURL == "" {
		log.Println("WARNING: DATABASE_URL is not set. The server may fail to start when DB is required.")
	}
	gormDB, err := db.Connect(pgURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := db.AutoMigrate(gormDB); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Initialize LLM provider (Mock by default)
	llm := provider.NewProviderFromEnv()

	// Start HTTP server
	r := httpserver.NewRouter(gormDB, llm)
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("Server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
