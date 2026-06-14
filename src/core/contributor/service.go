package contributor

import (
	"context"
	"fmt"
	"log"
	ght "ohman/src/core/github"
	"ohman/src/core/prompts"
	"os"
	"strings"
)

func NewService(client *ght.Client, ai AIService, config Config) *Service {
	if config.MaxIssuesPerRun <= 0 {
		config.MaxIssuesPerRun = 2
	}
	if config.MaxRepoSizeKB <= 0 {
		config.MaxRepoSizeKB = 50000
	}
	if config.MaxRepoStars <= 0 {
		config.MaxRepoStars = 5000
	}
	if config.MaxRepoForks <= 0 {
		config.MaxRepoForks = 1000
	}
	if config.WorkDir == "" {
		config.WorkDir = "work/contributions"
	}
	if len(config.Labels) == 0 {
		config.Labels = []string{"good first issue", "bug"}
	}

	return &Service{client: client, ai: ai, config: config}
}

func (s *Service) Run(ctx context.Context) ([]CandidateReport, error) {
	if !s.config.Enabled {
		return nil, nil
	}

	issues, err := s.findIssues(ctx)
	if err != nil {
		return nil, err
	}

	reports := make([]CandidateReport, 0, len(issues))
	for _, issue := range issues {
		complexity, err := s.client.GetRepoComplexity(ctx, issue.Owner, issue.Repo)
		if err != nil {
			reports = append(reports, CandidateReport{
				Issue:    issue,
				Accepted: false,
				Reasons:  []string{err.Error()},
			})
			continue
		}

		accepted, reasons := s.accepts(complexity)
		report := CandidateReport{
			Issue:      issue,
			Complexity: complexity,
			Reasons:    reasons,
		}

		if !accepted {
			reports = append(reports, report)
			continue
		}

		triage, err := s.triage(ctx, issue, complexity)
		if err != nil {
			report.Reasons = append(report.Reasons, err.Error())
			reports = append(reports, report)
			continue
		}
		report.Triage = triage
		if !triage.CanFix {
			report.Reasons = []string{triage.Reason}
			reports = append(reports, report)
			continue
		}

		report.Accepted = true
		report.Reasons = []string{triage.Reason}

		if s.config.AutoPR {
			worktree, err := s.clone(ctx, issue)
			if err != nil {
				report.Accepted = false
				report.Reasons = append(report.Reasons, err.Error())
			} else {
				report.Worktree = worktree
				func() {
					defer func() {
						if err := os.RemoveAll(worktree); err != nil {
							report.Reasons = append(report.Reasons, fmt.Sprintf("cleanup failed: %v", err))
							report.Cleaned = false
							return
						}
						report.Cleaned = true
					}()

					patch, pr, err := s.fixAndCreatePR(ctx, issue, complexity, worktree)
					report.Patch = patch
					report.PullRequest = pr
					if err != nil {
						report.Accepted = false
						report.Reasons = append(report.Reasons, err.Error())
						return
					}
					if pr != nil {
						report.Reasons = append(report.Reasons, "pull request created: "+pr.URL)
					}
				}()
			}
			if !report.Accepted {
				reports = append(reports, report)
				continue
			}
		}

		reports = append(reports, report)
	}

	log.Printf("Contributor scan finished: %d issues reviewed", len(reports))
	return reports, nil
}

func (s *Service) findIssues(ctx context.Context) ([]*ght.IssueCandidate, error) {
	remaining := s.config.MaxIssuesPerRun
	seen := map[string]bool{}
	var out []*ght.IssueCandidate

	for _, label := range s.config.Labels {
		if remaining <= 0 {
			break
		}
		issues, err := s.client.SearchIssueCandidates(ctx, label, remaining)
		if err != nil {
			return nil, err
		}
		for _, issue := range issues {
			key := issue.FullName + "#" + fmt.Sprint(issue.Number)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, issue)
			remaining--
			if remaining <= 0 {
				break
			}
		}
	}

	return out, nil
}

func (s *Service) triage(ctx context.Context, issue *ght.IssueCandidate, complexity *ght.RepoComplexity) (*TriageDecision, error) {
	if s.ai == nil {
		return &TriageDecision{
			CanFix:     false,
			Confidence: 0,
			Reason:     "AI triage is not configured",
			Risks:      []string{"missing AI service"},
		}, nil
	}

	body := issue.Body
	if len(body) > 2000 {
		body = body[:2000]
	}

	prompt, err := prompts.RenderContributorTriageIssue(prompts.ContributorTriageIssuePromptData{
		FullName:   issue.FullName,
		Number:     issue.Number,
		Title:      issue.Title,
		Labels:     strings.Join(issue.Labels, ", "),
		Comments:   issue.Comments,
		Body:       body,
		SizeKB:     complexity.SizeKB,
		Stars:      complexity.Stars,
		Forks:      complexity.Forks,
		OpenIssues: complexity.OpenIssues,
		Language:   complexity.Language,
	})
	if err != nil {
		return nil, err
	}

	response, err := s.ai.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI triage failed: %w", err)
	}

	decision, err := parseTriageDecision(response)
	if err != nil {
		return nil, fmt.Errorf("AI triage parse failed: %w", err)
	}
	if decision.Reason == "" {
		decision.Reason = "AI triage did not provide a reason"
	}
	return decision, nil
}

