// cmd/dummydash generates a dashboard.html with realistic dummy data for screenshots.
package main

import (
	"fmt"
	"os"

	"github.com/bajaj6/gh-pr-dashboard/internal/dashboard"
	gh "github.com/bajaj6/gh-pr-dashboard/internal/github"
)

func pr(title string, number int, org, repo, author string, isDraft bool, labels []gh.Label, teams []string, created, updated, reviewReqAt string) gh.PullRequest {
	p := gh.PullRequest{
		Title:     title,
		Number:    number,
		URL:       fmt.Sprintf("https://github.com/%s/%s/pull/%d", org, repo, number),
		CreatedAt: created,
		UpdatedAt: updated,
		State:     "OPEN",
		IsDraft:   isDraft,
		Author:    gh.Author{Login: author},
		Repository: gh.Repository{
			NameWithOwner: org + "/" + repo,
			Name:          repo,
		},
		Labels: labels,
		Teams:  teams,
	}
	if reviewReqAt != "" {
		p.ReviewRequestedAt = reviewReqAt
	}
	return p
}

func main() {
	data := &dashboard.Data{
		User:               "jchen",
		Generated:          "2026-03-21T14:30:00Z",
		Teams:              []string{"acme-corp/platform", "acme-corp/backend", "acme-corp/infra", "opensource/web-framework"},
		ExtraBotUsernames:  []string{"deploy-bot"},
		AutoRefreshMinutes: 15,
		StaleHours:         24,

		// My PRs — Approved (2)
		MyApproved: []gh.PullRequest{
			pr("feat: add OAuth2 PKCE flow for mobile clients", 1842, "acme-corp", "auth-service", "jchen", false,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}, {Name: "security", Color: "D93F0B"}},
				nil, "2026-03-18T09:15:00Z", "2026-03-21T11:30:00Z", ""),
			pr("fix: race condition in WebSocket connection pool", 3201, "acme-corp", "api-gateway", "jchen", false,
				[]gh.Label{{Name: "bug", Color: "D73A4A"}, {Name: "P1", Color: "B60205"}},
				nil, "2026-03-20T14:00:00Z", "2026-03-21T09:45:00Z", ""),
		},

		// My PRs — Draft (3)
		MyDraft: []gh.PullRequest{
			pr("feat: implement distributed rate limiter with Redis backend", 892, "acme-corp", "platform-core", "jchen", true,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}},
				nil, "2026-03-19T16:30:00Z", "2026-03-21T08:00:00Z", ""),
			pr("refactor: migrate user preferences to new schema v3", 445, "acme-corp", "user-service", "jchen", true,
				[]gh.Label{{Name: "refactor", Color: "E4E669"}},
				nil, "2026-03-17T11:00:00Z", "2026-03-20T15:20:00Z", ""),
			pr("docs: add architecture decision record for event sourcing", 156, "acme-corp", "platform-docs", "jchen", true,
				nil, nil, "2026-03-15T10:00:00Z", "2026-03-19T14:00:00Z", ""),
		},

		// My PRs — Open (not draft, not approved) (4)
		MyOpenRaw: []gh.PullRequest{
			pr("feat: add OAuth2 PKCE flow for mobile clients", 1842, "acme-corp", "auth-service", "jchen", false,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}, {Name: "security", Color: "D93F0B"}},
				nil, "2026-03-18T09:15:00Z", "2026-03-21T11:30:00Z", ""),
			pr("fix: race condition in WebSocket connection pool", 3201, "acme-corp", "api-gateway", "jchen", false,
				[]gh.Label{{Name: "bug", Color: "D73A4A"}, {Name: "P1", Color: "B60205"}},
				nil, "2026-03-20T14:00:00Z", "2026-03-21T09:45:00Z", ""),
			pr("feat: add gRPC health check endpoints for all microservices", 2105, "acme-corp", "service-mesh", "jchen", false,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}, {Name: "infrastructure", Color: "0075CA"}},
				nil, "2026-03-19T08:30:00Z", "2026-03-21T10:15:00Z", ""),
			pr("perf: optimize N+1 queries in order history endpoint", 567, "acme-corp", "order-service", "jchen", false,
				[]gh.Label{{Name: "performance", Color: "FBCA04"}},
				nil, "2026-03-16T13:45:00Z", "2026-03-20T17:00:00Z", ""),
			// include drafts too (myOpenRaw is the superset)
			pr("feat: implement distributed rate limiter with Redis backend", 892, "acme-corp", "platform-core", "jchen", true,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}},
				nil, "2026-03-19T16:30:00Z", "2026-03-21T08:00:00Z", ""),
			pr("refactor: migrate user preferences to new schema v3", 445, "acme-corp", "user-service", "jchen", true,
				[]gh.Label{{Name: "refactor", Color: "E4E669"}},
				nil, "2026-03-17T11:00:00Z", "2026-03-20T15:20:00Z", ""),
			pr("docs: add architecture decision record for event sourcing", 156, "acme-corp", "platform-docs", "jchen", true,
				nil, nil, "2026-03-15T10:00:00Z", "2026-03-19T14:00:00Z", ""),
		},

		// Review requested — directly to me (5)
		ReviewDirect: []gh.PullRequest{
			pr("feat: add request tracing with OpenTelemetry spans", 784, "acme-corp", "api-gateway", "amartinez", false,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}, {Name: "observability", Color: "5319E7"}},
				nil, "2026-03-21T08:00:00Z", "2026-03-21T12:30:00Z", "2026-03-21T08:01:00Z"),
			pr("fix: handle graceful shutdown for long-running batch jobs", 2341, "acme-corp", "job-scheduler", "priya-dev", false,
				[]gh.Label{{Name: "bug", Color: "D73A4A"}},
				nil, "2026-03-20T16:00:00Z", "2026-03-21T10:00:00Z", "2026-03-20T16:05:00Z"),
			pr("feat: add retry policy configuration for external API calls", 1156, "acme-corp", "platform-core", "liwei", false,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}},
				nil, "2026-03-20T11:30:00Z", "2026-03-21T09:00:00Z", "2026-03-20T11:35:00Z"),
			pr("security: upgrade JWT library to address CVE-2026-1234", 903, "acme-corp", "auth-service", "sarahk", false,
				[]gh.Label{{Name: "security", Color: "D93F0B"}, {Name: "P0", Color: "B60205"}},
				nil, "2026-03-19T15:00:00Z", "2026-03-20T18:00:00Z", "2026-03-19T15:10:00Z"),
			pr("refactor: extract payment processing into standalone module", 678, "acme-corp", "billing-service", "tomh", false,
				[]gh.Label{{Name: "refactor", Color: "E4E669"}},
				nil, "2026-03-18T09:00:00Z", "2026-03-20T14:30:00Z", "2026-03-18T09:05:00Z"),
		},

		// Review requested — via team (12)
		ReviewTeam: []gh.PullRequest{
			pr("feat: add Prometheus metrics for cache hit rates", 1567, "acme-corp", "metrics-collector", "dkumar", false,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}, {Name: "observability", Color: "5319E7"}},
				[]string{"acme-corp/platform"}, "2026-03-21T10:00:00Z", "2026-03-21T13:00:00Z", "2026-03-21T10:05:00Z"),
			pr("chore(deps): bump aws-sdk-go from v1.50 to v1.51", 445, "acme-corp", "cloud-infra", "dependabot", false,
				[]gh.Label{{Name: "dependencies", Color: "0366D6"}},
				[]string{"acme-corp/infra"}, "2026-03-21T06:00:00Z", "2026-03-21T06:01:00Z", "2026-03-21T06:01:00Z"),
			pr("fix: correct timezone handling in scheduled notifications", 2890, "acme-corp", "notification-service", "emilyr", false,
				[]gh.Label{{Name: "bug", Color: "D73A4A"}},
				[]string{"acme-corp/backend"}, "2026-03-20T19:00:00Z", "2026-03-21T08:30:00Z", "2026-03-20T19:05:00Z"),
			pr("feat: implement circuit breaker for downstream service calls", 934, "acme-corp", "platform-core", "alexj", false,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}, {Name: "reliability", Color: "0075CA"}},
				[]string{"acme-corp/platform"}, "2026-03-20T15:30:00Z", "2026-03-21T11:00:00Z", "2026-03-20T15:35:00Z"),
			pr("Update Terraform aws provider to v5.40", 312, "acme-corp", "cloud-infra", "renovate", false,
				[]gh.Label{{Name: "dependencies", Color: "0366D6"}, {Name: "terraform", Color: "5C4EE5"}},
				[]string{"acme-corp/infra"}, "2026-03-20T12:00:00Z", "2026-03-20T12:01:00Z", "2026-03-20T12:01:00Z"),
			pr("feat: add cursor-based pagination to search API", 1789, "acme-corp", "search-service", "marcow", false,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}},
				[]string{"acme-corp/backend"}, "2026-03-20T09:00:00Z", "2026-03-21T07:00:00Z", "2026-03-20T09:05:00Z"),
			pr("ci: add SAST scanning to CI pipeline", 456, "acme-corp", "ci-templates", "secureteam-bot", false,
				[]gh.Label{{Name: "ci", Color: "EDEDED"}, {Name: "security", Color: "D93F0B"}},
				[]string{"acme-corp/platform", "acme-corp/infra"}, "2026-03-19T14:00:00Z", "2026-03-20T16:00:00Z", "2026-03-19T14:05:00Z"),
			pr("fix: memory leak in event subscriber pool", 2456, "acme-corp", "event-bus", "liwei", false,
				[]gh.Label{{Name: "bug", Color: "D73A4A"}, {Name: "P1", Color: "B60205"}},
				[]string{"acme-corp/platform"}, "2026-03-19T11:00:00Z", "2026-03-20T10:00:00Z", "2026-03-19T11:05:00Z"),
			pr("chore: rotate database credentials and update vault paths", 789, "acme-corp", "secrets-manager", "ops-bot", false,
				[]gh.Label{{Name: "operations", Color: "FBCA04"}},
				[]string{"acme-corp/infra"}, "2026-03-18T16:00:00Z", "2026-03-19T09:00:00Z", "2026-03-18T16:05:00Z"),
			pr("feat: add support for custom webhook payloads", 345, "opensource", "web-framework", "contributor42", false,
				[]gh.Label{{Name: "enhancement", Color: "A2EEEF"}},
				[]string{"opensource/web-framework"}, "2026-03-18T08:00:00Z", "2026-03-20T14:00:00Z", "2026-03-18T08:05:00Z"),
			pr("docs: update API migration guide for v3 breaking changes", 678, "acme-corp", "platform-docs", "techwriter", false,
				[]gh.Label{{Name: "documentation", Color: "0075CA"}},
				[]string{"acme-corp/platform"}, "2026-03-17T10:00:00Z", "2026-03-19T16:00:00Z", "2026-03-17T10:05:00Z"),
			pr("Automated cruft update", 234, "acme-corp", "service-template", "github-actions", false,
				nil,
				[]string{"acme-corp/platform"}, "2026-03-16T09:00:00Z", "2026-03-16T09:01:00Z", "2026-03-16T09:01:00Z"),
		},

		// Approved by me (3)
		ApprovedMe: []gh.PullRequest{
			pr("feat: add structured logging with correlation IDs", 1234, "acme-corp", "logging-lib", "priya-dev", false,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}},
				nil, "2026-03-19T10:00:00Z", "2026-03-21T09:00:00Z", ""),
			pr("fix: prevent duplicate event processing in consumer group", 567, "acme-corp", "event-bus", "tomh", false,
				[]gh.Label{{Name: "bug", Color: "D73A4A"}},
				nil, "2026-03-18T14:00:00Z", "2026-03-20T11:00:00Z", ""),
			pr("perf: add connection pooling for Redis cluster client", 890, "acme-corp", "cache-layer", "alexj", false,
				[]gh.Label{{Name: "performance", Color: "FBCA04"}},
				nil, "2026-03-17T09:00:00Z", "2026-03-19T16:00:00Z", ""),
		},

		// Approved — via team (1)
		ApprovedTeam: []gh.PullRequest{
			pr("chore(deps): bump golang.org/x/crypto from 0.21 to 0.22", 456, "acme-corp", "auth-service", "dependabot", false,
				[]gh.Label{{Name: "dependencies", Color: "0366D6"}, {Name: "security", Color: "D93F0B"}},
				[]string{"acme-corp/platform"}, "2026-03-20T03:00:00Z", "2026-03-21T10:00:00Z", "2026-03-20T03:01:00Z"),
		},

		// My PRs — Changes Requested (1)
		MyChangesReq: []gh.PullRequest{
			pr("feat: add multi-region failover for primary database", 1098, "acme-corp", "cloud-infra", "jchen", false,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}, {Name: "infrastructure", Color: "0075CA"}},
				nil, "2026-03-14T10:00:00Z", "2026-03-20T13:00:00Z", ""),
		},

		// Review — Changes Requested (1)
		ReviewChangesReq: []gh.PullRequest{
			pr("feat: add custom serializer support for message queue", 1678, "acme-corp", "event-bus", "marcow", false,
				[]gh.Label{{Name: "feature", Color: "0E8A16"}},
				nil, "2026-03-17T14:00:00Z", "2026-03-20T09:00:00Z", ""),
		},
	}

	html, err := dashboard.Generate(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(html)
}
