package github

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	ghapi "github.com/cli/go-gh/v2/pkg/api"
)

type Client struct {
	rest    *ghapi.RESTClient
	graphql *ghapi.GraphQLClient
}

func NewClient() (*Client, error) {
	rest, err := ghapi.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client (are you authenticated? run 'gh auth login'): %w", err)
	}
	gql, err := ghapi.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}
	return &Client{rest: rest, graphql: gql}, nil
}

// FetchUser returns the authenticated user's login.
func (c *Client) FetchUser() (string, error) {
	var user struct {
		Login string `json:"login"`
	}
	if err := c.rest.Get("user", &user); err != nil {
		return "", fmt.Errorf("failed to detect GitHub username: %w", err)
	}
	return user.Login, nil
}

// FetchAllPRs runs the combined 7-alias GraphQL query with retries.
func (c *Client) FetchAllPRs(queryBase, user string, limit int) (*CombinedSearchResponse, error) {
	if limit > 100 {
		limit = 100
	}

	vars := map[string]interface{}{
		"qReviewAll":        queryBase + " review-requested:" + user,
		"qApprovedReq":      queryBase + " review-requested:" + user + " review:approved",
		"qMyApproved":       queryBase + " author:" + user + " review:approved",
		"qMyOpen":           queryBase + " author:" + user,
		"qApprovedMe":       queryBase + " reviewed-by:" + user + " review:approved",
		"qMyChangesReq":     queryBase + " author:" + user + " review:changes_requested",
		"qReviewChangesReq": queryBase + " review-requested:" + user + " review:changes_requested",
		"limit":             limit,
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		var resp CombinedSearchResponse
		err := c.graphql.Do(CombinedQuery, vars, &resp.Data)
		if err == nil {
			return &resp, nil
		}

		lastErr = err
		errStr := err.Error()
		if strings.Contains(strings.ToLower(errStr), "rate limit") {
			time.Sleep(time.Duration(attempt*10) * time.Second)
		} else {
			time.Sleep(2 * time.Second)
		}
	}

	return nil, fmt.Errorf("GraphQL query failed after 3 attempts: %w", lastErr)
}

// FetchTeamsFromAPI fetches team memberships from /user/teams, filtered by orgs.
func (c *Client) FetchTeamsFromAPI(orgs []string) ([]string, error) {
	orgSet := make(map[string]bool, len(orgs))
	for _, o := range orgs {
		orgSet[o] = true
	}

	var teams []string
	page := 1
	for {
		var pageTeams []Team
		path := fmt.Sprintf("user/teams?per_page=100&page=%d", page)
		if err := c.rest.Get(path, &pageTeams); err != nil {
			return teams, err
		}
		if len(pageTeams) == 0 {
			break
		}
		for _, t := range pageTeams {
			if len(orgSet) == 0 || orgSet[t.Organization.Login] {
				teams = append(teams, t.Organization.Login+"/"+t.Slug)
			}
		}
		page++
	}
	sort.Strings(teams)
	return teams, nil
}

// FetchNotificationPRs discovers team-approved PRs via the Notifications API.
func (c *Client) FetchNotificationPRs(orgs []string, user, createdAfter string) []PullRequest {
	orgSet := make(map[string]bool, len(orgs))
	for _, o := range orgs {
		orgSet[o] = true
	}

	// Step 1: Fetch notifications
	var notifications []Notification
	if err := c.rest.Get("notifications?reason=review_requested&all=true&per_page=30", &notifications); err != nil {
		return nil
	}

	// Step 2: Extract and filter PR references
	refs := extractPRRefs(notifications, orgSet)
	if len(refs) == 0 {
		return nil
	}
	if len(refs) > 25 {
		refs = refs[:25]
	}

	// Step 3: Build and execute dynamic GraphQL query
	query := buildNotifQuery(refs)

	var rawResp map[string]json.RawMessage
	if err := c.graphql.Do(query, nil, &rawResp); err != nil {
		return nil
	}

	// Step 4: Parse and filter results
	var result []PullRequest
	for i := range refs {
		key := fmt.Sprintf("pr%d", i)
		raw, ok := rawResp[key]
		if !ok {
			continue
		}
		var wrapper struct {
			PullRequest *NotifPRNode `json:"pullRequest"`
		}
		if err := json.Unmarshal(raw, &wrapper); err != nil || wrapper.PullRequest == nil {
			continue
		}
		pr := filterNotifPR(wrapper.PullRequest, user, createdAfter)
		if pr != nil {
			result = append(result, *pr)
		}
	}
	return result
}

