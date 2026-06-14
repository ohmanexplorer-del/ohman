package explorer

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

type Crawler struct {
	collector    *colly.Collector
	config       CrawlerConfig
	visited      map[string]bool
	visitedMutex sync.Mutex
	queue        chan string
	maxPages     int
	pagesVisited int
	ctx          context.Context
	cancel       context.CancelFunc
}

type CrawlerConfig struct {
	UserAgent         string
	MaxDepth          int
	MaxPages          int
	Delay             time.Duration
	AllowedDomains    []string
	DisallowedDomains []string
	FollowLinks       bool
	RespectRobotsTxt  bool
	Parallelism       int
}

type PageData struct {
	URL       string
	Title     string
	Content   string
	Links     []string
	VisitedAt time.Time
	Depth     int
}

func NewCrawler(config CrawlerConfig) (*Crawler, error) {
	c := colly.NewCollector(
		colly.UserAgent(config.UserAgent),
		colly.MaxDepth(config.MaxDepth),
		colly.Async(true),
	)

	c.AllowURLRevisit = false

	ctx, cancel := context.WithCancel(context.Background())

	crawler := &Crawler{
		collector: c,
		config:    config,
		visited:   make(map[string]bool),
		queue:     make(chan string, 1000),
		maxPages:  config.MaxPages,
		ctx:       ctx,
		cancel:    cancel,
	}

	if len(config.AllowedDomains) > 0 {
		c.AllowedDomains = config.AllowedDomains
	}

	if len(config.DisallowedDomains) > 0 {
		c.DisallowedDomains = config.DisallowedDomains
	}

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       config.Delay,
		RandomDelay: config.Delay / 2,
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error visiting %s: %v", r.Request.URL, err)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %s", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		crawler.visitedMutex.Lock()
		crawler.pagesVisited++
		crawler.visitedMutex.Unlock()
	})

	return crawler, nil
}

func (c *Crawler) Start(initialURLs []string) error {
	log.Printf("Starting crawler with %d initial URLs", len(initialURLs))

	for _, url := range initialURLs {
		c.queue <- url
	}

	for i := 0; i < c.config.Parallelism; i++ {
		go c.worker()
	}

	return nil
}

func (c *Crawler) worker() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case url := <-c.queue:
			c.visitedMutex.Lock()
			if c.visited[url] || c.pagesVisited >= c.maxPages {
				c.visitedMutex.Unlock()
				continue
			}
			c.visited[url] = true
			c.visitedMutex.Unlock()

			err := c.collector.Visit(url)
			if err != nil {
				log.Printf("Failed to visit %s: %v", url, err)
			}
		}
	}
}

func (c *Crawler) Stop() {
	log.Println("Stopping crawler...")
	c.cancel()
	c.collector.Wait()
	close(c.queue)
	log.Println("Crawler stopped")
}

func (c *Crawler) GetStats() CrawlerStats {
	c.visitedMutex.Lock()
	defer c.visitedMutex.Unlock()

	return CrawlerStats{
		PagesVisited: c.pagesVisited,
		MaxPages:     c.maxPages,
	}
}

type CrawlerStats struct {
	PagesVisited int
	MaxPages     int
}

func (c *Crawler) IsURLValid(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	if parsedURL.Host == "" {
		return false
	}

	return true
}

func (c *Crawler) NormalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	parsedURL.Fragment = ""

	parsedURL.Host = parsedURL.Hostname()

	return parsedURL.String(), nil
}
