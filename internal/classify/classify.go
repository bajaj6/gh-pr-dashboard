package classify

import (
	gh "github.com/bajaj6/gh-pr-dashboard/internal/github"
)

// Result holds all classified PR buckets.
type Result struct {
	MyApproved       []gh.PullRequest `json:"myApproved"`
	MyDraft          []gh.PullRequest `json:"myDraft"`
	MyOpenRaw        []gh.PullRequest `json:"myOpenRaw"`
	ReviewDirect     []gh.PullRequest `json:"reviewDirect"`
	ReviewTeam       []gh.PullRequest `json:"reviewTeam"`
	ApprovedMe       []gh.PullRequest `json:"approvedMe"`
	ApprovedTeam     []gh.PullRequest `json:"approvedTeam"`
	MyChangesReq     []gh.PullRequest `json:"myChangesReq"`
	ReviewChangesReq []gh.PullRequest `json:"reviewChangesReq"`
	DiscoveredTeams  []string         `json:"-"`
}

// Counts holds PR count breakdown for logging.
type Counts struct {
	MyApproved       int
	MyDraft          int
	MyOpen           int
	MyChangesReq     int
	ReviewAll        int
	ReviewChangesReq int
	ApprovedMe       int
	ApprovedReq      int
}

// Classify processes the raw GraphQL response into categorized buckets.
// This is a direct port of the jq JQ_PROCESS logic.
func Classify(resp *gh.CombinedSearchResponse) (*Result, *Counts) {
	// Extract and transform each alias
	reviewAll := extractTeamPRs(resp.Data.ReviewAll)
	approvedReq := extractTeamPRs(resp.Data.ApprovedReq)
	myApproved := extractStdPRs(resp.Data.MyApproved)
	myOpen := extractStdPRs(resp.Data.MyOpen)
	approvedMe := extractStdPRs(resp.Data.ApprovedMe)
	myChangesReq := extractStdPRs(resp.Data.MyChangesReq)
	reviewChangesReq := extractStdPRs(resp.Data.ReviewChangesReq)

	// Derive draft from open (excluding changes-requested PRs)
	myDraft := subtractByURL(filterDrafts(myOpen), myChangesReq)

	// Subtract changes-requested PRs from reviewAll before splitting
	reviewAllClean := subtractByURL(reviewAll, reviewChangesReq)

	// Split review into team vs direct
	var reviewTeam, reviewDirect []gh.PullRequest
	for _, pr := range reviewAllClean {
		if len(pr.Teams) > 0 {
			reviewTeam = append(reviewTeam, pr)
		} else {
			reviewDirect = append(reviewDirect, pr)
		}
	}

	// Approved sections
	approvedTeamRaw := subtractByURL(approvedReq, approvedMe)

	// Build team lookup from reviewTeam
	teamsByURL := make(map[string][]string)
	for _, pr := range reviewTeam {
		teamsByURL[pr.URL] = pr.Teams
	}

	// Inherit teams for approvedTeam where missing
	var approvedTeam []gh.PullRequest
	for _, pr := range approvedTeamRaw {
		if len(pr.Teams) == 0 {
			if teams, ok := teamsByURL[pr.URL]; ok {
				pr.Teams = teams
			}
		}
		approvedTeam = append(approvedTeam, pr)
	}

	approvedMeFinal := subtractByURL(approvedMe, approvedTeam)

	// Collect all discovered teams
	teamSet := make(map[string]bool)
	for _, pr := range reviewAll {
		for _, t := range pr.Teams {
			teamSet[t] = true
		}
	}
	for _, pr := range approvedReq {
		for _, t := range pr.Teams {
			teamSet[t] = true
		}
	}
	var discoveredTeams []string
	for t := range teamSet {
		discoveredTeams = append(discoveredTeams, t)
	}

	counts := &Counts{
		MyApproved:       len(myApproved),
		MyDraft:          len(myDraft),
		MyOpen:           len(myOpen),
		MyChangesReq:     len(myChangesReq),
		ReviewAll:        len(reviewAllClean),
		ReviewChangesReq: len(reviewChangesReq),
		ApprovedMe:       len(approvedMe),
		ApprovedReq:      len(approvedReq),
	}

	return &Result{
		MyApproved:       nonNil(myApproved),
		MyDraft:          nonNil(myDraft),
		MyOpenRaw:        nonNil(myOpen),
		ReviewDirect:     nonNil(reviewDirect),
		ReviewTeam:       nonNil(reviewTeam),
		ApprovedMe:       nonNil(approvedMeFinal),
		ApprovedTeam:     nonNil(approvedTeam),
		MyChangesReq:     nonNil(myChangesReq),
		ReviewChangesReq: nonNil(reviewChangesReq),
		DiscoveredTeams:  discoveredTeams,
	}, counts
}

