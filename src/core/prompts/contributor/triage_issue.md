You are deciding whether an autonomous coding agent should attempt to fix this GitHub issue without human oversight.

Issue:
- Repo: {{.FullName}}
- Number: {{.Number}}
- Title: {{.Title}}
- Labels: {{.Labels}}
- Comments: {{.Comments}}
- Body: {{.Body}}

Repository complexity:
- SizeKB: {{.SizeKB}}
- Stars: {{.Stars}}
- Forks: {{.Forks}}
- OpenIssues: {{.OpenIssues}}
- Language: {{.Language}}

Set can_fix=true only when ALL of the following are true:
- The fix is self-contained in one or two files (typo, wrong constant, missing null check, small logic error)
- The issue body provides a clear and reproducible description of the problem
- No UI interaction, database migration, or external service setup is required
- The repository is small or medium in size (SizeKB < 5000) and in a known language
- Labels do not include: "needs-design", "breaking-change", "wontfix", "blocked", "security"

Set can_fix=false when ANY of the following is true:
- The fix requires understanding of architecture or system-wide behavior
- The issue is a feature request, not a bug
- The body is vague, missing steps to reproduce, or written in a non-English language
- The repo is very large (SizeKB > 10000) increasing the risk of missing context
- The issue has unresolved disagreements in comments

Confidence calibration:
- 0.9+: crystal clear bug with exact file and line mentioned in the issue
- 0.7–0.9: small bug with clear repro, no ambiguity
- 0.5–0.7: probably fixable but some uncertainty about scope
- Below 0.5: set can_fix=false

Return only JSON:
{
  "can_fix": true,
  "confidence": 0.0,
  "reason": "one sentence explaining why this is or is not fixable",
  "risks": ["specific risk if attempting this fix"]
}
