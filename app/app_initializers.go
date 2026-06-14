package app

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

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

func (a *App) initAI() error {
	cfg := a.Config.AI
	svc, err := ai.NewService(ai.AIConfig{
		BaseURL:     cfg.BaseURL,
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		TopP:        cfg.TopP,
		Stream:      cfg.Stream,
	})
	if err != nil {
		return err
	}
	a.AI = svc
	log.Println("AI service initialized")
	return nil
}

func (a *App) initStorage() error {
	db, err := storage.New(a.Config.Storage.DatabaseURL)
	if err != nil {
		return err
	}
	a.DB = db
	log.Println("Storage initialized")
	return nil
}

func (a *App) initGitHub() {
	if !a.githubEnabled {
		log.Println("GitHub token not configured, GitHub features disabled")
		return
	}

	gh := ght.NewClient(ght.ClientConfig{
		Token:           a.Config.GitHub.Token,
		BotUsername:     a.Config.GitHub.BotUsername,
		BotRepo:         a.Config.GitHub.BotRepo,
		RateLimitBuffer: a.Config.GitHub.RateLimitBuffer,
		Cache:           a.DB,
		GraphQL:         a.Config.GitHub.GraphQL,
	})
	a.GitHub = gh
	log.Println("GitHub client initialized")
	if err := gh.EnsureBotRepoReady(context.Background()); err != nil {
		log.Printf("warning: failed to prepare GitHub bot repo: %v", err)
	} else {
		log.Println("GitHub bot repo ready")
	}

	a.GitHubCrawl = ght.NewCrawler(gh, ght.ExplorationConfig{
		Mode:               a.Config.Exploration.Mode,
		MinStars:           a.Config.Exploration.MinStars,
		Languages:          a.Config.Exploration.Languages,
		MaxUsersPerSession: a.Config.Exploration.MaxUsersPerSession,
		ReposPerUser:       a.Config.Exploration.ReposPerUser,
		MaxReposPerRun:     a.Config.Exploration.MaxReposPerRun,
		SeedUsers:          a.Config.GitHub.SeedUsers,
		SearchQueries:      a.Config.GitHub.SearchQueries,
		RateLimitBuffer:    a.Config.GitHub.RateLimitBuffer,
	}, a.AI, a.DB)
	log.Println("GitHub crawler initialized")
}

func (a *App) initCollector() {
	if !a.githubEnabled {
		return
	}
	a.Collector = collector.NewCollector(a.DB, a.GitHub, a.AI)
	log.Println("Collector initialized")
}

func (a *App) initContributor() {
	if !a.githubEnabled || !a.Config.Contribution.Enabled {
		return
	}
	a.Contributor = contributor.NewService(a.GitHub, a.AI, contributor.Config{
		Enabled:         a.Config.Contribution.Enabled,
		AutoPR:          a.Config.Contribution.AutoPR,
		Labels:          a.Config.Contribution.Labels,
		MaxIssuesPerRun: a.Config.Contribution.MaxIssuesPerRun,
		MaxRepoSizeKB:   a.Config.Contribution.MaxRepoSizeKB,
		MaxRepoStars:    a.Config.Contribution.MaxRepoStars,
		MaxRepoForks:    a.Config.Contribution.MaxRepoForks,
		WorkDir:         a.Config.Contribution.WorkDir,
	})
	log.Println("Contributor initialized")
}

func (a *App) initPublisher() {
	if !a.githubEnabled {
		return
	}
	a.Publisher = publisher.NewPublisher(a.DB, a.GitHub)
	log.Println("Publisher initialized")
}

