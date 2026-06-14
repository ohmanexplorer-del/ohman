package main

import (
	"log"
	"os"

	storage "ohman/src/core/db"
)

func main() {
	dsn := os.Getenv("OHMAN_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://ohman:ohman_secret@localhost:5432/ohman?sslmode=disable"
	}

	db, err := storage.New(dsn)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	defer db.Close()

	stats, err := db.Stats()
	if err != nil {
		log.Fatalf("migration completed but stats failed: %v", err)
	}
	log.Printf("migration completed: repos=%d users=%d sessions=%d social_graph=%d crawl_queue=%d configurations=%d accounts=%d",
		stats["repos"],
		stats["users"],
		stats["sessions"],
		stats["social_graph"],
		stats["crawl_queue"],
		stats["configurations"],
		stats["accounts"],
	)
}
