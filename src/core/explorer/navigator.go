package explorer

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

type Navigator struct {
	crawler     *Crawler
	aiService   AIService
	config      NavigatorConfig
	decisionLog []NavigationDecision
	logMutex    sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
}

type NavigatorConfig struct {
	ExplorationStrategy string
	TargetTopics        []string
	MaxDecisions        int
	DecisionInterval    time.Duration
}

type AIService interface {
	ShouldNavigate(ctx context.Context, currentURL, targetURL string) (bool, string, error)
	GenerateNextAction(ctx context.Context, currentPage PageData) (string, error)
}

type NavigationDecision struct {
	Timestamp  time.Time
	FromURL    string
	ToURL      string
	Reason     string
	Confidence float64
}

func NewNavigator(crawler *Crawler, aiService AIService, config NavigatorConfig) *Navigator {
	ctx, cancel := context.WithCancel(context.Background())

	return &Navigator{
		crawler:     crawler,
		aiService:   aiService,
		config:      config,
		decisionLog: make([]NavigationDecision, 0),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (n *Navigator) Start(initialURLs []string) error {
	log.Printf("Starting autonomous navigator with strategy: %s", n.config.ExplorationStrategy)

	n.crawler.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(link)

		if !n.crawler.IsURLValid(absoluteURL) {
			return
		}

		normalizedURL, err := n.crawler.NormalizeURL(absoluteURL)
		if err != nil {
			log.Printf("Failed to normalize URL %s: %v", absoluteURL, err)
			return
		}

		shouldNavigate, reason, err := n.makeNavigationDecision(e.Request.URL.String(), normalizedURL)
		if err != nil {
			log.Printf("Failed to make navigation decision: %v", err)
			return
		}

		if shouldNavigate {
			n.logDecision(e.Request.URL.String(), normalizedURL, reason)
			n.crawler.queue <- normalizedURL
		}
	})

	return n.crawler.Start(initialURLs)
}

func (n *Navigator) makeNavigationDecision(fromURL, toURL string) (bool, string, error) {
	switch n.config.ExplorationStrategy {
	case "random":
		return n.randomDecision(toURL)
	case "focused":
		return n.focusedDecision(toURL)
	case "ai-guided":
		return n.aiGuidedDecision(fromURL, toURL)
	default:
		return n.randomDecision(toURL)
	}
}

func (n *Navigator) randomDecision(_ string) (bool, string, error) {
	if rand.Float32() < 0.3 {
		return true, "Random exploration", nil
	}
	return false, "Skipped (random)", nil
}

func (n *Navigator) focusedDecision(toURL string) (bool, string, error) {
	for _, topic := range n.config.TargetTopics {
		if contains(toURL, topic) {
			return true, fmt.Sprintf("Matches target topic: %s", topic), nil
		}
	}
	return false, "Does not match target topics", nil
}

func (n *Navigator) aiGuidedDecision(fromURL, toURL string) (bool, string, error) {
	if n.aiService == nil {
		return n.randomDecision(toURL)
	}

	shouldNavigate, reason, err := n.aiService.ShouldNavigate(n.ctx, fromURL, toURL)
	if err != nil {
		log.Printf("AI decision failed, falling back to random: %v", err)
		return n.randomDecision(toURL)
	}

	return shouldNavigate, reason, nil
}

func (n *Navigator) logDecision(fromURL, toURL, reason string) {
	n.logMutex.Lock()
	defer n.logMutex.Unlock()

	decision := NavigationDecision{
		Timestamp:  time.Now(),
		FromURL:    fromURL,
		ToURL:      toURL,
		Reason:     reason,
		Confidence: rand.Float64(),
	}

	n.decisionLog = append(n.decisionLog, decision)

	if len(n.decisionLog) > n.config.MaxDecisions {
		n.decisionLog = n.decisionLog[1:]
	}

	log.Printf("Navigation decision: %s -> %s (Reason: %s)", fromURL, toURL, reason)
}

func (n *Navigator) Stop() {
	log.Println("Stopping autonomous navigator...")
	n.cancel()
	n.crawler.Stop()
	log.Println("Autonomous navigator stopped")
}

func (n *Navigator) GetDecisionLog() []NavigationDecision {
	n.logMutex.Lock()
	defer n.logMutex.Unlock()

	return n.decisionLog
}

func (n *Navigator) GetStats() NavigatorStats {
	crawlerStats := n.crawler.GetStats()
	n.logMutex.Lock()
	defer n.logMutex.Unlock()

	return NavigatorStats{
		CrawlerStats:  crawlerStats,
		DecisionsMade: len(n.decisionLog),
		Strategy:      n.config.ExplorationStrategy,
	}
}

type NavigatorStats struct {
	CrawlerStats  CrawlerStats
	DecisionsMade int
	Strategy      string
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
