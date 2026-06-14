package prompts

import (
	"embed"
	"text/template"
)

//go:embed system.md navigator.md explorer.md github/*.md contributor/*.md
var Files embed.FS

type NavigatorPromptData struct {
	CurrentURL string
	TargetURL  string
}

type ExplorerPromptData struct {
	URL            string
	Title          string
	ContentPreview string
}

type GitHubExploreUserPromptData struct {
	Username    string
	Name        string
	Bio         string
	Followers   int
	Following   int
	PublicRepos int
}

type GitHubDiscoverReposPromptData struct {
	Username        string
	Bio             string
	QueueSize       int
	DiscoveredCount int
}

type GitHubEvaluateRepoPromptData struct {
	FullName        string
	Owner           string
	Description     string
	Language        string
	Stars           int
	Forks           int
	OpenIssues      int
	SizeKB          int
	Topics          []string
	License         string
	RepoAgeDays     int
	HasHomepage     bool
	Archived        bool
	ReadmePreview   string
	HasCI           bool
	HasTests        bool
	HasContributing bool
}

type GitHubSummarizeReadmePromptData struct {
	FullName      string
	Description   string
	ReadmePreview string
}

type ContributorTriageIssuePromptData struct {
	FullName   string
	Number     int
	Title      string
	Labels     string
	Comments   int
	Body       string
	SizeKB     int
	Stars      int
	Forks      int
	OpenIssues int
	Language   string
}

type ContributorGeneratePatchPromptData struct {
	FullName      string
	Number        int
	Title         string
	Labels        string
	Body          string
	Language      string
	SizeKB        int
	DefaultBranch string
	ContextText   string
}

func RenderNavigator(data NavigatorPromptData) (string, error) {
	tmpl, err := template.ParseFS(Files, "navigator.md")
	if err != nil {
		return "", err
	}
	return renderTemplate(tmpl, data)
}

func RenderExplorer(data ExplorerPromptData) (string, error) {
	tmpl, err := template.ParseFS(Files, "explorer.md")
	if err != nil {
		return "", err
	}
	return renderTemplate(tmpl, data)
}

func RenderGitHubExploreUser(data GitHubExploreUserPromptData) (string, error) {
	tmpl, err := template.ParseFS(Files, "github/explore_user.md")
	if err != nil {
		return "", err
	}
	return renderTemplate(tmpl, data)
}

func RenderGitHubDiscoverRepos(data GitHubDiscoverReposPromptData) (string, error) {
	tmpl, err := template.ParseFS(Files, "github/discover_repos.md")
	if err != nil {
		return "", err
	}
	return renderTemplate(tmpl, data)
}

func RenderGitHubEvaluateRepo(data GitHubEvaluateRepoPromptData) (string, error) {
	tmpl, err := template.ParseFS(Files, "github/evaluate_repo.md")
	if err != nil {
		return "", err
	}
	return renderTemplate(tmpl, data)
}

func RenderGitHubSummarizeReadme(data GitHubSummarizeReadmePromptData) (string, error) {
	tmpl, err := template.ParseFS(Files, "github/summarize_readme.md")
	if err != nil {
		return "", err
	}
	return renderTemplate(tmpl, data)
}

func RenderContributorTriageIssue(data ContributorTriageIssuePromptData) (string, error) {
	tmpl, err := template.ParseFS(Files, "contributor/triage_issue.md")
	if err != nil {
		return "", err
	}
	return renderTemplate(tmpl, data)
}

func RenderContributorGeneratePatch(data ContributorGeneratePatchPromptData) (string, error) {
	tmpl, err := template.ParseFS(Files, "contributor/generate_patch.md")
	if err != nil {
		return "", err
	}
	return renderTemplate(tmpl, data)
}

func SystemPrompt() (string, error) {
	b, err := Files.ReadFile("system.md")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func renderTemplate(tmpl *template.Template, data any) (string, error) {
	var buf []byte
	w := &writer{buf: &buf}
	if err := tmpl.Execute(w, data); err != nil {
		return "", err
	}
	return string(buf), nil
}

type writer struct {
	buf *[]byte
}

func (w *writer) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
