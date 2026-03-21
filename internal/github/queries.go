package github

// CombinedQuery is the main GraphQL query with 7 search aliases.
const CombinedQuery = `
fragment prFields on PullRequest {
  title number url createdAt updatedAt state isDraft
  author { login }
  repository { nameWithOwner name }
  labels(first: 20) { nodes { name color } }
}
query($qReviewAll: String!, $qApprovedReq: String!, $qMyApproved: String!, $qMyOpen: String!, $qApprovedMe: String!, $qMyChangesReq: String!, $qReviewChangesReq: String!, $limit: Int!) {
  reviewAll: search(query: $qReviewAll, type: ISSUE, first: $limit) {
    edges { node { ... on PullRequest { ...prFields
      reviewRequests(first: 50) { nodes { requestedReviewer {
        __typename
        ... on Team { slug organization { login } }
        ... on User { login }
      }}}
      timelineItems(itemTypes: [REVIEW_REQUESTED_EVENT], last: 1) {
        nodes { ... on ReviewRequestedEvent { createdAt } }
      }
    }}}
  }
  approvedReq: search(query: $qApprovedReq, type: ISSUE, first: $limit) {
    edges { node { ... on PullRequest { ...prFields
      reviewRequests(first: 50) { nodes { requestedReviewer {
        __typename
        ... on Team { slug organization { login } }
        ... on User { login }
      }}}
      timelineItems(itemTypes: [REVIEW_REQUESTED_EVENT], last: 1) {
        nodes { ... on ReviewRequestedEvent { createdAt } }
      }
    }}}
  }
  myApproved: search(query: $qMyApproved, type: ISSUE, first: $limit) {
    edges { node { ... on PullRequest { ...prFields }}}
  }
  myOpen: search(query: $qMyOpen, type: ISSUE, first: $limit) {
    edges { node { ... on PullRequest { ...prFields }}}
  }
  approvedMe: search(query: $qApprovedMe, type: ISSUE, first: $limit) {
    edges { node { ... on PullRequest { ...prFields }}}
  }
  myChangesReq: search(query: $qMyChangesReq, type: ISSUE, first: $limit) {
    edges { node { ... on PullRequest { ...prFields }}}
  }
  reviewChangesReq: search(query: $qReviewChangesReq, type: ISSUE, first: $limit) {
    edges { node { ... on PullRequest { ...prFields
      timelineItems(itemTypes: [REVIEW_REQUESTED_EVENT], last: 1) {
        nodes { ... on ReviewRequestedEvent { createdAt } }
      }
    }}}
  }
}`

// NotifPRFragment is the GraphQL fragment for notification-discovered PRs.
const NotifPRFragment = `
fragment notifPrFields on PullRequest {
  title number url createdAt updatedAt state isDraft reviewDecision
  author { login }
  repository { nameWithOwner name }
  labels(first: 20) { nodes { name color } }
  reviews(states: APPROVED, first: 20) {
    nodes {
      author { login }
      onBehalfOf(first: 5) { nodes { slug organization { login } } }
    }
  }
}`
