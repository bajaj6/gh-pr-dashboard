package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/browser"
	flag "github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"

	"github.com/bajaj6/gh-pr-dashboard/internal/classify"
	"github.com/bajaj6/gh-pr-dashboard/internal/config"
	"github.com/bajaj6/gh-pr-dashboard/internal/dashboard"
	gh "github.com/bajaj6/gh-pr-dashboard/internal/github"
	"github.com/bajaj6/gh-pr-dashboard/internal/logging"
	"github.com/bajaj6/gh-pr-dashboard/internal/server"
)

const version = "2.0.0"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	// CLI flags
	var (
		cliOrgs     []string
		cliUser     string
		cliOutput   string
		cliLimit    int
		cliPort     int
		cliConfig   string
		noOpen      bool
		serveMode   bool
		stdoutMode  bool
		noProgress  bool
		initMode    bool
		showHelp    bool
		showVersion bool
	)

	flag.StringSliceVar(&cliOrgs, "org", nil, "GitHub org to scan (repeatable)")
	flag.StringVar(&cliUser, "user", "", "Override GitHub username")
	flag.StringVar(&cliOutput, "output", "", "Output HTML file")
	flag.IntVar(&cliLimit, "limit", 0, "Max PRs per search query")
	flag.IntVar(&cliPort, "port", 0, "Port for serve mode")
	flag.StringVar(&cliConfig, "config", "", "Path to config file")
	flag.BoolVar(&noOpen, "no-open", false, "Don't auto-open browser")
	flag.BoolVar(&serveMode, "serve", false, "Start local server")
	flag.BoolVar(&stdoutMode, "stdout", false, "Output HTML to stdout")
	flag.BoolVar(&noProgress, "no-progress", false, "Suppress progress messages")
	flag.BoolVar(&initMode, "init", false, "Create config file")
	flag.BoolVarP(&showHelp, "help", "h", false, "Show help")
	flag.BoolVarP(&showVersion, "version", "v", false, "Show version")

	// Custom usage
	flag.Usage = func() {
		fmt.Print(`gh-pr-dashboard — GitHub PR review dashboard

Usage:
  gh pr-dashboard [OPTIONS]
  gh pr-dashboard --serve [OPTIONS]
  gh pr-dashboard --init

Options:
  --org ORG        GitHub org to scan (repeatable)
  --user USER      Override GitHub username (auto-detected by default)
  --output FILE    Output HTML file (default: dashboard.html)
  --limit N        Max PRs per search query (default: 100)
  --port PORT      Port for serve mode (default: 8787)
  --no-open        Don't auto-open browser
  --serve          Start local server (refreshes on reload)
  --config FILE    Path to config file
  --init           Create config file at ~/.config/gh-pr-dashboard/config.json
  --help           Show this help
  --version        Show version

Config file:
  ~/.config/gh-pr-dashboard/config.json

  {
    "orgs": ["my-org", "other-org"],
    "ignored_teams": ["my-org/noisy-team"],
    "extra_bot_usernames": ["some-bot"],
    "output_file": "dashboard.html",
    "search_limit": 100
  }

Examples:
  gh pr-dashboard --org my-org                  # quick one-time run
  gh pr-dashboard --init                        # create blank config file, then edit it
  gh pr-dashboard                               # use config file
  gh pr-dashboard --serve                       # live dashboard at localhost:8787
  gh pr-dashboard --org a --org b --no-open     # multiple orgs, don't open browser
`)
	}

	flag.Parse()

	if showHelp {
		flag.Usage()
		return nil
	}
	if showVersion {
		fmt.Printf("gh-pr-dashboard %s\n", version)
		return nil
	}

	if initMode {
		return config.Init()
	}

	// Load config
	confPath := cliConfig
	if confPath == "" {
		confPath = config.DefaultConfigPath()
	}
	cfg, err := config.Load(confPath)
	if err != nil {
		return err
	}

	opts := config.Merge(cfg, cliOrgs, cliUser, cliOutput, cliLimit)
	if cliPort > 0 {
		opts.Port = cliPort
	}
	opts.NoOpen = noOpen
	opts.ServeMode = serveMode
	opts.StdoutMode = stdoutMode
	opts.NoProgress = noProgress || stdoutMode

	log := &logging.Logger{Quiet: opts.NoProgress}

	// Create GitHub client
	client, err := gh.NewClient()
	if err != nil {
		return err
	}

	// Detect user
	if opts.User == "" {
		u, err := client.FetchUser()
		if err != nil {
			return err
		}
		opts.User = u
	}

	log.Logf("=== GitHub PR Dashboard ===")
	log.Logf("User: @%s", opts.User)
	if len(opts.Orgs) > 0 {
		log.Logf("Orgs: %s", strings.Join(opts.Orgs, " "))
	}

	if opts.ServeMode {
		// In serve mode, suppress per-request logging — only show startup info
		quietLog := &logging.Logger{Quiet: true}
		generateFn := func() (string, error) {
			return generate(client, opts, quietLog)
		}
		return server.Serve(opts.Port, opts.NoOpen, log, generateFn)
	}

	// Generate function (single-run mode)
	generateFn := func() (string, error) {
		return generate(client, opts, log)
	}

	html, err := generateFn()
	if err != nil {
		return err
	}

	if opts.StdoutMode {
		fmt.Print(html)
		return nil
	}

	if err := os.WriteFile(opts.OutputFile, []byte(html), 0644); err != nil {
		return err
	}
	log.Logf("Dashboard written to: %s", opts.OutputFile)

	if !opts.NoOpen {
		b := browser.New("", os.Stdout, os.Stderr)
		if err := b.Browse(opts.OutputFile); err != nil {
			log.Logf("  (could not open browser: %s)", err)
		}
	}

	return nil
}

