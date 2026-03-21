package dashboard

import (
	"bytes"
	"embed"
	"encoding/json"
	"strings"
	"text/template"
	"time"

	"github.com/bajaj6/gh-pr-dashboard/internal/classify"
	gh "github.com/bajaj6/gh-pr-dashboard/internal/github"
)

//go:embed all:templates
var templateFS embed.FS

// Data is the top-level JSON structure embedded in the HTML.
// Field names MUST match what the frontend JS expects.
type Data struct {
	User               string           `json:"user"`
	Generated          string           `json:"generated"`
	Teams              []string         `json:"teams"`
	ExtraBotUsernames  []string         `json:"extraBotUsernames"`
	AutoRefreshMinutes int              `json:"autoRefreshMinutes"`
	StaleHours         int              `json:"staleHours"`
	MyApproved         []gh.PullRequest `json:"myApproved"`
	MyDraft            []gh.PullRequest `json:"myDraft"`
	MyOpenRaw          []gh.PullRequest `json:"myOpenRaw"`
	ReviewDirect       []gh.PullRequest `json:"reviewDirect"`
	ReviewTeam         []gh.PullRequest `json:"reviewTeam"`
	ApprovedMe         []gh.PullRequest `json:"approvedMe"`
	ApprovedTeam       []gh.PullRequest `json:"approvedTeam"`
	MyChangesReq       []gh.PullRequest `json:"myChangesReq"`
	ReviewChangesReq   []gh.PullRequest `json:"reviewChangesReq"`
}

// NewData builds the DashboardData from classified results.
func NewData(result *classify.Result, user string, teams, extraBots []string, autoRefreshMinutes, staleHours int) *Data {
	return &Data{
		User:               user,
		Generated:          time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		Teams:              nonNilStrings(teams),
		ExtraBotUsernames:  nonNilStrings(extraBots),
		AutoRefreshMinutes: autoRefreshMinutes,
		StaleHours:         staleHours,
		MyApproved:         result.MyApproved,
		MyDraft:            result.MyDraft,
		MyOpenRaw:          result.MyOpenRaw,
		ReviewDirect:       result.ReviewDirect,
		ReviewTeam:         result.ReviewTeam,
		ApprovedMe:         result.ApprovedMe,
		ApprovedTeam:       result.ApprovedTeam,
		MyChangesReq:       result.MyChangesReq,
		ReviewChangesReq:   result.ReviewChangesReq,
	}
}

type templateData struct {
	CSS      string
	JS       string
	DataJSON string
}

// Generate produces the complete self-contained HTML dashboard.
func Generate(data *Data) (string, error) {
	// Marshal the data JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Escape </ to <\/ to prevent </script> injection in PR titles
	jsonStr := strings.ReplaceAll(string(jsonBytes), "</", `<\/`)

	// Load embedded template files
	css, err := templateFS.ReadFile("templates/style.css")
	if err != nil {
		return "", err
	}
	js, err := templateFS.ReadFile("templates/script.js")
	if err != nil {
		return "", err
	}
	tmplStr, err := templateFS.ReadFile("templates/dashboard.html.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("dashboard").Parse(string(tmplStr))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData{
		CSS:      string(css),
		JS:       string(js),
		DataJSON: jsonStr,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func nonNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
