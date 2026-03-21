package github

// FilterIgnoredTeams removes ignored teams from the list and returns both filtered and ignored lists.
func FilterIgnoredTeams(teams, ignored []string) (filtered, skipped []string) {
	ignoredSet := make(map[string]bool, len(ignored))
	for _, t := range ignored {
		ignoredSet[t] = true
	}
	for _, t := range teams {
		if ignoredSet[t] {
			skipped = append(skipped, t)
		} else {
			filtered = append(filtered, t)
		}
	}
	return
}
