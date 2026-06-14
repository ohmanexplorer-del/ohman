package contributor

import (
	"context"
	ght "ohman/src/core/github"
)

type Config struct {
	Enabled         bool
	AutoPR          bool
	Labels          []string
	MaxIssuesPerRun int
	MaxRepoSizeKB   int
	MaxRepoStars    int
	MaxRepoForks    int
	WorkDir         string
}

type Service struct {
	client *ght.Client
	ai     AIService
	config Config
}

type AIService interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

type TriageDecision struct {
	CanFix     bool     `json:"can_fix"`
	Confidence float64  `json:"confidence"`
	Reason     string   `json:"reason"`
	Risks      []string `json:"risks"`
}

type PatchDecision struct {
	CanPatch      bool   `json:"can_patch"`
	Summary       string `json:"summary"`
	BranchName    string `json:"branch_name"`
	CommitMessage string `json:"commit_message"`
	PRTitle       string `json:"pr_title"`
	PRBody        string `json:"pr_body"`
	Diff          string `json:"diff"`
}

type CandidateReport struct {
	Issue       *ght.IssueCandidate
	Complexity  *ght.RepoComplexity
	Triage      *TriageDecision
	Patch       *PatchDecision
	PullRequest *ght.PullRequestData
	Accepted    bool
	Reasons     []string
	Worktree    string
	Cleaned     bool
}
