package contributor

import "testing"

func TestParseTriageDecision(t *testing.T) {
	decision, err := parseTriageDecision(`answer:
{
  "can_fix": true,
  "confidence": 0.82,
  "reason": "Small issue with clear reproduction.",
  "risks": ["May need project-specific setup"]
}`)
	if err != nil {
		t.Fatal(err)
	}
	if !decision.CanFix {
		t.Fatal("can_fix should be true")
	}
	if decision.Confidence != 0.82 {
		t.Fatalf("confidence = %v", decision.Confidence)
	}
	if decision.Reason == "" {
		t.Fatal("reason should not be empty")
	}
}

func TestParseTriageDecisionClampsConfidence(t *testing.T) {
	decision, err := parseTriageDecision(`{"can_fix":false,"confidence":2,"reason":"too broad","risks":[]}`)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Confidence != 1 {
		t.Fatalf("confidence = %v", decision.Confidence)
	}
}

func TestParsePatchDecision(t *testing.T) {
	patch, err := parsePatchDecision(`text before
{
  "can_patch": true,
  "summary": "Updates README",
  "branch_name": "ohman/fix-readme",
  "commit_message": "Fix README typo",
  "pr_title": "Fix README typo",
  "pr_body": "Fixes #1\n\nSummary",
  "diff": "diff --git a/README.md b/README.md\n--- a/README.md\n+++ b/README.md\n"
}`)
	if err != nil {
		t.Fatal(err)
	}
	if !patch.CanPatch {
		t.Fatal("can_patch should be true")
	}
	if patch.BranchName != "ohman/fix-readme" {
		t.Fatalf("branch_name = %q", patch.BranchName)
	}
	if patch.Diff == "" {
		t.Fatal("diff should not be empty")
	}
}
