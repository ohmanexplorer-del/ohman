package github

import (
	"log"
	"time"
)

func (c *Crawler) loadQueue() error {
	if c.store == nil {
		return nil
	}

	queued, err := c.store.GetQueuedUsers()
	if err != nil {
		return err
	}

	for _, username := range queued {
		if c.shouldQueueUser(username) {
			c.queue = append(c.queue, username)
			c.visitedUsers[username] = true
		}
	}
	return nil
}

func (c *Crawler) shouldQueueUser(username string) bool {
	if username == "" {
		return false
	}
	if c.visitedUsers[username] {
		return false
	}
	if c.visitedRepos[username] {
		return false
	}
	if c.store != nil {
		if c.store.IsUserQueued(username) {
			return false
		}
		if c.store.HasUser(username) {
			return false
		}
	}
	return true
}

func (c *Crawler) waitForRateLimit() {
	limits, _, err := c.client.client.RateLimit.Get(c.ctx)
	if err != nil {
		time.Sleep(60 * time.Second)
		return
	}
	waitTime := time.Until(limits.Core.Reset.Time) + 10*time.Second
	if waitTime < 30*time.Second {
		waitTime = 30 * time.Second
	}
	if waitTime > 30*time.Minute {
		waitTime = 30 * time.Minute
	}
	log.Printf("Waiting %v for rate limit reset", waitTime.Round(time.Second))
	select {
	case <-c.ctx.Done():
	case <-time.After(waitTime):
	}
}

func (c *Crawler) Stop() {
	c.cancel()
}

func (c *Crawler) DecisionsMade() int {
	return c.decisions
}
