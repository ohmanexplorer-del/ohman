package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	storage "ohman/src/core/db"
	ght "ohman/src/core/github"
)

func (a *App) runContributionOnce() error {
	if a.Contributor == nil {
		return nil
	}

	a.notifyActivity("Ohman mulai mencari issue kandidat untuk contribution.")
	reports, err := a.Contributor.Run(context.Background())
	if err != nil {
		a.notifyActivity("Contribution scan gagal: " + err.Error())
		return fmt.Errorf("contributor scan failed: %w", err)
	}
	accepted := 0
	for _, report := range reports {
		if report.Issue == nil {
			continue
		}
		status := "skipped"
		if report.Accepted {
			status = "accepted"
			accepted++
		}
		log.Printf("Contribution candidate %s#%d %s: %s",
			report.Issue.FullName,
			report.Issue.Number,
			status,
			strings.Join(report.Reasons, "; "),
		)
	}
	a.notifyActivity(fmt.Sprintf("Contribution scan selesai: %d kandidat dicek, %d diterima.", len(reports), accepted))
	return nil
}

func (a *App) githubReposPerRunLimit() int {
	if a.Config == nil || a.Config.Exploration.MaxReposPerRun <= 0 {
		return 10
	}
	return a.Config.Exploration.MaxReposPerRun
}

func (a *App) RunGitHubDiscoveryNow(limit int) error {
	if !a.botRunning {
		if err := a.StartBot(); err != nil {
			return err
		}
		return nil
	}
	if a.GitHubCrawl == nil {
		return fmt.Errorf("GitHub crawler is not initialized")
	}
	if limit <= 0 {
		limit = a.githubReposPerRunLimit()
	}
	a.lastGitHubDiscoveryAt = nil
	return a.runGitHubDiscoveryOnce(limit)
}

func (a *App) RescoreNow(limit int) error {
	if a.GitHub == nil {
		a.initGitHub()
	}
	if a.AI == nil {
		if err := a.initAI(); err != nil {
			return fmt.Errorf("gagal inisialisasi AI: %w", err)
		}
	}
	if a.GitHub == nil {
		return fmt.Errorf("GitHub belum dikonfigurasi (github.token kosong?)")
	}
	if a.AI == nil {
		return fmt.Errorf("AI belum dikonfigurasi")
	}
	if limit <= 0 {
		limit = 20
	}
	go func() {
		a.notifyActivity(fmt.Sprintf("Rescore dimulai: akan mengevaluasi ulang %d repo dengan prompt terbaru.", limit))
		repos, err := a.DB.GetReposForRescore(limit)
		if err != nil {
			a.notifyActivity("Rescore gagal memuat repo: " + err.Error())
			return
		}

		ctx := context.Background()
		rescored, rejected, skipped := 0, 0, 0
		for i, dbRepo := range repos {
			if _, safe, _ := a.GitHub.CheckRateLimit(ctx); !safe {
				a.notifyActivity(fmt.Sprintf("Rescore berhenti di repo ke-%d karena rate limit GitHub.", i+1))
				break
			}

			ghRepo := dbRepoToGitHub(dbRepo)
			ght.EvaluateRepo(ctx, a.GitHub, a.AI, ghRepo)

			updated := githubRepoToStorage(ghRepo, dbRepo)
			if err := a.DB.UpsertRepo(updated); err != nil {
				log.Printf("Rescore: gagal simpan %s: %v", dbRepo.FullName, err)
				skipped++
				continue
			}
			if ghRepo.Publish {
				rescored++
			} else {
				rejected++
			}
		}
		a.notifyActivity(fmt.Sprintf("Rescore selesai: %d publish=true | %d publish=false | %d error.", rescored, rejected, skipped))
	}()
	return nil
}

func (a *App) PublishLibraryNow() error {
	if a.Publisher == nil {
		a.initGitHub()
		a.initPublisher()
	}
	if a.Publisher == nil {
		return fmt.Errorf("publisher is not initialized")
	}
	return a.Publisher.PublishLibrary()
}

