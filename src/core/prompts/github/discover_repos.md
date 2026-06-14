You are deciding the next discovery strategy for an autonomous GitHub explorer.

User Profile:
- Username: {{.Username}}
- Bio: {{.Bio}}
- Followers: {{.Followers}}
- Following: {{.Following}}

Current Discovery Queue Size: {{.QueueSize}}
Repos Discovered So Far: {{.DiscoveredCount}}

Choose one strategy:

DEEP_DIVE — explore more repos from this specific user
  Use when: user has a strong profile, many public repos, and we have not yet sampled their work deeply.
  Avoid when: we already have several repos from this user or their profile is thin.

EXPAND — queue followers and following of this user for future exploration
  Use when: user has a notable follower base (100+) suggesting they are embedded in a real tech community.
  Avoid when: queue is already large (>20 pending) or user has very few connections.

SEARCH — trigger a GitHub keyword search based on topics related to this user's work
  Use when: user's bio hints at a niche (e.g., "Rust", "distributed systems", "ML infra") and we want to surface similar projects beyond their personal graph.
  Avoid when: bio is generic ("developer", "student") or DiscoveredCount is already at target.

SKIP — move to the next user without further action
  Use when: profile is thin, bio is empty, or no clear signal about what they build.

Format:
STRATEGY: [DEEP_DIVE/EXPAND/SEARCH/SKIP]
REASON: [one sentence explaining why this strategy fits right now]