func (s *Service) fixAndCreatePR(ctx context.Context, issue *ght.IssueCandidate, complexity *ght.RepoComplexity, worktree string) (*PatchDecision, *ght.PullRequestData, error) {
	patch, err := s.generatePatch(ctx, issue, complexity, worktree)
	if err != nil {
		return nil, nil, err
	}
	if !patch.CanPatch {
		if patch.Summary == "" {
			patch.Summary = "AI declined to patch after inspecting the cloned repository"
		}
		return patch, nil, fmt.Errorf("%s", patch.Summary)
	}
	if strings.TrimSpace(patch.Diff) == "" {
		return patch, nil, fmt.Errorf("AI patch response did not include a diff")
	}

	branch := safeBranchName(patch.BranchName, issue)
	if err := git(ctx, worktree, "checkout", "-b", branch); err != nil {
		return patch, nil, err
	}
	if err := applyDiff(ctx, worktree, patch.Diff); err != nil {
		return patch, nil, err
	}
	if clean, err := gitClean(ctx, worktree); err != nil {
		return patch, nil, err
	} else if clean {
		return patch, nil, fmt.Errorf("AI patch produced no file changes")
	}
	if err := runValidation(ctx, worktree); err != nil {
		return patch, nil, err
	}

	commitMessage := strings.TrimSpace(patch.CommitMessage)
	if commitMessage == "" {
		commitMessage = fmt.Sprintf("Fix #%d: %s", issue.Number, issue.Title)
	}
	if err := git(ctx, worktree, "config", "user.name", "Ohman Explorer Bot"); err != nil {
		return patch, nil, err
	}
	if err := git(ctx, worktree, "config", "user.email", "ohmanexplorer-del@users.noreply.github.com"); err != nil {
		return patch, nil, err
	}
	if err := git(ctx, worktree, "add", "-A"); err != nil {
		return patch, nil, err
	}
	if err := git(ctx, worktree, "commit", "-m", commitMessage); err != nil {
		return patch, nil, err
	}

	if err := s.client.EnsureFork(ctx, issue.Owner, issue.Repo); err != nil {
		return patch, nil, err
	}
	if err := pushBranch(ctx, worktree, s.client.Token(), s.client.BotUsername(), issue.Repo, branch); err != nil {
		return patch, nil, err
	}

	title := strings.TrimSpace(patch.PRTitle)
	if title == "" {
		title = commitMessage
	}
	body := strings.TrimSpace(patch.PRBody)
	if body == "" {
		body = fmt.Sprintf("Fixes #%d\n\n%s", issue.Number, patch.Summary)
	}

	pr, err := s.client.CreatePullRequest(ctx, issue.Owner, issue.Repo, branch, complexity.DefaultBranch, title, body)
	if err != nil {
		return patch, nil, err
	}
	return patch, pr, nil
}

func (s *Service) generatePatch(ctx context.Context, issue *ght.IssueCandidate, complexity *ght.RepoComplexity, worktree string) (*PatchDecision, error) {
	if s.ai == nil {
		return nil, fmt.Errorf("AI service is not configured")
	}

	contextText := repoContext(worktree)
	body := issue.Body
	if len(body) > 2000 {
		body = body[:2000]
	}

	prompt, err := prompts.RenderContributorGeneratePatch(prompts.ContributorGeneratePatchPromptData{
		FullName:      issue.FullName,
		Number:        issue.Number,
		Title:         issue.Title,
		Labels:        strings.Join(issue.Labels, ", "),
		Body:          body,
		Language:      complexity.Language,
		SizeKB:        complexity.SizeKB,
		DefaultBranch: complexity.DefaultBranch,
		ContextText:   contextText,
	})
	if err != nil {
		return nil, err
	}

	response, err := s.ai.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI patch generation failed: %w", err)
	}
	patch, err := parsePatchDecision(response)
	if err != nil {
		return nil, fmt.Errorf("AI patch parse failed: %w", err)
	}
	return patch, nil
}

func (s *Service) accepts(repo *ght.RepoComplexity) (bool, []string) {
	var reasons []string

	if repo.Archived {
		reasons = append(reasons, "repo is archived")
	}
	if repo.Disabled {
		reasons = append(reasons, "repo is disabled")
	}
	if repo.SizeKB > s.config.MaxRepoSizeKB {
		reasons = append(reasons, fmt.Sprintf("repo size %dKB exceeds limit %dKB", repo.SizeKB, s.config.MaxRepoSizeKB))
	}
	if repo.Stars > s.config.MaxRepoStars {
		reasons = append(reasons, fmt.Sprintf("stars %d exceeds limit %d", repo.Stars, s.config.MaxRepoStars))
	}
	if repo.Forks > s.config.MaxRepoForks {
		reasons = append(reasons, fmt.Sprintf("forks %d exceeds limit %d", repo.Forks, s.config.MaxRepoForks))
	}
	if repo.DefaultBranch == "" {
		reasons = append(reasons, "default branch is empty")
	}

	if len(reasons) > 0 {
		return false, reasons
	}
	return true, []string{"repo is within contribution budget"}
}
