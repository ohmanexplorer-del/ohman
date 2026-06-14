package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"ohman/src/core/textutil"
	"sort"
	"strings"
	"time"

	storage "ohman/src/core/db"
	ght "ohman/src/core/github"
)

func (p *Publisher) PublishDiscoveryList() error {
	repos, err := p.db.GetVisitedRepos(1000)
	if err != nil {
		return fmt.Errorf("failed to get repos for publishing: %w", err)
	}
	if len(repos) == 0 {
		log.Println("No repos to publish")
		return nil
	}

	sort.Slice(repos, func(i, j int) bool {
		if repos[i].AIScore != repos[j].AIScore {
			return repos[i].AIScore > repos[j].AIScore
		}
		return repos[i].Stars > repos[j].Stars
	})

	readme := p.buildReadme(repos)
	if err := p.ghClient.CreateFile(p.ctx, "README.md",
		fmt.Sprintf("Update discovery list — %d repos", len(repos)),
		readme); err != nil {
		return fmt.Errorf("failed to publish README: %w", err)
	}

	dailyMarkdown := p.buildDailyLog(repos)
	dailyPath := fmt.Sprintf("daily/%s.md", time.Now().Format("2006-01-02"))
	if err := p.ghClient.CreateFile(p.ctx, dailyPath,
		fmt.Sprintf("Daily exploration log %s", time.Now().Format("2006-01-02")),
		dailyMarkdown); err != nil {
		log.Printf("Warning: failed to publish daily log: %v", err)
	}

	log.Printf("Published discovery list: README.md + %s", dailyPath)
	return nil
}

func (p *Publisher) PublishFindings(repos []*ght.RepoData) error {
	if len(repos) == 0 {
		log.Println("No GitHub findings to publish")
		return nil
	}
	if err := p.PublishLibrary(); err != nil {
		return err
	}
	log.Printf("Published MVP findings: %d new repos", len(repos))
	return nil
}

func (p *Publisher) PublishLibrary() error {
	library, err := p.libraryRepos()
	if err != nil {
		return err
	}
	if len(library) == 0 {
		log.Println("No repository library to publish")
		return nil
	}

	readme := p.buildFindingsReadme(library)
	if err := p.ghClient.CreateOrUpdateFile(p.ctx, "README.md",
		fmt.Sprintf("Update project library: %d repos", len(library)),
		readme); err != nil {
		return fmt.Errorf("failed to publish MVP findings: %w", err)
	}

	dailyPath := fmt.Sprintf("findings/%s.md", time.Now().Format("2006-01-02"))
	if err := p.ghClient.CreateOrUpdateFile(p.ctx, dailyPath,
		fmt.Sprintf("Update findings for %s", time.Now().Format("2006-01-02")),
		p.buildFindingsArchive(library)); err != nil {
		return fmt.Errorf("failed to publish findings archive: %w", err)
	}

	groups := groupByCategory(library)
	activeCategoryPaths := map[string]bool{}
	for _, category := range sortedCategories(groups) {
		categoryRepos := groups[category]
		path := fmt.Sprintf("categories/%s.md", category)
		activeCategoryPaths[path] = true
		if err := p.ghClient.CreateOrUpdateFile(p.ctx, path,
			fmt.Sprintf("Update %s category", category),
			p.buildCategoryPage(category, categoryRepos)); err != nil {
			return fmt.Errorf("failed to publish category %s: %w", category, err)
		}
	}
	p.cleanupStaleCategoryPages(activeCategoryPaths)

	data, err := json.MarshalIndent(library, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode repo data: %w", err)
	}
	if err := p.ghClient.CreateOrUpdateFile(p.ctx, "data/repos.json",
		fmt.Sprintf("Update repo data: %d repos", len(library)),
		string(data)+"\n"); err != nil {
		return fmt.Errorf("failed to publish repo data: %w", err)
	}

	log.Printf("Published project library: %d repos", len(library))
	return nil
}

func (p *Publisher) cleanupStaleCategoryPages(active map[string]bool) {
	contents, err := p.ghClient.GetRepoContents(p.ctx, p.ghClient.BotUsername(), p.ghClient.BotRepo(), "categories")
	if err != nil {
		log.Printf("warning: failed to inspect category pages for cleanup: %v", err)
		return
	}
	for _, item := range contents {
		if item == nil || item.Type != "file" || !strings.HasPrefix(item.Path, "categories/") || !strings.HasSuffix(item.Path, ".md") {
			continue
		}
		if active[item.Path] {
			continue
		}
		if err := p.ghClient.DeleteFile(p.ctx, item.Path, "Remove stale category page"); err != nil {
			log.Printf("warning: failed to delete stale category page %s: %v", item.Path, err)
		}
	}
}

func (p *Publisher) libraryRepos() ([]*ght.RepoData, error) {
	rows, err := p.db.GetVisitedRepos(5000)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository library: %w", err)
	}
	repos := make([]*ght.RepoData, 0, len(rows))
	for _, row := range rows {
		if !publishableStorageRepo(row) {
			continue
		}
		repos = append(repos, storageRepoToGitHub(row))
	}
	sortGithubRepos(repos)
	return repos, nil
}

func publishableStorageRepo(repo *storage.RepoData) bool {
	return repo != nil && (repo.Publish || repo.Category == "")
}

func storageRepoToGitHub(repo *storage.RepoData) *ght.RepoData {
	ageDays, ageLabel := repoAge(repo.CreatedAt, time.Now().UTC())
	return &ght.RepoData{
		GithubID:      repo.GithubID,
		FullName:      repo.FullName,
		Owner:         repo.Owner,
		Description:   textutil.NormalizeText(repo.Description),
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
		RepoAgeDays:   ageDays,
		RepoAgeLabel:  ageLabel,
		AIScore:       repo.AIScore,
		AINotes:       textutil.NormalizeText(repo.AINotes),
		ReadmeSummary: textutil.NormalizeText(repo.ReadmeSummary),
		Category:      repo.Category,
		ProjectType:   repo.ProjectType,
		Novelty:       repo.Novelty,
		Maturity:      repo.Maturity,
		SmallRepoFit:  repo.SmallRepoFit,
		Strengths:     normalizeStringSlice(repo.Strengths),
		Weaknesses:    normalizeStringSlice(repo.Weaknesses),
		Publish:       repo.Publish,
	}
}

func normalizeStringSlice(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = textutil.NormalizeText(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func sortGithubRepos(repos []*ght.RepoData) {
	sort.Slice(repos, func(i, j int) bool {
		if repos[i].AIScore != repos[j].AIScore {
			return repos[i].AIScore > repos[j].AIScore
		}
		if repos[i].SmallRepoFit != repos[j].SmallRepoFit {
			return repos[i].SmallRepoFit > repos[j].SmallRepoFit
		}
		if repos[i].Stars != repos[j].Stars {
			return repos[i].Stars > repos[j].Stars
		}
		return repos[i].FullName < repos[j].FullName
	})
}
