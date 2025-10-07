package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"AIBackend/internal/models"
)

// Connect opens a PostgreSQL connection using DATABASE_URL.
func Connect(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		// Provide a friendly default to help first run; it will still fail if DB not available.
		databaseURL = "postgres://postgres:postgres@localhost:5432/aibackend?sslmode=disable"
	}
	dsn := databaseURL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	return db, nil
}

// AutoMigrate applies database schema for all models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Conversation{},
		&models.Message{},
	)
}
