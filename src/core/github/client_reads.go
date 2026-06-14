package github

import (
	"context"
	"fmt"
	"time"

	gh "github.com/google/go-github/v58/github"
)

func (c *Client) GetUser(ctx context.Context, username string) (*UserData, error) {
	user, _, err := c.client.Users.Get(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", username, err)
	}
	return &UserData{
		GithubID:    user.GetID(),
		Username:    user.GetLogin(),
		Name:        user.GetName(),
		Bio:         user.GetBio(),
		Followers:   user.GetFollowers(),
		Following:   user.GetFollowing(),
		PublicRepos: user.GetPublicRepos(),
		VisitedAt:   time.Now(),
	}, nil
}

func (c *Client) GetUserRepos(ctx context.Context, username string, opts *RepoListOptions) ([]*RepoData, error) {
	if opts == nil {
		opts = &RepoListOptions{}
	}

	listOpts := &gh.RepositoryListByUserOptions{
		ListOptions: gh.ListOptions{PerPage: opts.PerPage},
		Type:        "owner",
		Sort:        opts.Sort,
		Direction:   opts.Direction,
	}

	var allRepos []*RepoData
	for {
		repos, resp, err := c.client.Repositories.ListByUser(ctx, username, listOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repos for %s: %w", username, err)
		}
		for _, repo := range repos {
			allRepos = append(allRepos, repoToData(repo))
			if len(allRepos) >= opts.MaxRepos {
				return allRepos, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}
	return allRepos, nil
}

func (c *Client) GetRepo(ctx context.Context, owner, repo string) (*RepoData, error) {
	r, _, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo %s/%s: %w", owner, repo, err)
	}
	return repoToData(r), nil
}

func (c *Client) GetReadme(ctx context.Context, owner, repo string) (string, error) {
	readme, _, err := c.client.Repositories.GetReadme(ctx, owner, repo, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get README for %s/%s: %w", owner, repo, err)
	}
	content, err := readme.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to decode README for %s/%s: %w", owner, repo, err)
	}
	return content, nil
}

func (c *Client) GetRepoTopics(ctx context.Context, owner, repo string) ([]string, error) {
	topics, _, err := c.client.Repositories.ListAllTopics(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get topics for %s/%s: %w", owner, repo, err)
	}
	return topics, nil
}

func (c *Client) GetFollowers(ctx context.Context, username string, maxResults int) ([]string, error) {
	opts := &gh.ListOptions{PerPage: 100}
	var users []string
	for {
		followers, resp, err := c.client.Users.ListFollowers(ctx, username, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to get followers for %s: %w", username, err)
		}
		for _, u := range followers {
			users = append(users, u.GetLogin())
			if len(users) >= maxResults {
				return users, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return users, nil
}

func (c *Client) GetFollowing(ctx context.Context, username string, maxResults int) ([]string, error) {
	opts := &gh.ListOptions{PerPage: 100}
	var users []string
	for {
		following, resp, err := c.client.Users.ListFollowing(ctx, username, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to get following for %s: %w", username, err)
		}
		for _, u := range following {
			users = append(users, u.GetLogin())
			if len(users) >= maxResults {
				return users, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return users, nil
}

func (c *Client) GetRepoComplexity(ctx context.Context, owner, repoName string) (*RepoComplexity, error) {
	if c.config.GraphQL {
		complexity, err := c.GetRepoComplexityGraphQL(ctx, owner, repoName)
		if err == nil {
			return complexity, nil
		}
	}

	repo, _, err := c.client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo complexity for %s/%s: %w", owner, repoName, err)
	}
	return &RepoComplexity{
		FullName:      repo.GetFullName(),
		Stars:         repo.GetStargazersCount(),
		Forks:         repo.GetForksCount(),
		OpenIssues:    repo.GetOpenIssuesCount(),
		SizeKB:        repo.GetSize(),
		DefaultBranch: repo.GetDefaultBranch(),
		Archived:      repo.GetArchived(),
		Disabled:      repo.GetDisabled(),
		Language:      repo.GetLanguage(),
	}, nil
}

func (c *Client) GetRepoContents(ctx context.Context, owner, repo, path string) ([]*RepoContent, error) {
	fileContent, dirContents, _, err := c.client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get contents for %s/%s:%s: %w", owner, repo, path, err)
	}

	if fileContent != nil {
		content, err := fileContent.GetContent()
		if err != nil {
			return nil, fmt.Errorf("failed to decode file content for %s/%s:%s: %w", owner, repo, path, err)
		}
		return []*RepoContent{{
			Path:    fileContent.GetPath(),
			Type:    fileContent.GetType(),
			Content: content,
			SHA:     fileContent.GetSHA(),
			URL:     fileContent.GetHTMLURL(),
		}}, nil
	}

	if len(dirContents) > 0 {
		contents := make([]*RepoContent, 0, len(dirContents))
		for _, item := range dirContents {
			contents = append(contents, &RepoContent{
				Path: item.GetPath(),
				Type: item.GetType(),
				SHA:  item.GetSHA(),
				URL:  item.GetHTMLURL(),
			})
		}
		return contents, nil
	}

	return nil, fmt.Errorf("no content found at %s/%s:%s", owner, repo, path)
}
