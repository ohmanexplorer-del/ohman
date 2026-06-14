package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"ohman/config"
	"ohman/src/core/ai"
	storage "ohman/src/core/db"
	ght "ohman/src/core/github"
)

func main() {
	limit := flag.Int("limit", 0, "jumlah maksimal repo yang di-rescore (0 = semua)")
	dryRun := flag.Bool("dry-run", false, "tampilkan keputusan tanpa menyimpan ke DB")
	minScore := flag.Float64("min-score", 0, "hanya rescore repo dengan ai_score <= nilai ini (0 = semua)")
	delay := flag.Duration("delay", 600*time.Millisecond, "jeda antar repo untuk menghindari rate limit")
	flag.Parse()

	dsn := os.Getenv("OHMAN_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://ohman:ohman_secret@localhost:5432/ohman?sslmode=disable"
	}

	db, err := storage.New(dsn)
	if err != nil {
		log.Fatalf("database gagal: %v", err)
	}
	defer db.Close()

	cfg, err := config.LoadFromDB(db)
	if err != nil {
		log.Fatalf("config gagal: %v", err)
	}
	if cfg.GitHub.Token == "" {
		log.Fatal("github.token wajib diisi")
	}
	if cfg.AI.APIKey == "" && cfg.AI.BaseURL == "" {
		log.Fatal("ai service wajib dikonfigurasi")
	}

	ghClient := ght.NewClient(ght.ClientConfig{
		Token:           cfg.GitHub.Token,
		BotUsername:     cfg.GitHub.BotUsername,
		BotRepo:         cfg.GitHub.BotRepo,
		RateLimitBuffer: cfg.GitHub.RateLimitBuffer,
		Cache:           db,
		GraphQL:         cfg.GitHub.GraphQL,
	})

	aiSvc, err := ai.NewService(ai.AIConfig{
		BaseURL:     cfg.AI.BaseURL,
		APIKey:      cfg.AI.APIKey,
		Model:       cfg.AI.Model,
		MaxTokens:   cfg.AI.MaxTokens,
		Temperature: cfg.AI.Temperature,
		TopP:        cfg.AI.TopP,
	})
	if err != nil {
		log.Fatalf("ai service gagal: %v", err)
	}

	repos, err := db.GetReposForRescore(*limit)
	if err != nil {
		log.Fatalf("gagal memuat repo: %v", err)
	}
	log.Printf("Loaded %d repo untuk rescore", len(repos))

	ctx := context.Background()
	rescored, rejected, skipped := 0, 0, 0

	for i, dbRepo := range repos {
		if *minScore > 0 && dbRepo.AIScore > *minScore {
			skipped++
			continue
		}

		if _, safe, _ := ghClient.CheckRateLimit(ctx); !safe {
			log.Printf("Rate limit tercapai setelah %d repo, berhenti", i)
			break
		}

		log.Printf("[%d/%d] Rescoring %s (score=%.1f publish=%v)...",
			i+1, len(repos), dbRepo.FullName, dbRepo.AIScore, dbRepo.Publish)

		ghRepo := dbToGitHub(dbRepo)
		ght.EvaluateRepo(ctx, ghClient, aiSvc, ghRepo)

		publishStr := "PUBLISH"
		if !ghRepo.Publish {
			publishStr = "REJECT"
		}
		log.Printf("  => [%s] score=%.1f novelty=%.1f maturity=%.1f | %s",
			publishStr, ghRepo.AIScore, ghRepo.Novelty, ghRepo.Maturity, ghRepo.AINotes)

		if *dryRun {
			if ghRepo.Publish {
				rescored++
			} else {
				rejected++
			}
			continue
		}

		updated := githubToStorage(ghRepo, dbRepo)
		if err := db.UpsertRepo(updated); err != nil {
			log.Printf("  gagal simpan ke DB: %v", err)
			skipped++
			continue
		}

		if ghRepo.Publish {
			rescored++
		} else {
			rejected++
		}

		time.Sleep(*delay)
	}

	log.Printf("Selesai: %d publish=true | %d publish=false | %d dilewati", rescored, rejected, skipped)
	if *dryRun {
		log.Println("(dry-run: tidak ada perubahan yang disimpan)")
	}
}

func dbToGitHub(r *storage.RepoData) *ght.RepoData {
	return &ght.RepoData{
		GithubID:      r.GithubID,
		FullName:      r.FullName,
		Owner:         ownerFromFullName(r.FullName, r.Owner),
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

func githubToStorage(g *ght.RepoData, original *storage.RepoData) *storage.RepoData {
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
		VisitedAt:     original.VisitedAt,
		AIScore:       g.AIScore,
		AINotes:       g.AINotes,
		ReadmeSummary: original.ReadmeSummary,
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

func ownerFromFullName(fullName, fallback string) string {
	if idx := strings.IndexByte(fullName, '/'); idx > 0 {
		return fullName[:idx]
	}
	return fallback
}
