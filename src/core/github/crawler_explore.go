package github

import (
	"context"
	"fmt"
	"log"
	storage "ohman/src/core/db"
	"ohman/src/core/prompts"
	"strings"
	"time"
)

func (c *Crawler) exploreUser(username string) error {
	log.Printf("Exploring user: %s", username)

	if c.store != nil && c.store.HasUser(username) {
		log.Printf("User %s already visited in storage, skipping", username)
		return nil
	}

	user, err := c.client.GetUser(c.ctx, username)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	log.Printf("  Profile: %s (%s) - %d followers, %d repos",
		user.Name, user.Username, user.Followers, user.PublicRepos)

	if c.store != nil {
		if err := c.store.UpsertUser(&storage.UserData{
			GithubID:    user.GithubID,
			Username:    user.Username,
			Name:        user.Name,
			Bio:         user.Bio,
			Followers:   user.Followers,
			Following:   user.Following,
			PublicRepos: user.PublicRepos,
			VisitedAt:   user.VisitedAt,
		}); err != nil {
			log.Printf("Failed to persist user %s: %v", username, err)
		}
	}

	if c.ai != nil {
		if !c.shouldExploreUser(user) {
			log.Printf("Skipping user %s based on AI evaluation", username)
			return nil
		}
	}

	opts := &RepoListOptions{
		PerPage:  30,
		MaxRepos: c.config.ReposPerUser,
		Sort:     "updated",
	}
	repos, err := c.client.GetUserRepos(c.ctx, username, opts)
	if err != nil {
		return fmt.Errorf("failed to get repos: %w", err)
	}
	log.Printf("  Fetched %d repos", len(repos))

	for _, repo := range repos {
		if repo.Stars < c.config.MinStars {
			continue
		}
		if c.visitedRepos[repo.FullName] {
			continue
		}
		if c.store != nil && c.store.HasRepo(repo.FullName) {
			continue
		}
		c.visitedRepos[repo.FullName] = true

		if c.ai != nil {
			c.evaluateRepo(repo)
			c.decisions++
		}
		if ok, reasons := qualityGate(repo); !ok {
			log.Printf("Skipping repo %s: %s", repo.FullName, strings.Join(reasons, "; "))
			continue
		}

		c.repos = append(c.repos, repo)
		if c.discoveryLimit > 0 && len(c.repos) >= c.discoveryLimit {
			break
		}
	}

	c.expandGraph(username, user.Followers)
	return nil
}

type repoQualitySignals struct {
	HasCI           bool
	HasTests        bool
	HasContributing bool
	Archived        bool
}

func detectQualitySignals(ctx context.Context, client *Client, owner, repoName string) repoQualitySignals {
	contents, err := client.GetRepoContents(ctx, owner, repoName, "")
	if err != nil {
		return repoQualitySignals{}
	}
	var s repoQualitySignals
	for _, item := range contents {
		p := strings.ToLower(item.Path)
		switch p {
		case ".github", ".circleci", ".travis.yml", ".gitlab-ci.yml", "jenkinsfile", ".drone.yml", "azure-pipelines.yml":
			s.HasCI = true
		case "test", "tests", "spec", "__tests__", "testing", "e2e":
			s.HasTests = true
		case "contributing.md", "changelog.md", "code_of_conduct.md":
			s.HasContributing = true
		}
	}
	return s
}