// MergeNotifications deduplicates notification-discovered PRs against all
// existing buckets and merges new ones into ApprovedTeam.
func MergeNotifications(result *Result, notifPRs []gh.PullRequest, ignoredTeams []string) int {
	if len(notifPRs) == 0 {
		return 0
	}

	// Collect all known URLs
	known := make(map[string]bool)
	for _, bucket := range [][]gh.PullRequest{
		result.MyApproved, result.MyDraft, result.MyOpenRaw,
		result.ReviewDirect, result.ReviewTeam,
		result.ApprovedMe, result.ApprovedTeam,
		result.MyChangesReq, result.ReviewChangesReq,
	} {
		for _, pr := range bucket {
			known[pr.URL] = true
		}
	}

	ignoredSet := make(map[string]bool, len(ignoredTeams))
	for _, t := range ignoredTeams {
		ignoredSet[t] = true
	}

	var newPRs []gh.PullRequest
	for _, pr := range notifPRs {
		if known[pr.URL] {
			continue
		}
		// Filter ignored teams from the PR's team list
		var filteredTeams []string
		for _, t := range pr.Teams {
			if !ignoredSet[t] {
				filteredTeams = append(filteredTeams, t)
			}
		}
		// Skip PRs that have no non-ignored teams
		if len(filteredTeams) == 0 {
			continue
		}
		pr.Teams = filteredTeams
		newPRs = append(newPRs, pr)
	}

	if len(newPRs) > 0 {
		result.ApprovedTeam = append(result.ApprovedTeam, newPRs...)

		// Discover new teams
		teamSet := make(map[string]bool)
		for _, t := range result.DiscoveredTeams {
			teamSet[t] = true
		}
		for _, pr := range newPRs {
			for _, t := range pr.Teams {
				if !ignoredSet[t] {
					teamSet[t] = true
				}
			}
		}
		result.DiscoveredTeams = nil
		for t := range teamSet {
			result.DiscoveredTeams = append(result.DiscoveredTeams, t)
		}
	}

	return len(newPRs)
}

// Helper functions

func extractStdPRs(sr gh.SearchResult) []gh.PullRequest {
	prs := make([]gh.PullRequest, 0, len(sr.Edges))
	for _, e := range sr.Edges {
		prs = append(prs, e.Node.ToStdPR())
	}
	return prs
}

func extractTeamPRs(sr gh.SearchResult) []gh.PullRequest {
	prs := make([]gh.PullRequest, 0, len(sr.Edges))
	for _, e := range sr.Edges {
		prs = append(prs, e.Node.ToTeamPR())
	}
	return prs
}

func subtractByURL(source, exclude []gh.PullRequest) []gh.PullRequest {
	urls := make(map[string]bool, len(exclude))
	for _, pr := range exclude {
		urls[pr.URL] = true
	}
	var result []gh.PullRequest
	for _, pr := range source {
		if !urls[pr.URL] {
			result = append(result, pr)
		}
	}
	return result
}

func filterDrafts(prs []gh.PullRequest) []gh.PullRequest {
	var result []gh.PullRequest
	for _, pr := range prs {
		if pr.IsDraft {
			result = append(result, pr)
		}
	}
	return result
}

// nonNil ensures a nil slice becomes an empty slice for JSON marshaling.
func nonNil(prs []gh.PullRequest) []gh.PullRequest {
	if prs == nil {
		return []gh.PullRequest{}
	}
	return prs
}
