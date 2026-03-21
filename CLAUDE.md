# Project Context

## What This Is
A GitHub CLI extension (`gh pr-dashboard`) that generates a self-contained HTML dashboard for reviewing PRs across multiple GitHub organizations. Written in Go, distributed as a `gh` extension.

## Architecture

### Go binary (`gh-pr-dashboard`)
- Compiled Go binary, distributed via `gh extension install`
- Dependencies: `gh` CLI (for authentication), Go standard library + a few packages
- Generates a self-contained HTML file with embedded JSON data and inline JS/CSS
- No tokens stored in the HTML — all API calls use `gh` CLI authentication

### Data flow
1. Parse CLI args → load config file → merge (CLI wins over config wins over defaults)
2. Discover user's team memberships via `gh api /user/teams`
3. Run a single combined GraphQL query with 7 search aliases
4. Classify each PR as direct vs team review using `reviewRequests` from GraphQL
5. Recover team-approved PRs from Notifications API
6. Generate HTML with embedded JSON → open in browser or serve via HTTP

### Key design decisions
- **Why static HTML, not client/server**: Each engineer needs their own `gh` auth. A server would require OAuth, token storage, hosting. Static HTML is a feature.
- **Why `gh extension`**: Zero-friction install (`gh extension install owner/gh-pr-dashboard`), auto-update, no PATH setup.
- **Why GraphQL**: Single API call fetches all 7 search result sets. REST would require 7 separate calls.
- **Why not enumerate teams in search**: Some orgs have 100+ child teams — combined query is too complex and returns 0 results.

### Serve mode (`--serve`)
- Starts a Go HTTP server
- Each page load regenerates fresh data
- No caching — reload = fresh API data

## Files

- `main.go` — CLI entry point, flag parsing, orchestration
- `internal/github/` — GitHub API client (GraphQL + REST), types, queries
- `internal/classify/` — PR classification logic (direct vs team, approved, etc.)
- `internal/config/` — Config file loading, CLI flag merging
- `internal/dashboard/` — HTML generation with embedded templates
- `internal/dashboard/templates/` — HTML template, CSS, JS
- `internal/server/` — HTTP serve mode
- `internal/logging/` — Simple logger
- `README.md` — open-source docs
- `LICENSE` — MIT
- `config.example.json` — example config for users to copy

## Config

Location: `~/.config/gh-pr-dashboard/config.json`

```json
{
  "orgs": ["my-org", "other-org"],
  "ignored_teams": [],
  "extra_bot_usernames": [],
  "output_file": "dashboard.html",
  "search_limit": 100
}
```

## CLI flags
```
--org ORG        Repeatable, overrides config orgs
--user USER      Override auto-detected GitHub username
--output FILE    Output HTML path (default: dashboard.html)
--limit N        Max PRs per query (default: 100)
--port PORT      Serve mode port (default: 8787)
--no-open        Don't auto-open browser
--serve          Start local HTTP server (reload = refresh)
--stdout         Output HTML to stdout (used internally by serve mode)
--config FILE    Custom config file path
--init           Create config file at default location
--help / --version
```

## Known patterns and gotchas

- **Bot detection**: Pattern-based (dependabot, renovate, etc.) in JS + exact usernames from config `extra_bot_usernames`.
- **Team filters**: Persisted in browser localStorage. Toggle individual teams on/off.
- **Time filter**: 1 week / 3 weeks / 1 month options, persisted in localStorage.
- **Dark mode**: Follows system preference via `prefers-color-scheme`, with manual toggle.
- **GraphQL retries**: Combined query retries up to 3 times with backoff on rate limit errors.
- **Notifications API**: Used to recover team-approved PRs that don't appear in search results.