// EvaluateRepo menjalankan evaluasi AI untuk satu repo, mengambil README dan sinyal kualitas dari GitHub.
// Hasilnya ditulis langsung ke field-field repo yang diberikan.
func EvaluateRepo(ctx context.Context, client *Client, ai AIService, repo *RepoData) {
	owner, repoName, _ := strings.Cut(repo.FullName, "/")

	readmePreview := ""
	if readme, err := client.GetReadme(ctx, owner, repoName); err == nil {
		readmePreview = trimToLength(readme, 1500)
	}

	signals := detectQualitySignals(ctx, client, owner, repoName)

	repoAgeDays := 0
	if !repo.CreatedAt.IsZero() {
		repoAgeDays = int(time.Since(repo.CreatedAt).Hours() / 24)
	}

	prompt, err := prompts.RenderGitHubEvaluateRepo(prompts.GitHubEvaluateRepoPromptData{
		FullName:        repo.FullName,
		Owner:           repo.Owner,
		Description:     repo.Description,
		Language:        repo.Language,
		Stars:           repo.Stars,
		Forks:           repo.Forks,
		OpenIssues:      repo.OpenIssues,
		SizeKB:          repo.SizeKB,
		Topics:          repo.Topics,
		License:         repo.License,
		RepoAgeDays:     repoAgeDays,
		HasHomepage:     repo.Homepage != "",
		Archived:        false,
		ReadmePreview:   readmePreview,
		HasCI:           signals.HasCI,
		HasTests:        signals.HasTests,
		HasContributing: signals.HasContributing,
	})
	if err != nil {
		log.Printf("failed to render GitHub repo evaluation prompt: %v", err)
		applyFallbackEvaluation(repo)
		return
	}

	response, err := ai.GenerateContent(ctx, prompt)
	if err != nil {
		log.Printf("AI evaluation failed for %s: %v", repo.FullName, err)
		applyFallbackEvaluation(repo)
		return
	}

	evaluation, err := parseRepoEvaluation(response)
	if err != nil {
		log.Printf("AI evaluation parse failed for %s: %v", repo.FullName, err)
		applyFallbackEvaluation(repo)
		repo.AINotes = trimToLength(response, 500)
		return
	}

	repo.Category = normalizeCategory(evaluation.Category)
	repo.ProjectType = strings.TrimSpace(evaluation.ProjectType)
	repo.AIScore = clampScore(evaluation.Score)
	repo.Novelty = clampScore(evaluation.Novelty)
	repo.Maturity = clampScore(evaluation.Maturity)
	repo.SmallRepoFit = clampScore(evaluation.SmallRepoFit)
	repo.AINotes = trimToLength(strings.TrimSpace(evaluation.Reason), 500)
	repo.Strengths = trimStringSlice(evaluation.Strengths, 5)
	repo.Weaknesses = trimStringSlice(evaluation.Weaknesses, 5)
	repo.Publish = evaluation.Publish
}

func (c *Crawler) evaluateRepo(repo *RepoData) {
	EvaluateRepo(c.ctx, c.client, c.ai, repo)
}


func (c *Crawler) shouldExploreUser(user *UserData) bool {
	prompt, err := prompts.RenderGitHubExploreUser(prompts.GitHubExploreUserPromptData{
		Username:    user.Username,
		Name:        user.Name,
		Bio:         user.Bio,
		Followers:   user.Followers,
		Following:   user.Following,
		PublicRepos: user.PublicRepos,
	})
	if err != nil {
		log.Printf("failed to render GitHub explore user prompt: %v", err)
		return true
	}

	response, err := c.ai.GenerateContent(c.ctx, prompt)
	if err != nil {
		log.Printf("AI user evaluation failed for %s: %v", user.Username, err)
		return true
	}

	response = strings.TrimSpace(strings.ToUpper(response))
	return !strings.HasPrefix(response, "SKIP")
}

func (c *Crawler) expandGraph(username string, followerCount int) {
	if len(c.queue)+c.config.MaxUsersPerSession/2 < c.config.MaxUsersPerSession*10 && followerCount > 0 {
		followers, err := c.client.GetFollowers(c.ctx, username, max(5, c.config.MaxUsersPerSession/10))
		if err != nil {
			log.Printf("Failed to get followers for %s: %v", username, err)
		} else {
			for _, f := range followers {
				if c.shouldQueueUser(f) {
					c.visitedUsers[f] = true
					c.queue = append(c.queue, f)
					if c.store != nil {
						if c.currentSessionID > 0 {
							if err := c.store.EnqueueUserWithSession(f, c.currentSessionID); err != nil {
								log.Printf("Failed to enqueue follower %s: %v", f, err)
							}
						} else {
							if err := c.store.EnqueueUser(f); err != nil {
								log.Printf("Failed to enqueue follower %s: %v", f, err)
							}
						}
					}
				}
				if c.store != nil {
					if err := c.store.InsertEdge(f, username, "follower"); err != nil {
						log.Printf("Failed to insert follower edge %s->%s: %v", f, username, err)
					}
				}
			}
			log.Printf("  Added %d followers to queue", len(followers))
		}
	}

	following, err := c.client.GetFollowing(c.ctx, username, max(5, c.config.MaxUsersPerSession/10))
	if err != nil {
		log.Printf("Failed to get following for %s: %v", username, err)
	} else {
		for _, f := range following {
			if c.shouldQueueUser(f) {
				c.visitedUsers[f] = true
				c.queue = append(c.queue, f)
				if c.store != nil {
					if c.currentSessionID > 0 {
						if err := c.store.EnqueueUserWithSession(f, c.currentSessionID); err != nil {
							log.Printf("Failed to enqueue following %s: %v", f, err)
						}
					} else {
						if err := c.store.EnqueueUser(f); err != nil {
							log.Printf("Failed to enqueue following %s: %v", f, err)
						}
					}
				}
			}
			if c.store != nil {
				if err := c.store.InsertEdge(username, f, "following"); err != nil {
					log.Printf("Failed to insert following edge %s->%s: %v", username, f, err)
				}
			}
		}
		log.Printf("  Added %d following to queue", len(following))
	}
}
