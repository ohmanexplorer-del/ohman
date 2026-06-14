package collector

import (
	"context"
	"fmt"
	"log"
	"strings"

	storage "ohman/src/core/db"
	"ohman/src/core/github"
	"ohman/src/core/prompts"
)

type Collector struct {
	db       *storage.DB
	ghClient *github.Client
	ai       AIService
	ctx      context.Context
}

type AIService interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

func NewCollector(db *storage.DB, ghClient *github.Client, ai AIService) *Collector {
	return &Collector{
		db:       db,
		ghClient: ghClient,
		ai:       ai,
		ctx:      context.Background(),
	}
}

func (c *Collector) Collect(repos []*github.RepoData) (int, error) {
	saved := 0
	for _, repo := range repos {
		if err := c.collectOne(repo); err != nil {
			log.Printf("Failed to collect repo %s: %v", repo.FullName, err)
			continue
		}
		saved++
	}
	log.Printf("Collector saved %d/%d repos", saved, len(repos))
	return saved, nil
}

func (c *Collector) collectOne(repo *github.RepoData) error {
	parts := strings.SplitN(repo.FullName, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid full_name: %s", repo.FullName)
	}
	owner, name := parts[0], parts[1]

	if c.ai != nil && repo.AIScore == 0 {
		readme, err := c.ghClient.GetReadme(c.ctx, owner, name)
		if err != nil {
			log.Printf("  No README for %s: %v", repo.FullName, err)
		} else {
			summary, err := c.summarizeReadme(repo.FullName, repo.Description, readme)
			if err != nil {
				log.Printf("  README summary failed for %s: %v", repo.FullName, err)
			} else {
				repo.ReadmeSummary = summary
			}
		}

		if len(repo.Topics) == 0 {
			topics, err := c.ghClient.GetRepoTopics(c.ctx, owner, name)
			if err != nil {
				log.Printf("  Topics fetch failed for %s: %v", repo.FullName, err)
			} else {
				repo.Topics = topics
			}
		}
	}

	if err := c.db.UpsertRepo(&storage.RepoData{
		GithubID:      repo.GithubID,
		FullName:      repo.FullName,
		Owner:         repo.Owner,
		Description:   repo.Description,
		Stars:         repo.Stars,
		Forks:         repo.Forks,
		Language:      repo.Language,
		Topics:        repo.Topics,
		License:       repo.License,
		DefaultBranch: repo.DefaultBranch,
		Homepage:      repo.Homepage,
		VisitedAt:     repo.VisitedAt,
		CreatedAt:     repo.CreatedAt,
		UpdatedAt:     repo.UpdatedAt,
		PushedAt:      repo.PushedAt,
		AIScore:       repo.AIScore,
		AINotes:       repo.AINotes,
		ReadmeSummary: repo.ReadmeSummary,
		Category:      repo.Category,
		ProjectType:   repo.ProjectType,
		Novelty:       repo.Novelty,
		Maturity:      repo.Maturity,
		SmallRepoFit:  repo.SmallRepoFit,
		Strengths:     repo.Strengths,
		Weaknesses:    repo.Weaknesses,
		Publish:       repo.Publish,
	}); err != nil {
		return fmt.Errorf("db upsert failed: %w", err)
	}
	return nil
}

func (c *Collector) summarizeReadme(fullName, description, readme string) (string, error) {

	readmePreview := readme
	if len(readmePreview) > 2000 {
		readmePreview = readmePreview[:2000]
	}

	prompt, err := prompts.RenderGitHubSummarizeReadme(prompts.GitHubSummarizeReadmePromptData{
		FullName:      fullName,
		Description:   description,
		ReadmePreview: readmePreview,
	})
	if err != nil {
		return "", err
	}

	response, err := c.ai.GenerateContent(c.ctx, prompt)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(response), nil
}

func (c *Collector) EnrichRepo(repo *github.RepoData) error {
	return c.collectOne(repo)
}
