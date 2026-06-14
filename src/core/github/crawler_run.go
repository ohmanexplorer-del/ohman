package github

import (
	"log"
	"time"
)

func (c *Crawler) Run(sessionID int64) ([]*RepoData, error) {
	return c.RunLimit(sessionID, 0)
}

func (c *Crawler) RunLimit(sessionID int64, limit int) ([]*RepoData, error) {
	if limit <= 0 {
		limit = c.config.MaxReposPerRun
	}
	if limit <= 0 {
		limit = 10
	}
	log.Printf("GitHub crawler starting in '%s' mode with %d seed users",
		c.config.Mode, len(c.config.SeedUsers))
	c.currentSessionID = sessionID
	c.repos = nil
	c.decisions = 0
	c.discoveryLimit = limit
	defer func() {
		c.discoveryLimit = 0
	}()

	c.searchRepositories(limit)
	if limit > 0 && len(c.repos) >= limit {
		log.Printf("GitHub crawler finished from search: %d repos discovered, %d decisions made",
			len(c.repos), c.decisions)
		return c.repos, nil
	}

	for _, user := range c.config.SeedUsers {
		if c.shouldQueueUser(user) {
			if sessionID > 0 && c.store != nil {
				if err := c.store.EnqueueUserWithSession(user, sessionID); err != nil {
					log.Printf("Failed to enqueue seed user %s: %v", user, err)
				}
			} else {
				c.queue = append(c.queue, user)
				c.visitedUsers[user] = true
				if c.store != nil {
					if err := c.store.EnqueueUser(user); err != nil {
						log.Printf("Failed to enqueue seed user %s: %v", user, err)
					}
				}
			}
		}
	}

	usersVisited := 0
	for usersVisited < c.config.MaxUsersPerSession {
		if limit > 0 && len(c.repos) >= limit {
			break
		}
		if sessionID <= 0 && len(c.queue) == 0 {
			break
		}
		select {
		case <-c.ctx.Done():
			return c.repos, nil
		default:
		}

		if _, safe, _ := c.client.CheckRateLimit(c.ctx); !safe {
			log.Printf("Rate limit reached, pausing crawler")
			c.waitForRateLimit()
			continue
		}

		var username string
		var ok bool
		if sessionID > 0 && c.store != nil {
			u, found, err := c.store.PopNextUser(&sessionID)
			if err != nil {
				log.Printf("Failed to pop next user from queue: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
			if !found {
				log.Printf("No pending users in queue for session %d", sessionID)
				break
			}
			username = u
			ok = true
		} else {
			if len(c.queue) == 0 {
				break
			}
			username = c.queue[0]
			c.queue = c.queue[1:]
			ok = true
			if c.store != nil {
				if err := c.store.MarkUserProcessing(username); err != nil {
					log.Printf("Failed to mark user %s processing: %v", username, err)
				}
			}
		}

		if !ok {
			break
		}

		if err := c.exploreUser(username); err != nil {
			log.Printf("Failed to explore user %s: %v", username, err)
		}

		if sessionID > 0 && c.store != nil {
			if err := c.store.MarkUserProcessed(username); err != nil {
				log.Printf("Failed to mark user %s processed: %v", username, err)
			}
		}

		usersVisited++
	}

	log.Printf("GitHub crawler finished: %d users visited, %d repos discovered, %d decisions made",
		usersVisited, len(c.repos), c.decisions)
	return c.repos, nil
}
