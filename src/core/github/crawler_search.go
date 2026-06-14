package github

import (
	"log"
	"strings"
)

func (c *Crawler) searchRepositories(limit int) {
	queries := c.config.SearchQueries
	if len(queries) == 0 {
		return
	}

	for _, query := range queries {
		if limit > 0 && len(c.repos) >= limit {
			return
		}
		if _, safe, _ := c.client.CheckRateLimit(c.ctx); !safe {
			log.Printf("Rate limit reached before search query")
			return
		}

		remaining := limit - len(c.repos)
		if limit <= 0 {
			remaining = c.config.MaxReposPerRun
		}
		if remaining <= 0 {
			remaining = 10
		}

		repos, err := c.client.SearchRepos(c.ctx, query, &RepoListOptions{
			PerPage:   min(max(remaining*2, 10), 100),
			MaxRepos:  remaining * 2,
			Sort:      "updated",
			Direction: "desc",
		})
		if err != nil {
			log.Printf("GitHub search failed for query %q: %v", query, err)
			continue
		}

		for _, repo := range repos {
			if c.acceptSearchRepo(repo) {
				c.repos = append(c.repos, repo)
			}
			if limit > 0 && len(c.repos) >= limit {
				return
			}
		}
	}
}

func (c *Crawler) acceptSearchRepo(repo *RepoData) bool {
	if repo == nil || repo.FullName == "" {
		return false
	}
	if repo.Stars < c.config.MinStars {
		return false
	}
	if c.visitedRepos[repo.FullName] {
		return false
	}
	if c.store != nil && c.store.HasRepo(repo.FullName) {
		return false
	}

	c.visitedRepos[repo.FullName] = true
	if c.ai != nil {
		c.evaluateRepo(repo)
		c.decisions++
	}
	if ok, reasons := qualityGate(repo); !ok {
		log.Printf("Skipping repo %s: %s", repo.FullName, strings.Join(reasons, "; "))
		return false
	}
	return true
}
