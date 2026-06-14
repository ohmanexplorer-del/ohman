package github

import (
	"context"
	"net/http"
	"ohman/src/core/textutil"
	"strings"

	gh "github.com/google/go-github/v58/github"
)

func (c *Client) defaultBranch(ctx context.Context, owner, repoName string) (string, bool) {
	repo, _, err := c.client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return "", false
	}
	branch := repo.GetDefaultBranch()
	if branch == "" {
		return "", false
	}
	_, resp, err := c.client.Git.GetRef(ctx, owner, repoName, "refs/heads/"+branch)
	if err != nil && resp != nil && resp.StatusCode == http.StatusConflict {
		return "", false
	}
	if err != nil {
		return "", false
	}
	return branch, true
}

func issueToCandidate(issue *gh.Issue) (*IssueCandidate, bool) {
	owner, repo, ok := parseIssueRepo(issue.GetHTMLURL())
	if !ok {
		return nil, false
	}

	labels := make([]string, 0, len(issue.Labels))
	for _, label := range issue.Labels {
		labels = append(labels, label.GetName())
	}

	return &IssueCandidate{
		Owner:     owner,
		Repo:      repo,
		FullName:  owner + "/" + repo,
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		Body:      issue.GetBody(),
		URL:       issue.GetHTMLURL(),
		Labels:    labels,
		Comments:  issue.GetComments(),
		CreatedAt: issue.GetCreatedAt().Time,
		UpdatedAt: issue.GetUpdatedAt().Time,
	}, true
}

func parseIssueRepo(url string) (string, string, bool) {
	parts := strings.Split(url, "/")
	for i := 0; i+4 < len(parts); i++ {
		if parts[i] == "github.com" && parts[i+3] == "issues" {
			return parts[i+1], parts[i+2], true
		}
	}
	return "", "", false
}

func repoToData(r *gh.Repository) *RepoData {
	license := ""
	if r.GetLicense() != nil {
		license = r.GetLicense().GetName()
	}

	return &RepoData{
		GithubID:      r.GetID(),
		FullName:      r.GetFullName(),
		Owner:         r.GetOwner().GetLogin(),
		Description:   textutil.NormalizeText(r.GetDescription()),
		Stars:         r.GetStargazersCount(),
		Forks:         r.GetForksCount(),
		OpenIssues:    r.GetOpenIssuesCount(),
		SizeKB:        r.GetSize(),
		Language:      r.GetLanguage(),
		Topics:        r.Topics,
		License:       license,
		DefaultBranch: r.GetDefaultBranch(),
		Homepage:      r.GetHomepage(),
		CreatedAt:     r.GetCreatedAt().Time,
		UpdatedAt:     r.GetUpdatedAt().Time,
		PushedAt:      r.GetPushedAt().Time,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func firstNonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