func generate(client *gh.Client, opts *config.Options, log *logging.Logger) (string, error) {
	createdAfter := time.Now().UTC().AddDate(0, -1, 0).Format("2006-01-02")
	log.Logf("")
	log.Logf("Filtering PRs created after: %s", createdAfter)

	// Build query base
	var queryBase string
	if len(opts.Orgs) > 0 {
		orgParts := make([]string, len(opts.Orgs))
		for i, o := range opts.Orgs {
			orgParts[i] = "org:" + o
		}
		queryBase = fmt.Sprintf("is:pr is:open %s created:>%s", strings.Join(orgParts, " "), createdAfter)
	} else {
		queryBase = fmt.Sprintf("is:pr is:open created:>%s", createdAfter)
	}

	var (
		gqlResult *gh.CombinedSearchResponse
		apiTeams  []string
		notifPRs  []gh.PullRequest
	)

	log.Logf("")
	log.Logf("Fetching PR data + team memberships + notifications (in parallel)...")

	g := new(errgroup.Group)

	g.Go(func() error {
		var err error
		gqlResult, err = client.FetchAllPRs(queryBase, opts.User, opts.SearchLimit)
		return err
	})

	g.Go(func() error {
		notifPRs = client.FetchNotificationPRs(opts.Orgs, opts.User, createdAfter)
		return nil
	})

	g.Go(func() error {
		var err error
		apiTeams, err = client.FetchTeamsFromAPI(opts.Orgs)
		return err
	})

	if err := g.Wait(); err != nil {
		return "", err
	}

	// Filter ignored teams
	teams, skipped := gh.FilterIgnoredTeams(apiTeams, opts.IgnoredTeams)
	for _, t := range skipped {
		log.Logf("  Ignoring team: %s", t)
	}

	if len(teams) == 0 {
		log.Logf("  Warning: No teams found in %s orgs.", strings.Join(opts.Orgs, ", "))
	} else {
		log.Logf("  Teams found:")
		for _, t := range teams {
			log.Logf("    - %s", t)
		}
	}

	// Classify PRs
	result, counts := classify.Classify(gqlResult)

	log.Logf("  My PRs — Approved:    %d", counts.MyApproved)
	log.Logf("  My PRs — Draft:       %d", counts.MyDraft)
	log.Logf("  My PRs — Open:        %d", counts.MyOpen)
	log.Logf("  My PRs — Changes Req: %d", counts.MyChangesReq)
	log.Logf("  Review requested:     %d", counts.ReviewAll)
	log.Logf("  Review — Changes Req: %d", counts.ReviewChangesReq)
	log.Logf("  Approved (I reviewed): %d", counts.ApprovedMe)
	log.Logf("  Approved (requested):  %d", counts.ApprovedReq)

	// Merge notification-discovered PRs
	newCount := classify.MergeNotifications(result, notifPRs, opts.IgnoredTeams)
	if newCount > 0 {
		log.Logf("  Notifications: recovered %d team-approved PR(s)", newCount)
	}

	// Discover additional teams from results
	for _, dteam := range result.DiscoveredTeams {
		found := false
		for _, t := range teams {
			if t == dteam {
				found = true
				break
			}
		}
		if !found {
			teams = append(teams, dteam)
			log.Logf("  Discovered additional team: %s", dteam)
		}
	}

	log.Logf("")
	log.Logf("Generating dashboard...")

	// Build dashboard data and generate HTML
	data := dashboard.NewData(result, opts.User, teams, opts.ExtraBotUsernames, opts.AutoRefreshMinutes, opts.StaleHours)
	return dashboard.Generate(data)
}
