package main

import (
	"log"
	"os"

	"ohman/config"
	storage "ohman/src/core/db"
	ght "ohman/src/core/github"
	"ohman/src/core/publisher"
)

func main() {
	dsn := os.Getenv("OHMAN_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://ohman:ohman_secret@localhost:5432/ohman?sslmode=disable"
	}

	db, err := storage.New(dsn)
	if err != nil {
		log.Fatalf("database failed: %v", err)
	}
	defer db.Close()

	cfg, err := config.LoadFromDB(db)
	if err != nil {
		log.Fatalf("config failed: %v", err)
	}
	if cfg.GitHub.Token == "" {
		log.Fatal("github.token is required")
	}

	client := ght.NewClient(ght.ClientConfig{
		Token:           cfg.GitHub.Token,
		BotUsername:     cfg.GitHub.BotUsername,
		BotRepo:         cfg.GitHub.BotRepo,
		RateLimitBuffer: cfg.GitHub.RateLimitBuffer,
		Cache:           db,
		GraphQL:         cfg.GitHub.GraphQL,
	})

	if err := publisher.NewPublisher(db, client).PublishLibrary(); err != nil {
		log.Fatalf("publish failed: %v", err)
	}
}
