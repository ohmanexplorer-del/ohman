package github

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	gh "github.com/google/go-github/v58/github"
)

func (c *Client) EnsureBotRepoReady(ctx context.Context) error {
	_, _, resp, err := c.client.Repositories.GetContents(ctx, c.username, c.repoName, "", nil)
	if err == nil {
		return nil
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to inspect bot repo contents: %w", err)
	}

	content := fmt.Sprintf(`# Ohman Explorer

Autonomous discovery log managed by %s.
`, c.username)

	if err := c.CreateOrUpdateFile(ctx, "README.md", "Initialize Ohman Explorer repository", content); err != nil {
		return fmt.Errorf("failed to initialize bot repo: %w", err)
	}
	return nil
}

func (c *Client) StarRepo(ctx context.Context, owner, repo string) error {
	_, err := c.client.Activity.Star(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to star %s/%s: %w", owner, repo, err)
	}
	return nil
}

func (c *Client) UnstarRepo(ctx context.Context, owner, repo string) error {
	_, err := c.client.Activity.Unstar(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to unstar %s/%s: %w", owner, repo, err)
	}
	return nil
}

func (c *Client) FollowUser(ctx context.Context, username string) error {
	_, err := c.client.Users.Follow(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to follow user %s: %w", username, err)
	}
	return nil
}

func (c *Client) CreateIssue(ctx context.Context, owner, repo, title, body string) (*IssueData, error) {
	issue, _, err := c.client.Issues.Create(ctx, owner, repo, &gh.IssueRequest{Title: gh.String(title), Body: gh.String(body)})
	if err != nil {
		return nil, fmt.Errorf("failed to create issue in %s/%s: %w", owner, repo, err)
	}
	return &IssueData{
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		Body:      issue.GetBody(),
		State:     issue.GetState(),
		URL:       issue.GetHTMLURL(),
		CreatedAt: issue.GetCreatedAt().Time,
		UpdatedAt: issue.GetUpdatedAt().Time,
	}, nil
}

func (c *Client) CreateComment(ctx context.Context, owner, repo string, issueNumber int, body string) (*CommentData, error) {
	comment, _, err := c.client.Issues.CreateComment(ctx, owner, repo, issueNumber, &gh.IssueComment{Body: gh.String(body)})
	if err != nil {
		return nil, fmt.Errorf("failed to create comment in %s/%s#%d: %w", owner, repo, issueNumber, err)
	}
	return &CommentData{
		ID:        comment.GetID(),
		Body:      comment.GetBody(),
		User:      comment.GetUser().GetLogin(),
		CreatedAt: comment.GetCreatedAt().Time,
		UpdatedAt: comment.GetUpdatedAt().Time,
	}, nil
}

func (c *Client) EnsureFork(ctx context.Context, owner, repo string) error {
	if _, _, err := c.client.Repositories.Get(ctx, c.username, repo); err == nil {
		return nil
	}

	_, _, err := c.client.Repositories.CreateFork(ctx, owner, repo, nil)
	if err != nil {
		return fmt.Errorf("failed to create fork for %s/%s: %w", owner, repo, err)
	}

	for i := 0; i < 12; i++ {
		if _, _, err := c.client.Repositories.Get(ctx, c.username, repo); err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("fork %s/%s was not ready in time", c.username, repo)
}

func (c *Client) CreatePullRequest(ctx context.Context, owner, repo, headBranch, baseBranch, title, body string) (*PullRequestData, error) {
	pr, _, err := c.client.PullRequests.Create(ctx, owner, repo, &gh.NewPullRequest{
		Title: new(title),
		Head:  new(c.username + ":" + headBranch),
		Base:  new(baseBranch),
		Body:  new(body),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request for %s/%s: %w", owner, repo, err)
	}
	return &PullRequestData{
		Number: pr.GetNumber(),
		Title:  pr.GetTitle(),
		URL:    pr.GetHTMLURL(),
		State:  pr.GetState(),
	}, nil
}

func (c *Client) CreateFile(ctx context.Context, path, message, content string) error {
	return c.CreateOrUpdateFile(ctx, path, message, content)
}

func (c *Client) CreateOrUpdateFile(ctx context.Context, path, message, content string) error {
	file, _, resp, err := c.client.Repositories.GetContents(ctx, c.username, c.repoName, path, nil)
	if err != nil && (resp == nil || resp.StatusCode != http.StatusNotFound) {
		return fmt.Errorf("failed to inspect file %s: %w", path, err)
	}

	var sha string
	if file != nil {
		sha = file.GetSHA()
	}

	opts := &gh.RepositoryContentFileOptions{
		Message: new(message),
		Content: []byte(content),
	}

	if branch, ok := c.defaultBranch(ctx, c.username, c.repoName); ok {
		opts.Branch = new(branch)
	}

	if sha != "" {
		opts.SHA = new(sha)
		_, _, err = c.client.Repositories.UpdateFile(ctx, c.username, c.repoName, path, opts)
	} else {
		_, _, err = c.client.Repositories.CreateFile(ctx, c.username, c.repoName, path, opts)
	}
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	log.Printf("Published file: %s", path)
	return nil
}

func (c *Client) DeleteFile(ctx context.Context, path, message string) error {
	file, _, _, err := c.client.Repositories.GetContents(ctx, c.username, c.repoName, path, nil)
	if err != nil {
		return fmt.Errorf("failed to inspect file %s: %w", path, err)
	}
	if file == nil || file.GetSHA() == "" {
		return fmt.Errorf("file %s is not a regular file", path)
	}

	opts := &gh.RepositoryContentFileOptions{
		Message: new(message),
		SHA:     new(file.GetSHA()),
	}
	if branch, ok := c.defaultBranch(ctx, c.username, c.repoName); ok {
		opts.Branch = new(branch)
	}

	if _, _, err := c.client.Repositories.DeleteFile(ctx, c.username, c.repoName, path, opts); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	log.Printf("Deleted file: %s", path)
	return nil
}
