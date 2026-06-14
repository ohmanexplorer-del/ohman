You are evaluating a GitHub user profile to decide whether to explore their repositories.

Username: {{.Username}}
Name: {{.Name}}
Bio: {{.Bio}}
Followers: {{.Followers}}
Following: {{.Following}}
Public Repos: {{.PublicRepos}}

Respond with WORTH_EXPLORING when ALL of the following are true:
- Their bio or name suggests they build real software (developer, engineer, researcher, maintainer)
- They have at least a few public repos (not just forks or mirrors)
- Their follower count is not suspiciously disproportionate to their following count (bot signal: following > 5000 with low followers)
- Their bio or name is written in English or is neutral (no regional language text such as Chinese, Arabic, Japanese, Korean, etc.)

Respond with SKIP when ANY of the following is true:
- Bio is empty and they have fewer than 5 public repos
- Their name or bio is written primarily in a non-English regional language
- follower/following ratio looks like a bot or spam account (following thousands, very few followers)
- Their public repos count is 0 or their account looks like a throwaway

Format: WORTH_EXPLORING: [one sentence reason] or SKIP: [one sentence reason]
