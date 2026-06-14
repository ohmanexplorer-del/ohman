package github

import "time"

type RepoData struct {
	GithubID      int64     `json:"github_id"`
	FullName      string    `json:"full_name"`
	Owner         string    `json:"owner"`
	Description   string    `json:"description"`
	Stars         int       `json:"stars"`
	Forks         int       `json:"forks"`
	Language      string    `json:"language"`
	Topics        []string  `json:"topics"`
	License       string    `json:"license"`
	DefaultBranch string    `json:"default_branch"`
	Homepage      string    `json:"homepage"`
	VisitedAt     time.Time `json:"visited_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	PushedAt      time.Time `json:"pushed_at"`
	RepoAgeDays   int       `json:"repo_age_days,omitempty"`
	RepoAgeLabel  string    `json:"repo_age_label,omitempty"`
	AIScore       float64   `json:"ai_score"`
	AINotes       string    `json:"ai_notes"`
	ReadmeSummary string    `json:"readme_summary"`
	Category      string    `json:"category"`
	ProjectType   string    `json:"project_type"`
	Novelty       float64   `json:"novelty"`
	Maturity      float64   `json:"maturity"`
	SmallRepoFit  float64   `json:"small_repo_fit"`
	Strengths     []string  `json:"strengths"`
	Weaknesses    []string  `json:"weaknesses"`
	Publish       bool      `json:"publish"`
}

type UserData struct {
	GithubID    int64     `json:"github_id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	Bio         string    `json:"bio"`
	Followers   int       `json:"followers"`
	Following   int       `json:"following"`
	PublicRepos int       `json:"public_repos"`
	VisitedAt   time.Time `json:"visited_at"`
}

type Session struct {
	ID            int64     `json:"id"`
	Strategy      string    `json:"strategy"`
	StartedAt     time.Time `json:"started_at"`
	EndedAt       time.Time `json:"ended_at"`
	ReposVisited  int       `json:"repos_visited"`
	DecisionsMade int       `json:"decisions_made"`
	AITokensUsed  int       `json:"ai_tokens_used"`
}

type Edge struct {
	FromUser     string    `json:"from_user"`
	ToUser       string    `json:"to_user"`
	Relation     string    `json:"relation"`
	DiscoveredAt time.Time `json:"discovered_at"`
}

type ExplorationConfig struct {
	Mode               string
	MinStars           int
	Languages          []string
	MaxUsersPerSession int
	ReposPerUser       int
	MaxReposPerRun     int
	SeedUsers          []string
	SearchQueries      []string
	RateLimitBuffer    int
}
