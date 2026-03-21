package github

// PullRequest is the standard PR shape used throughout the pipeline.
// Field names must match what the frontend JS expects.
type PullRequest struct {
	Title              string     `json:"title"`
	Number             int        `json:"number"`
	URL                string     `json:"url"`
	CreatedAt          string     `json:"createdAt"`
	UpdatedAt          string     `json:"updatedAt"`
	State              string     `json:"state"`
	IsDraft            bool       `json:"isDraft"`
	Author             Author     `json:"author"`
	Repository         Repository `json:"repository"`
	Labels             []Label    `json:"labels"`
	Teams              []string   `json:"_teams,omitempty"`
	ReviewRequestedAt  string     `json:"reviewRequestedAt,omitempty"`

	// ReviewDecision is used internally during classification but not sent to frontend.
	ReviewDecision string `json:"-"`
}

type Author struct {
	Login string `json:"login"`
}

type Repository struct {
	NameWithOwner string `json:"nameWithOwner"`
	Name          string `json:"name"`
}

type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// GraphQL response types for the combined 7-alias query.

type CombinedSearchResponse struct {
	Data struct {
		ReviewAll        SearchResult `json:"reviewAll"`
		ApprovedReq      SearchResult `json:"approvedReq"`
		MyApproved       SearchResult `json:"myApproved"`
		MyOpen           SearchResult `json:"myOpen"`
		ApprovedMe       SearchResult `json:"approvedMe"`
		MyChangesReq     SearchResult `json:"myChangesReq"`
		ReviewChangesReq SearchResult `json:"reviewChangesReq"`
	} `json:"data"`
}

type SearchResult struct {
	Edges []struct {
		Node PRNode `json:"node"`
	} `json:"edges"`
}

type PRNode struct {
	Title      string `json:"title"`
	Number     int    `json:"number"`
	URL        string `json:"url"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
	State      string `json:"state"`
	IsDraft    bool   `json:"isDraft"`
	Author     Author `json:"author"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
		Name          string `json:"name"`
	} `json:"repository"`
	Labels struct {
		Nodes []Label `json:"nodes"`
	} `json:"labels"`
	ReviewRequests *struct {
		Nodes []struct {
			RequestedReviewer ReviewerNode `json:"requestedReviewer"`
		} `json:"nodes"`
	} `json:"reviewRequests,omitempty"`
	TimelineItems *struct {
		Nodes []struct {
			CreatedAt string `json:"createdAt"`
		} `json:"nodes"`
	} `json:"timelineItems,omitempty"`
}

type ReviewerNode struct {
	Typename     string `json:"__typename"`
	Login        string `json:"login,omitempty"`
	Slug         string `json:"slug,omitempty"`
	Organization *struct {
		Login string `json:"login"`
	} `json:"organization,omitempty"`
}

// Notification types

type Notification struct {
	Subject struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"subject"`
}

type NotifPRRef struct {
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
	Number int    `json:"number"`
}

// Notification GraphQL response types

type NotifPRNode struct {
	Title          string `json:"title"`
	Number         int    `json:"number"`
	URL            string `json:"url"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
	State          string `json:"state"`
	IsDraft        bool   `json:"isDraft"`
	ReviewDecision string `json:"reviewDecision"`
	Author         Author `json:"author"`
	Repository     struct {
		NameWithOwner string `json:"nameWithOwner"`
		Name          string `json:"name"`
	} `json:"repository"`
	Labels struct {
		Nodes []Label `json:"nodes"`
	} `json:"labels"`
	Reviews struct {
		Nodes []struct {
			Author    Author `json:"author"`
			OnBehalfOf struct {
				Nodes []struct {
					Slug         string `json:"slug"`
					Organization struct {
						Login string `json:"login"`
					} `json:"organization"`
				} `json:"nodes"`
			} `json:"onBehalfOf"`
		} `json:"nodes"`
	} `json:"reviews"`
}

// Team types

type Team struct {
	Slug         string `json:"slug"`
	Organization struct {
		Login string `json:"login"`
	} `json:"organization"`
}

// ToStdPR converts a PRNode to a PullRequest (no team info).
func (n *PRNode) ToStdPR() PullRequest {
	labels := make([]Label, 0, len(n.Labels.Nodes))
	for _, l := range n.Labels.Nodes {
		labels = append(labels, Label{Name: l.Name, Color: l.Color})
	}
	pr := PullRequest{
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
	}
	if n.TimelineItems != nil && len(n.TimelineItems.Nodes) > 0 && n.TimelineItems.Nodes[0].CreatedAt != "" {
		pr.ReviewRequestedAt = n.TimelineItems.Nodes[0].CreatedAt
	}
	return pr
}

// ToTeamPR converts a PRNode to a PullRequest with teams extracted from reviewRequests.
func (n *PRNode) ToTeamPR() PullRequest {
	pr := n.ToStdPR()
	if n.ReviewRequests != nil {
		for _, rr := range n.ReviewRequests.Nodes {
			r := rr.RequestedReviewer
			if r.Typename == "Team" && r.Organization != nil {
				pr.Teams = append(pr.Teams, r.Organization.Login+"/"+r.Slug)
			}
		}
	}
	return pr
}
