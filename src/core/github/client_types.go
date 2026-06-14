package github

import "time"

type HTTPCache interface {
	GetHTTPCache(key string) (etag string, body []byte, status int, found bool, err error)
	SetHTTPCache(key, etag string, body []byte, status int) error
}

type RepoContent struct {
	Path    string
	Type    string
	Content string
	SHA     string
	URL     string
}

type IssueData struct {
	Number    int
	Title     string
	Body      string
	State     string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type IssueCandidate struct {
	Owner     string
	Repo      string
	FullName  string
	Number    int
	Title     string
	Body      string
	URL       string
	Labels    []string
	Comments  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RepoComplexity struct {
	FullName      string
	Stars         int
	Forks         int
	OpenIssues    int
	SizeKB        int
	DefaultBranch string
	Archived      bool
	Disabled      bool
	Language      string
}

type CommentData struct {
	ID        int64
	Body      string
	User      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PullRequestData struct {
	Number int
	Title  string
	URL    string
	State  string
}

type ClientConfig struct {
	Token           string
	BotUsername     string
	BotRepo         string
	RateLimitBuffer int
	Cache           HTTPCache
	GraphQL         bool
}

type RepoListOptions struct {
	PerPage   int
	MaxRepos  int
	Sort      string
	Direction string
	Languages []string
}
