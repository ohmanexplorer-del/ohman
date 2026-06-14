package contributor

import (
	"context"
	"fmt"
	ght "ohman/src/core/github"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (s *Service) clone(ctx context.Context, issue *ght.IssueCandidate) (string, error) {
	if err := os.MkdirAll(s.config.WorkDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create work dir: %w", err)
	}

	worktree := filepath.Join(s.config.WorkDir, safePath(issue.FullName))
	if _, err := os.Stat(worktree); err == nil {
		return worktree, nil
	}

	cloneURL := fmt.Sprintf("https://github.com/%s.git", issue.FullName)
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", cloneURL, worktree)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to clone %s: %w: %s", issue.FullName, err, strings.TrimSpace(string(output)))
	}
	return worktree, nil
}

func repoContext(worktree string) string {
	var sb strings.Builder

	if output, err := exec.Command("git", "-C", worktree, "ls-files").Output(); err == nil {
		files := strings.Split(strings.TrimSpace(string(output)), "\n")
		sb.WriteString("Files:\n")
		for i, file := range files {
			if file == "" || i >= 120 {
				break
			}
			sb.WriteString("- ")
			sb.WriteString(file)
			sb.WriteString("\n")
		}

		keyFiles := selectKeyFiles(files)
		for _, file := range keyFiles {
			content, err := os.ReadFile(filepath.Join(worktree, file))
			if err != nil || len(content) > 20000 {
				continue
			}
			sb.WriteString("\n--- ")
			sb.WriteString(file)
			sb.WriteString(" ---\n")
			text := string(content)
			if len(text) > 4000 {
				text = text[:4000]
			}
			sb.WriteString(text)
			sb.WriteString("\n")
		}
	}

	text := sb.String()
	if len(text) > 16000 {
		return text[:16000]
	}
	return text
}

func selectKeyFiles(files []string) []string {
	names := map[string]bool{
		"README.md":                true,
		"package.json":             true,
		"go.mod":                   true,
		"pyproject.toml":           true,
		"requirements.txt":         true,
		"Cargo.toml":               true,
		"composer.json":            true,
		"Gemfile":                  true,
		"Makefile":                 true,
		"CONTRIBUTING.md":          true,
		".github/workflows/ci.yml": true,
	}

	var selected []string
	for _, file := range files {
		if names[file] {
			selected = append(selected, file)
		}
		if len(selected) >= 8 {
			break
		}
	}
	return selected
}

func safePath(value string) string {
	value = strings.ReplaceAll(value, "/", "__")
	value = strings.ReplaceAll(value, " ", "-")
	return value
}
