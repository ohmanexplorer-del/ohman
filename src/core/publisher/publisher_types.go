package publisher

import (
	"context"

	storage "ohman/src/core/db"
	ght "ohman/src/core/github"
)

type Publisher struct {
	db       *storage.DB
	ghClient *ght.Client
	ctx      context.Context
}

func NewPublisher(db *storage.DB, ghClient *ght.Client) *Publisher {
	return &Publisher{
		db:       db,
		ghClient: ghClient,
		ctx:      context.Background(),
	}
}
