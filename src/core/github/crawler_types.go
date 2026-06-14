package github

import (
	"context"
	storage "ohman/src/core/db"
)

type Crawler struct {
	client           *Client
	config           ExplorationConfig
	ai               AIService
	store            *storage.DB
	ctx              context.Context
	cancel           context.CancelFunc
	visitedUsers     map[string]bool
	visitedRepos     map[string]bool
	queue            []string
	repos            []*RepoData
	decisions        int
	currentSessionID int64
	discoveryLimit   int
}

type AIService interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

type repoEvaluation struct {
	Category     string   `json:"category"`
	ProjectType  string   `json:"project_type"`
	Score        float64  `json:"score"`
	Novelty      float64  `json:"novelty"`
	Maturity     float64  `json:"maturity"`
	SmallRepoFit float64  `json:"small_repo_fit"`
	Reason       string   `json:"reason"`
	Strengths    []string `json:"strengths"`
	Weaknesses   []string `json:"weaknesses"`
	Publish      bool     `json:"publish"`
}

func NewCrawler(client *Client, config ExplorationConfig, ai AIService, store *storage.DB) *Crawler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Crawler{
		client:       client,
		config:       config,
		ai:           ai,
		store:        store,
		ctx:          ctx,
		cancel:       cancel,
		visitedUsers: make(map[string]bool),
		visitedRepos: make(map[string]bool),
	}
}
