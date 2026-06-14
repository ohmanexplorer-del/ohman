package github

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ohman/src/core/taxonomy"
)

func trimToLength(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func parseRepoEvaluation(response string) (repoEvaluation, error) {
	response = strings.TrimSpace(response)
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start < 0 || end < start {
		return repoEvaluation{}, fmt.Errorf("json object not found")
	}

	var evaluation repoEvaluation
	if err := json.Unmarshal([]byte(response[start:end+1]), &evaluation); err != nil {
		return repoEvaluation{}, err
	}
	return evaluation, nil
}

func applyFallbackEvaluation(repo *RepoData) {
	repo.Category = normalizeCategory(repo.Language)
	if repo.Category == "uncategorized" {
		repo.Category = "general"
	}
	repo.ProjectType = "unknown"
	repo.AIScore = fallbackScore(repo)
	repo.Novelty = 5
	repo.Maturity = fallbackMaturity(repo)
	repo.SmallRepoFit = fallbackSmallRepoFit(repo)
	repo.AINotes = "Selected from repository metadata."
	repo.Strengths = []string{"Interesting repository metadata"}
	repo.Weaknesses = []string{"Needs deeper review"}
	repo.Publish = true
}

func fallbackScore(repo *RepoData) float64 {
	score := 5.0
	if repo.Stars > 10000 {
		score += 2
	} else if repo.Stars > 1000 {
		score += 1.5
	} else if repo.Stars > 100 {
		score += 1
	} else if repo.Stars < 100 && repo.PushedAt.After(time.Now().AddDate(0, -6, 0)) {
		score += 1
	}
	return clampScore(score)
}

func fallbackMaturity(repo *RepoData) float64 {
	score := 4.0
	if repo.Stars > 1000 {
		score += 2
	}
	if repo.DefaultBranch != "" {
		score += 1
	}
	if repo.PushedAt.After(time.Now().AddDate(0, -6, 0)) {
		score += 1
	}
	return clampScore(score)
}

func fallbackSmallRepoFit(repo *RepoData) float64 {
	if repo.Stars < 100 {
		return 8
	}
	if repo.Stars < 1000 {
		return 6
	}
	return 3
}

func clampScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 10 {
		return 10
	}
	return score
}

func normalizeCategory(category string) string {
	return taxonomy.NormalizeCategory(category)
}

func trimStringSlice(values []string, limit int) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out = append(out, trimToLength(value, 160))
		if len(out) >= limit {
			break
		}
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
