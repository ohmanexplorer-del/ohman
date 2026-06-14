package app

import (
	"log"
	"os"
	"sync"
	"time"

	"ohman/config"
	api "ohman/src/controllers"
	"ohman/src/core/ai"
	"ohman/src/core/collector"
	"ohman/src/core/contributor"
	storage "ohman/src/core/db"
	"ohman/src/core/explorer"
	ght "ohman/src/core/github"
	"ohman/src/core/publisher"
	"ohman/src/core/scheduler"
	"ohman/src/core/telegram"
)

type App struct {
	Config      *config.Config
	AI          *ai.Service
	DB          *storage.DB
	GitHub      *ght.Client
	GitHubCrawl *ght.Crawler
	Collector   *collector.Collector
	Contributor *contributor.Service
	Publisher   *publisher.Publisher
	Navigator   *explorer.Navigator
	Scheduler   *scheduler.Scheduler
	Telegram    *telegram.Service
	API         *api.Server

	githubEnabled bool
	botRunning    bool
	botStartedAt  *time.Time
	botStoppedAt  *time.Time
	botLastError  string
	botMu         sync.Mutex

	githubDiscoveryMu     sync.Mutex
	lastGitHubDiscoveryAt *time.Time
}

func New(configPath string) (*App, error) {
	dsn := os.Getenv("OHMAN_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://ohman:ohman_secret@localhost:5432/ohman?sslmode=disable"
	}

	a := &App{}

	db, err := storage.New(dsn)
	if err != nil {
		return nil, err
	}
	a.DB = db

	cfg, err := config.LoadFromDB(a.DB)
	if err != nil {
		log.Printf("warning: failed to load config from DB: %v", err)
		cfg = &config.Config{}
	}
	a.Config = cfg
	a.githubEnabled = cfg.GitHub.Token != ""

	if cfg.API.Enabled {
		a.initAPI()
	}
	a.initTelegram()

	return a, nil
}
