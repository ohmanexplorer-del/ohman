package publisher

import (
	"strings"
	"testing"
	"time"

	ght "ohman/src/core/github"
)

func TestRepoAge(t *testing.T) {
	now := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	days, label := repoAge(now.AddDate(0, -2, 0), now)
	if days < 59 || days > 62 {
		t.Fatalf("unexpected days: %d", days)
	}
	if label != "2 months" {
		t.Fatalf("unexpected label: %s", label)
	}
}

func TestRepoAssessmentContextForYoungRepo(t *testing.T) {
	repo := &ght.RepoData{CreatedAt: time.Now().UTC().AddDate(0, 0, -10)}
	got := repoAssessmentContext(repo)
	if !strings.Contains(got, "very young repo") {
		t.Fatalf("unexpected context: %s", got)
	}
}

func TestWriteRepoCardIncludesAgeContext(t *testing.T) {
	var sb strings.Builder
	writeRepoCard(&sb, 1, &ght.RepoData{
		FullName:    "owner/project",
		Description: "Useful project with a clear purpose.",
		Category:    "dev-tools",
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		PushedAt:    time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		AIScore:     7.5,
	})
	out := sb.String()
	for _, expected := range []string{"- Created: 2026-01-01", "- Age:", "- Last pushed: 2026-06-01", "- Assessment context:"} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, out)
		}
	}
}
