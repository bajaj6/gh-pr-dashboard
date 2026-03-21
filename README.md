# gh-pr-dashboard

A GitHub CLI extension that gives you a single view of every PR that needs your attention — across teams, repos and orgs. No more tab-hopping between GitHub notifications, Slack pings, and email reminders. One dashboard, done.

![Dashboard Demo](demo.gif)

## Install & Start

```bash
gh extension install deliveroo/gh-pr-dashboard
# Live dashboard
gh pr-dashboard --serve
```

### Prerequisites

- [GitHub CLI](https://cli.github.com/) (`gh`) — authenticated with `gh auth login`

If you don't have `gh`:

```bash
brew install gh  # e.g. if on macOS
```

Or, see the [installation instructions](https://github.com/cli/cli#installation) for other platforms.

### Upgrade

```bash
gh extension upgrade pr-dashboard
```

## Dashboard Layout

### Pending Reviews
- **Direct review requests** — PRs where you are personally requested
- **Team review requests** — PRs where your teams are requested (with team toggle filters)
- **Changes requested** — PRs where you previously requested changes
- **Bot (Automated) PRs** — Bot-authored PRs needing review (collapsed by default)
- **Drafts** — Draft PRs where you are a reviewer (collapsed by default)

### Approved Reviews
- **Approved by me** — PRs you have approved
- **Approved by team** — PRs your teams approved

### My Pull Requests
- **Approved** — Your PRs that are approved and ready to merge
- **Changes requested** — Your PRs where reviewers requested changes
- **Open** — Your PRs awaiting review
- **Drafts** — Your draft PRs

## Configuration

Optionally create a config file at `~/.config/gh-pr-dashboard/config.json` to customize behavior:

```json
{
  "orgs": ["my-org", "other-org"],
  "ignored_teams": ["my-org/noisy-team"],
  "extra_bot_usernames": ["our-deploy-bot"],
  "output_file": "dashboard.html",
  "search_limit": 100,
  "auto_refresh_minutes": 15,
  "stale_hours": 24
}
```

You can create this file manually, or run `gh pr-dashboard --init` to generate a starter config.

| Field | Default | Description |
|---|---|---|
| `orgs` | `[]` (all) | GitHub organizations to scan. Empty = all orgs you have access to |
| `ignored_teams` | `[]` | Teams to hide from the dashboard (format: `org/team-slug`) |
| `extra_bot_usernames` | `[]` | Additional usernames to treat as bots (e.g. `["deploy-bot", "ci-user"]`) |
| `output_file` | `"dashboard.html"` | Output HTML file path |
| `search_limit` | `100` | Max PRs per search query (max 100) |
| `auto_refresh_minutes` | `15` | Auto-refresh interval in minutes (0 to disable) |
| `stale_hours` | `24` | Hours after which a PR is highlighted as "Waiting" |

### CLI Flags

```
--org ORG        Scan specific org (repeatable, overrides config)
--user USER      Override auto-detected GitHub username
--output FILE    Output HTML path (default: dashboard.html)
--limit N        Max PRs per query (default: 100)
--serve          Start local HTTP server (reload = refresh)
--port PORT      Serve mode port (default: 8787)
--no-open        Don't auto-open browser
--init           Create starter config file
--config FILE    Custom config file path
--stdout         Output HTML to stdout
--help           Show help
--version        Show version
```

CLI flags override config values for `orgs`, `output_file`, and `search_limit`.

> **Note:** In serve mode, config is read at startup. Restart the server after editing `config.json` for changes to take effect.

## Features

- **Zero-config** — Works out of the box, no setup required
- **Single GraphQL query** — All 7 search queries combined into one API call (~4s total)
- **Smart team classification** — Uses GraphQL `reviewRequests` to accurately split direct vs team review requests
- **Time filter** — Filter PRs by 1 week / 3 weeks / 1 month
- **Team filter** — Toggle individual teams on/off (sorted by PR count, persisted in localStorage)
- **Bot detection** — Common bot patterns + configurable extra usernames
- **Waiting indicator** — PRs pending review longer than the configured threshold (default: 24h)
- **Dark mode** — Toggle between light, dark, and system theme
- **Serve mode** — `--serve` starts a local server; reload the page to refresh data
- **Auto-refresh** — Configurable auto-refresh interval (default: 15 min)
- **Self-contained** — Generated HTML has no external dependencies; share it, bookmark it, or host it

## How It Works

1. Fetches all PR data in a single GitHub GraphQL API call (7 searches via aliases)
2. Classifies each PR as direct vs team review using `reviewRequests` from GraphQL
3. Generates a self-contained HTML file with embedded JSON data
4. Opens it in your browser (or serves it via local HTTP server)

No tokens are stored in the HTML. All API calls use your `gh` CLI authentication.

## License

MIT
