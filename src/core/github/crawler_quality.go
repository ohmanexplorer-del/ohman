package github

import (
	"strings"
	"time"
	"unicode"

	"ohman/src/core/textutil"
)

func qualityGate(repo *RepoData) (bool, []string) {
	var reasons []string
	if repo == nil {
		return false, []string{"repo is nil"}
	}
	if !strings.Contains(repo.FullName, "/") {
		reasons = append(reasons, "invalid full name")
	}
	description := textutil.NormalizeText(repo.Description)
	repo.Description = description
	if len(description) < 20 {
		reasons = append(reasons, "description too short")
	}
	if looksGeneratedRepo(repo) {
		reasons = append(reasons, "looks generated or low signal")
	}
	topicsText := strings.Join(repo.Topics, " ")
	if hasRegionalScript(description) || hasRegionalScript(topicsText) || hasRegionalScript(repo.FullName) {
		reasons = append(reasons, "uses regional language script (non-international)")
	}
	if !repo.PushedAt.IsZero() && repo.PushedAt.Before(time.Now().AddDate(-1, 0, 0)) {
		reasons = append(reasons, "not updated in the last year")
	}
	if repo.AIScore > 0 && repo.AIScore < 5.5 {
		reasons = append(reasons, "AI score below threshold")
	}
	if repo.AIScore > 0 && !repo.Publish {
		reasons = append(reasons, "AI declined publish")
	}
	return len(reasons) == 0, reasons
}

func hasRegionalScript(text string) bool {
	if text == "" {
		return false
	}
	var counts [5]int
	for _, r := range text {
		switch {
		case unicode.Is(unicode.Han, r) ||
			unicode.Is(unicode.Hiragana, r) ||
			unicode.Is(unicode.Katakana, r) ||
			unicode.Is(unicode.Hangul, r):
			counts[0]++
		case unicode.Is(unicode.Arabic, r):
			counts[1]++
		case unicode.Is(unicode.Thai, r):
			counts[2]++
		case unicode.Is(unicode.Devanagari, r):
			counts[3]++
		case unicode.Is(unicode.Hebrew, r):
			counts[4]++
		}
	}
	for _, c := range counts {
		if c >= 3 {
			return true
		}
	}
	return false
}

func looksGeneratedRepo(repo *RepoData) bool {
	name := strings.ToLower(repo.FullName)
	description := strings.ToLower(repo.Description)
	markers := []string{
		"test repo",
		"demo repo",
		"my first",
		"tutorial",
		"learning",
		"practice",
		"generated",
		"template",
	}
	for _, marker := range markers {
		if strings.Contains(name, marker) || strings.Contains(description, marker) {
			return true
		}
	}
	return false
}
