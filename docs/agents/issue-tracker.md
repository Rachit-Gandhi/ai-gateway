# Issue tracker: GitHub

Issues for this checkout are maintained on the upstream GitHub repository:

- Upstream issue tracker: `ferro-labs/ai-gateway`
- Local fork checkout: `Rachit-Gandhi/ai-gateway`

Use the `gh` CLI for issue operations. Prefer passing `--repo ferro-labs/ai-gateway` explicitly so issue reads and writes go to the upstream tracker rather than the fork.

## Conventions

- Create an issue: `gh issue create --repo ferro-labs/ai-gateway --title "..." --body "..."`
- Read an issue: `gh issue view --repo ferro-labs/ai-gateway <number> --comments`
- List open issues: `gh issue list --repo ferro-labs/ai-gateway --state open`
- List issues assigned to the current user: `gh issue list --repo ferro-labs/ai-gateway --assignee @me --state all`
- Comment on an issue: `gh issue comment --repo ferro-labs/ai-gateway <number> --body "..."`
- Apply or remove labels: `gh issue edit --repo ferro-labs/ai-gateway <number> --add-label "..."` / `--remove-label "..."`
- Close an issue: `gh issue close --repo ferro-labs/ai-gateway <number> --comment "..."`

Use heredocs for multi-line issue bodies.

## Local assigned-issue mirror

Issues assigned to `Rachit-Gandhi` from the upstream repository may be mirrored locally under `.scratch/issues/assigned/` for quick reference. This mirror is not the source of truth and should not be treated as an issue tracker.

Refresh the mirror with:

```bash
gh issue list --repo ferro-labs/ai-gateway --assignee @me --state all --limit 100
```

## When a skill says "publish to the issue tracker"

Create or update a GitHub issue on `ferro-labs/ai-gateway`, unless the user explicitly asks for a fork-local note or a `.scratch/` artifact.

## When a skill says "fetch the relevant ticket"

Run `gh issue view --repo ferro-labs/ai-gateway <number> --comments`.
