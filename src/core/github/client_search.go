package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v58/github"
)

func (c *Client) SearchRepos(ctx context.Context, query string, opts *RepoListOptions) ([]*RepoData, error) {
	if opts == nil {
		opts = &RepoListOptions{}
	}

	searchOpts := &gh.SearchOptions{
		ListOptions: gh.ListOptions{PerPage: opts.PerPage},
		Sort:        firstNonEmpty(opts.Sort, "stars"),
		Order:       firstNonEmpty(opts.Direction, "desc"),
	}

	if len(opts.Languages) > 0 {
		for _, lang := range opts.Languages {
			query += fmt.Sprintf(" language:%s", lang)
		}
	}

	var allRepos []*RepoData
	for {
		result, resp, err := c.client.Search.Repositories(ctx, query, searchOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to search repos: %w", err)
		}
		for _, repo := range result.Repositories {
			allRepos = append(allRepos, repoToData(repo))
			if len(allRepos) >= opts.MaxRepos {
				return allRepos, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		searchOpts.Page = resp.NextPage
	}
	return allRepos, nil
}

func (c *Client) SearchIssueCandidates(ctx context.Context, label string, maxResults int) ([]*IssueCandidate, error) {
	if maxResults <= 0 {
		return nil, nil
	}

	query := fmt.Sprintf(`is:issue is:open archived:false label:"%s"`, label)
	opts := &gh.SearchOptions{
		Sort:        "updated",
		Order:       "desc",
		ListOptions: gh.ListOptions{PerPage: min(maxResults, 100)},
	}

	var candidates []*IssueCandidate
	for {
		result, resp, err := c.client.Search.Issues(ctx, query, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to search issue candidates: %w", err)
		}
		for _, issue := range result.Issues {
			candidate, ok := issueToCandidate(issue)
			if !ok {
				continue
			}
			candidates = append(candidates, candidate)
			if len(candidates) >= maxResults {
				return candidates, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return candidates, nil
}