func (a *App) runGitHubDiscoveryOnce(limit int) error {
	if !a.githubDiscoveryMu.TryLock() {
		log.Println("GitHub discovery already running, skipping this tick")
		a.notifyActivity("GitHub discovery masih berjalan, tick ini dilewati.")
		return nil
	}
	defer a.githubDiscoveryMu.Unlock()

	now := time.Now()
	if a.lastGitHubDiscoveryAt != nil && now.Sub(*a.lastGitHubDiscoveryAt) < 10*time.Minute {
		log.Println("GitHub discovery ran recently, skipping this tick")
		return nil
	}
	a.lastGitHubDiscoveryAt = &now

	if limit > 0 {
		log.Printf("Starting GitHub autonomous discovery for up to %d repos", limit)
		a.notifyActivity(fmt.Sprintf("Ohman mulai mencari repo GitHub menarik. Target run: sampai %d repo.", limit))
	} else {
		log.Println("Starting GitHub autonomous discovery")
		a.notifyActivity("Ohman mulai mencari repo GitHub menarik.")
	}

	sessionID, err := a.DB.CreateSession(a.Config.Exploration.Mode)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		a.notifyActivity("Gagal membuat session discovery: " + err.Error())
	}

	repos, err := a.GitHubCrawl.RunLimit(sessionID, limit)
	if err != nil {
		a.notifyActivity("GitHub crawl gagal: " + err.Error())
		return fmt.Errorf("github crawl failed: %w", err)
	}
	if limit > 0 && len(repos) > limit {
		repos = repos[:limit]
	}
	log.Printf("GitHub crawler selected %d repos", len(repos))
	publishableRepos := publishableGitHubRepos(repos)
	log.Printf("GitHub crawler approved %d repos for publishing", len(publishableRepos))
	a.notifyActivity(fmt.Sprintf("Discovery selesai memilih %d repo, %d layak publish.", len(repos), len(publishableRepos)))

	followedOwners := map[string]bool{}
	for _, repo := range publishableRepos {
		owner := repo.Owner
		if owner == "" {
			var ok bool
			owner, _, ok = splitRepoFullName(repo.FullName)
			if !ok {
				log.Printf("Skipping follow for invalid repo name: %s", repo.FullName)
				continue
			}
		}
		if followedOwners[owner] {
			continue
		}
		if err := a.GitHub.FollowUser(context.Background(), owner); err != nil {
			log.Printf("Failed to follow owner %s for repo %s: %v", owner, repo.FullName, err)
			continue
		}
		followedOwners[owner] = true
		log.Printf("Followed owner: %s", owner)
	}
	if len(followedOwners) > 0 {
		a.notifyActivity(fmt.Sprintf("Ohman follow %d owner repo baru.", len(followedOwners)))
	}

	saved, err := a.Collector.Collect(repos)
	if err != nil {
		log.Printf("Collector error: %v", err)
		a.notifyActivity("Collector error: " + err.Error())
	}
	log.Printf("Saved %d repos to database", saved)

	if len(repos) > 0 {
		if err := a.Publisher.PublishFindings(publishableRepos); err != nil {
			a.notifyActivity("Publisher gagal: " + err.Error())
			return fmt.Errorf("publisher failed: %w", err)
		}
		a.notifyActivity(fmt.Sprintf("Publisher selesai update README/findings dengan %d repo publishable.", len(publishableRepos)))
		a.broadcastGitHubFindings(publishableRepos)
	}

	if sessionID > 0 {
		_ = a.DB.UpdateSession(sessionID, saved, a.GitHubCrawl.DecisionsMade(), 0)
	}

	stats, _ := a.DB.Stats()
	log.Printf("Database stats: repos=%d, users=%d, sessions=%d",
		stats["repos"], stats["users"], stats["sessions"])
	a.notifyActivity(fmt.Sprintf("Stats DB: repos=%d users=%d sessions=%d.", stats["repos"], stats["users"], stats["sessions"]))
	return nil
}

func dbRepoToGitHub(r *storage.RepoData) *ght.RepoData {
	return &ght.RepoData{
		GithubID:      r.GithubID,
		FullName:      r.FullName,
		Owner:         r.Owner,
		Description:   r.Description,
		Stars:         r.Stars,
		Forks:         r.Forks,
		Language:      r.Language,
		Topics:        r.Topics,
		License:       r.License,
		DefaultBranch: r.DefaultBranch,
		Homepage:      r.Homepage,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		PushedAt:      r.PushedAt,
		VisitedAt:     r.VisitedAt,
		AIScore:       r.AIScore,
		AINotes:       r.AINotes,
		ReadmeSummary: r.ReadmeSummary,
		Category:      r.Category,
		ProjectType:   r.ProjectType,
		Novelty:       r.Novelty,
		Maturity:      r.Maturity,
		SmallRepoFit:  r.SmallRepoFit,
		Strengths:     r.Strengths,
		Weaknesses:    r.Weaknesses,
		Publish:       r.Publish,
	}
}

func githubRepoToStorage(g *ght.RepoData, orig *storage.RepoData) *storage.RepoData {
	return &storage.RepoData{
		GithubID:      g.GithubID,
		FullName:      g.FullName,
		Owner:         g.Owner,
		Description:   g.Description,
		Stars:         g.Stars,
		Forks:         g.Forks,
		Language:      g.Language,
		Topics:        g.Topics,
		License:       g.License,
		DefaultBranch: g.DefaultBranch,
		Homepage:      g.Homepage,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
		PushedAt:      g.PushedAt,
		VisitedAt:     orig.VisitedAt,
		AIScore:       g.AIScore,
		AINotes:       g.AINotes,
		ReadmeSummary: orig.ReadmeSummary,
		Category:      g.Category,
		ProjectType:   g.ProjectType,
		Novelty:       g.Novelty,
		Maturity:      g.Maturity,
		SmallRepoFit:  g.SmallRepoFit,
		Strengths:     g.Strengths,
		Weaknesses:    g.Weaknesses,
		Publish:       g.Publish,
	}
}
