package github

import (
	"strings"
	"testing"
	"time"
)

func TestQualityGateAcceptsHealthyRepo(t *testing.T) {
	repo := &RepoData{
		FullName:    "owner/project",
		Description: "Focused developer tool with active maintenance.",
		PushedAt:    time.Now(),
		AIScore:     7.5,
		Publish:     true,
	}

	ok, reasons := qualityGate(repo)
	if !ok {
		t.Fatalf("expected repo to pass, got reasons: %v", reasons)
	}
}

func TestQualityGateRejectsLowSignalRepo(t *testing.T) {
	repo := &RepoData{
		FullName:    "owner/tutorial-project",
		Description: "demo",
		PushedAt:    time.Now().AddDate(-2, 0, 0),
		AIScore:     4,
		Publish:     false,
	}

	ok, reasons := qualityGate(repo)
	if ok {
		t.Fatal("expected repo to be rejected")
	}
	joined := strings.Join(reasons, " ")
	for _, expected := range []string{"description too short", "low signal", "last year", "AI score"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected reason containing %q, got %v", expected, reasons)
		}
	}
}
