package db

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(dsn string) (*DB, error) {
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Printf("database connected and migrated")
	return db, nil
}

func (db *DB) Migrate() error {
	return db.conn.AutoMigrate(
		&repoRow{},
		&userRow{},
		&sessionRow{},
		&edgeRow{},
		&crawlQueueRow{},
		&configRow{},
		&accountRow{},
		&githubHTTPCacheRow{},
	)
}
