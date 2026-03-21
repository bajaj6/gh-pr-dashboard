package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Orgs               []string `json:"orgs"`
	IgnoredTeams       []string `json:"ignored_teams"`
	ExtraBotUsernames  []string `json:"extra_bot_usernames"`
	OutputFile         string   `json:"output_file"`
	SearchLimit        int      `json:"search_limit"`
	AutoRefreshMinutes *int     `json:"auto_refresh_minutes,omitempty"`
	StaleHours         *int     `json:"stale_hours,omitempty"`
}

type Options struct {
	Orgs               []string
	IgnoredTeams       []string
	ExtraBotUsernames  []string
	OutputFile         string
	SearchLimit        int
	AutoRefreshMinutes int
	StaleHours         int
	User               string
	Port               int
	NoOpen             bool
	ServeMode          bool
	StdoutMode         bool
	NoProgress         bool
	ConfigPath         string
}

func DefaultConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "gh-pr-dashboard")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "gh-pr-dashboard")
}

func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.json")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", path, err)
	}
	return &cfg, nil
}

func Merge(cfg *Config, cliOrgs []string, cliUser, cliOutput string, cliLimit int) *Options {
	opts := &Options{}

	// Orgs: CLI > config
	if len(cliOrgs) > 0 {
		opts.Orgs = cliOrgs
	} else if len(cfg.Orgs) > 0 {
		opts.Orgs = cfg.Orgs
	}

	opts.IgnoredTeams = cfg.IgnoredTeams
	opts.ExtraBotUsernames = cfg.ExtraBotUsernames

	// Output: CLI > config > default
	switch {
	case cliOutput != "":
		opts.OutputFile = cliOutput
	case cfg.OutputFile != "":
		opts.OutputFile = cfg.OutputFile
	default:
		opts.OutputFile = "dashboard.html"
	}

	// Limit: CLI > config > default
	switch {
	case cliLimit > 0:
		opts.SearchLimit = cliLimit
	case cfg.SearchLimit > 0:
		opts.SearchLimit = cfg.SearchLimit
	default:
		opts.SearchLimit = 100
	}

	if cliUser != "" {
		opts.User = cliUser
	}

	// Auto-refresh: config > default (15 minutes)
	if cfg.AutoRefreshMinutes != nil {
		opts.AutoRefreshMinutes = *cfg.AutoRefreshMinutes
	} else {
		opts.AutoRefreshMinutes = 15
	}

	// Stale hours: config > default (18 hours)
	if cfg.StaleHours != nil {
		opts.StaleHours = *cfg.StaleHours
	} else {
		opts.StaleHours = 24
	}

	// Port default
	opts.Port = 8787

	return opts
}

func Init() error {
	dir := DefaultConfigDir()
	path := DefaultConfigPath()

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Config already exists: %s\n", path)
		fmt.Println("Edit it with: $EDITOR", path)
		return nil
	}

	cfg := Config{
		Orgs:              []string{},
		IgnoredTeams:      []string{},
		ExtraBotUsernames: []string{},
		OutputFile:        "dashboard.html",
		SearchLimit:       100,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	fmt.Printf("Created config: %s\n", path)
	fmt.Printf("Edit it with: $EDITOR %s\n", path)
	fmt.Println()
	fmt.Println(`Add your GitHub orgs to the "orgs" array, then run: gh pr-dashboard`)
	return nil
}
