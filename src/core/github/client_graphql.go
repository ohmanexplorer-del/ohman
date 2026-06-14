package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type graphQLResponse struct {
	Data   graphQLData    `json:"data"`
	Errors []graphQLError `json:"errors"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type graphQLData struct {
	Repository graphQLRepository `json:"repository"`
}

type graphQLRepository struct {
	NameWithOwner    string             `json:"nameWithOwner"`
	StargazerCount   int                `json:"stargazerCount"`
	ForkCount        int                `json:"forkCount"`
	DiskUsage        int                `json:"diskUsage"`
	IsArchived       bool               `json:"isArchived"`
	IsDisabled       bool               `json:"isDisabled"`
	OpenIssues       graphQLTotalCount  `json:"issues"`
	DefaultBranchRef graphQLBranchRef   `json:"defaultBranchRef"`
	PrimaryLanguage  graphQLNamePayload `json:"primaryLanguage"`
}

type graphQLTotalCount struct {
	TotalCount int `json:"totalCount"`
}

type graphQLBranchRef struct {
	Name string `json:"name"`
}

type graphQLNamePayload struct {
	Name string `json:"name"`
}

func (c *Client) GetRepoComplexityGraphQL(ctx context.Context, owner, repoName string) (*RepoComplexity, error) {
	query := `query($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) {
    nameWithOwner
    stargazerCount
    forkCount
    diskUsage
    isArchived
    isDisabled
    issues(states: OPEN) {
      totalCount
    }
    defaultBranchRef {
      name
    }
    primaryLanguage {
      name
    }
  }
}`

	payload, err := json.Marshal(graphQLRequest{
		Query: query,
		Variables: map[string]any{
			"owner": owner,
			"name":  repoName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode graphql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.github.com/graphql", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to build graphql request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call graphql api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("graphql api returned status %d", resp.StatusCode)
	}

	var out graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode graphql response: %w", err)
	}
	if len(out.Errors) > 0 {
		return nil, fmt.Errorf("graphql api error: %s", out.Errors[0].Message)
	}

	repo := out.Data.Repository
	if repo.NameWithOwner == "" {
		return nil, fmt.Errorf("graphql repository %s/%s not found", owner, repoName)
	}

	return &RepoComplexity{
		FullName:      repo.NameWithOwner,
		Stars:         repo.StargazerCount,
		Forks:         repo.ForkCount,
		OpenIssues:    repo.OpenIssues.TotalCount,
		SizeKB:        repo.DiskUsage,
		DefaultBranch: repo.DefaultBranchRef.Name,
		Archived:      repo.IsArchived,
		Disabled:      repo.IsDisabled,
		Language:      repo.PrimaryLanguage.Name,
	}, nil
}