func (a *App) initWebExplorer() error {
	cfg := a.Config.Crawler
	crawler, err := explorer.NewCrawler(explorer.CrawlerConfig{
		UserAgent:         cfg.UserAgent,
		MaxDepth:          cfg.MaxDepth,
		MaxPages:          cfg.MaxPages,
		Delay:             cfg.Delay,
		AllowedDomains:    cfg.AllowedDomains,
		DisallowedDomains: cfg.DisallowedDomains,
		FollowLinks:       cfg.FollowLinks,
		RespectRobotsTxt:  cfg.RespectRobotsTxt,
		Parallelism:       cfg.Parallelism,
	})
	if err != nil {
		return err
	}

	aiNav := explorer.NewAINavigator(a.AI, explorer.AINavigatorConfig{
		MaxTokens:     a.Config.AI.MaxTokens,
		Temperature:   float32(a.Config.AI.Temperature),
		DecisionModel: a.Config.AI.Model,
	})

	a.Navigator = explorer.NewNavigator(crawler, aiNav, explorer.NavigatorConfig{
		ExplorationStrategy: "ai-guided",
		TargetTopics:        []string{},
		MaxDecisions:        1000,
		DecisionInterval:    5 * time.Minute,
	})
	log.Println("Web crawler components initialized")
	return nil
}

func (a *App) initScheduler() {
	cfg := a.Config
	sched := scheduler.NewScheduler()

	if a.githubEnabled {
		sched.AddTask(scheduler.Task{
			Name:     "github_exploration",
			Schedule: cfg.Schedule.ExplorationInterval,
			Func: func() error {
				return a.runGitHubDiscoveryOnce(a.githubReposPerRunLimit())
			},
		})
		log.Println("GitHub exploration task scheduled")

		if a.Contributor != nil {
			sched.AddTask(scheduler.Task{
				Name:     "github_contribution",
				Schedule: cfg.Schedule.ExplorationInterval,
				Func: func() error {
					return a.runContributionOnce()
				},
			})
			log.Println("GitHub contribution task scheduled")
		}
	}

	if !a.githubEnabled {
		sched.AddTask(scheduler.Task{
			Name:     "web_exploration",
			Schedule: cfg.Schedule.ExplorationInterval,
			Func: func() error {
				log.Println("Starting scheduled web exploration...")
				if err := a.Navigator.Start(cfg.Crawler.InitialURLs); err != nil {
					log.Printf("Web exploration failed: %v", err)
					return err
				}
				time.Sleep(30 * time.Minute)
				a.Navigator.Stop()

				stats := a.Navigator.GetStats()
				log.Printf("Web exploration stats: Pages visited: %d/%d, Decisions made: %d",
					stats.CrawlerStats.PagesVisited, stats.CrawlerStats.MaxPages, stats.DecisionsMade)
				return nil
			},
		})
		log.Println("Web exploration task scheduled")
	}

	a.Scheduler = sched
}

func (a *App) initAPI() {
	a.API = api.NewServer(a.DB.Conn(), a, a.apiPort())
}

func (a *App) initTelegram() {
	cfg := a.Config.Telegram
	if !cfg.Enabled || cfg.Token == "" {
		return
	}
	interval, err := time.ParseDuration(cfg.PollInterval)
	if err != nil {
		log.Printf("warning: invalid telegram.poll_interval=%q; using 3s", cfg.PollInterval)
		interval = 3 * time.Second
	}
	a.Telegram = telegram.NewService(telegram.Config{
		Enabled:         cfg.Enabled,
		Token:           cfg.Token,
		AllowedChatIDs:  cfg.AllowedChatIDs,
		ChannelChatID:   cfg.ChannelChatID,
		PollInterval:    interval,
		ActivityReports: cfg.ActivityReports,
	}, telegramController{app: a})
	log.Println("Telegram control initialized")
}

func (a *App) apiPort() int {
	if raw := os.Getenv("OHMAN_API_PORT"); raw != "" {
		port, err := strconv.Atoi(raw)
		if err == nil && port > 0 {
			return port
		}
		log.Printf("warning: invalid OHMAN_API_PORT=%q; using configured port", raw)
	}
	if a.Config != nil && a.Config.API.Port > 0 {
		return a.Config.API.Port
	}
	return 8081
}
