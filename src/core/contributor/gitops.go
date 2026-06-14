package contributor

import (
	"context"
	"fmt"
	ght "ohman/src/core/github"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func applyDiff(ctx context.Context, worktree, diff string) error {
	path := filepath.Join(worktree, ".ohman.patch")
	if err := os.WriteFile(path, []byte(diff), 0o600); err != nil {
		return fmt.Errorf("failed to write patch file: %w", err)
	}
	defer os.Remove(path)

	cmd := exec.CommandContext(ctx, "git", "apply", "--whitespace=fix", path)
	cmd.Dir = worktree
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply AI patch: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func runValidation(ctx context.Context, worktree string) error {
	commands := validationCommands(worktree)
	for _, args := range commands {
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = worktree
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("validation failed (%s): %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
		}
	}
	return nil
}

func validationCommands(worktree string) [][]string {
	switch {
	case fileExists(worktree, "go.mod"):
		return [][]string{{"go", "test", "./..."}}
	case fileExists(worktree, "package.json"):
		return [][]string{{"npm", "test", "--", "--watch=false"}}
	case fileExists(worktree, "pyproject.toml"):
		return [][]string{{"python", "-m", "pytest"}}
	case fileExists(worktree, "Cargo.toml"):
		return [][]string{{"cargo", "test"}}
	default:
		return nil
	}
}

func fileExists(worktree, name string) bool {
	_, err := os.Stat(filepath.Join(worktree, name))
	return err == nil
}

func git(ctx context.Context, worktree string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", worktree}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func gitClean(ctx context.Context, worktree string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", worktree, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}
	return strings.TrimSpace(string(output)) == "", nil
}

func pushBranch(ctx context.Context, worktree, token, owner, repo, branch string) error {
	remote := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	header := "http.https://github.com/.extraheader=AUTHORIZATION: bearer " + token
	cmd := exec.CommandContext(ctx, "git", "-C", worktree, "-c", header, "push", "--force", remote, "HEAD:"+branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w: %s", err, sanitizeToken(string(output), token))
	}
	return nil
}

func safeBranchName(name string, issue *ght.IssueCandidate) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		name = fmt.Sprintf("ohman/fix-%d-%d", issue.Number, time.Now().Unix())
	}
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.Trim(name, "/-")
	if !strings.HasPrefix(name, "ohman/") {
		name = "ohman/" + name
	}
	return name
}

func sanitizeToken(output, token string) string {
	output = strings.TrimSpace(output)
	if token == "" {
		return output
	}
	return strings.ReplaceAll(output, token, "[redacted]")
}
