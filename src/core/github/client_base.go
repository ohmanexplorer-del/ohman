package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	gh "github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client    *gh.Client
	http      *http.Client
	token     string
	username  string
	repoName  string
	config    ClientConfig
	lastLimit *gh.RateLimits
}

func NewClient(cfg ClientConfig) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Token},
	)
	base := http.DefaultTransport
	if cfg.Cache != nil {
		base = &etagTransport{
			base:  base,
			cache: cfg.Cache,
		}
	}
	tc := &http.Client{
		Transport: &oauth2.Transport{
			Source: ts,
			Base:   base,
		},
		Timeout: 30 * time.Second,
	}

	return &Client{
		client:   gh.NewClient(tc),
		http:     tc,
		token:    cfg.Token,
		username: cfg.BotUsername,
		repoName: cfg.BotRepo,
		config:   cfg,
	}
}

func (c *Client) Token() string {
	return c.token
}

func (c *Client) BotUsername() string {
	return c.username
}

func (c *Client) BotRepo() string {
	return c.repoName
}

func (c *Client) CheckRateLimit(ctx context.Context) (remaining int, safe bool, err error) {
	limits, _, err := c.client.RateLimit.Get(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("failed to check rate limit: %w", err)
	}
	c.lastLimit = limits
	coreRemaining := limits.Core.Remaining
	if coreRemaining <= c.config.RateLimitBuffer {
		resetAt := limits.Core.Reset.Time
		log.Printf("Rate limit low: %d remaining (buffer: %d), resets at %s",
			coreRemaining, c.config.RateLimitBuffer, resetAt.Format(time.RFC3339))
		return coreRemaining, false, nil
	}
	return coreRemaining, true, nil
}