func extractPRRefs(notifications []Notification, orgSet map[string]bool) []NotifPRRef {
	seen := make(map[string]bool)
	var refs []NotifPRRef

	for _, n := range notifications {
		if n.Subject.Type != "PullRequest" {
			continue
		}
		// URL format: https://api.github.com/repos/{owner}/{repo}/pulls/{number}
		url := n.Subject.URL
		parts := strings.Split(url, "/")
		if len(parts) < 4 {
			continue
		}
		// Find "repos" index to extract owner/repo/number
		for i, p := range parts {
			if p == "repos" && i+3 < len(parts) {
				owner := parts[i+1]
				repo := parts[i+2]
				// parts[i+3] should be "pulls", parts[i+4] is number
				if i+4 < len(parts) && parts[i+3] == "pulls" {
					var number int
					if _, err := fmt.Sscanf(parts[i+4], "%d", &number); err == nil && (len(orgSet) == 0 || orgSet[owner]) {
						key := fmt.Sprintf("%s/%s/%d", owner, repo, number)
						if !seen[key] {
							seen[key] = true
							refs = append(refs, NotifPRRef{Owner: owner, Repo: repo, Number: number})
						}
					}
				}
				break
			}
		}
	}
	return refs
}

func buildNotifQuery(refs []NotifPRRef) string {
	var sb strings.Builder
	sb.WriteString(NotifPRFragment)
	sb.WriteString("\nquery {\n")
	for i, ref := range refs {
		fmt.Fprintf(&sb, "  pr%d: repository(owner: %q, name: %q) {\n", i, ref.Owner, ref.Repo)
		fmt.Fprintf(&sb, "    pullRequest(number: %d) { ...notifPrFields }\n", ref.Number)
		sb.WriteString("  }\n")
	}
	sb.WriteString("}")
	return sb.String()
}

func filterNotifPR(n *NotifPRNode, user, createdAfter string) *PullRequest {
	if n.State != "OPEN" || n.Author.Login == user {
		return nil
	}
	if n.ReviewDecision != "APPROVED" {
		return nil
	}
	// Filter by creation time if a cutoff date is provided.
	if createdAfter != "" {
		// Try to parse the PR creation time as RFC3339.
		prCreatedAt, errPr := time.Parse(time.RFC3339, n.CreatedAt)
		// Try to parse the cutoff date (YYYY-MM-DD) and set it to midnight UTC.
		cutoffDate, errCutoff := time.Parse("2006-01-02", createdAfter)
		if errPr == nil && errCutoff == nil {
			cutoff := time.Date(cutoffDate.Year(), cutoffDate.Month(), cutoffDate.Day(), 0, 0, 0, 0, time.UTC)
			if prCreatedAt.Before(cutoff) {
				return nil
			}
		} else {
			// Fall back to the original string comparison behavior on parse errors.
			if n.CreatedAt < createdAfter {
				return nil
			}
		}
	}

	// Extract teams from onBehalfOf
	teamSet := make(map[string]bool)
	for _, review := range n.Reviews.Nodes {
		for _, team := range review.OnBehalfOf.Nodes {
			teamSet[team.Organization.Login+"/"+team.Slug] = true
		}
	}
	if len(teamSet) == 0 {
		return nil
	}

	teams := make([]string, 0, len(teamSet))
	for t := range teamSet {
		teams = append(teams, t)
	}

	labels := make([]Label, 0, len(n.Labels.Nodes))
	for _, l := range n.Labels.Nodes {
		labels = append(labels, Label{Name: l.Name, Color: l.Color})
	}

	return &PullRequest{
		Title:     n.Title,
		Number:    n.Number,
		URL:       n.URL,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
		State:     n.State,
		IsDraft:   n.IsDraft,
		Author:    n.Author,
		Repository: Repository{
			NameWithOwner: n.Repository.NameWithOwner,
			Name:          n.Repository.Name,
		},
		Labels: labels,
		Teams:  teams,
	}
}
