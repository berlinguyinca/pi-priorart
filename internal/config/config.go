// Package config loads priorart configuration. Secrets come from env only.
package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
)

type GitHub struct {
	OwnAccount  string `json:"own_account"`
	UseOwnLayer bool   `json:"use_own_layer"`
}

type Context7 struct {
	Enabled bool `json:"enabled"`
}

type Config struct {
	OwnRoots        []string `json:"own_roots"`
	OwnSSH          []string `json:"own_ssh"`
	Ecosystems      []string `json:"ecosystems"`
	GitHub          GitHub   `json:"github"`
	LedgerPath      string   `json:"ledger_path"`
	Context7        Context7 `json:"context7"`
	BlockLicenses   []string `json:"block_licenses"`
	CacheTTLSeconds int      `json:"cache_ttl_seconds"`

	GitHubToken    string `json:"-"` // env: GITHUB_TOKEN
	Context7APIKey string `json:"-"` // env: CONTEXT7_API_KEY
}

func Default() Config {
	return Config{
		Ecosystems:      []string{"py", "js", "rs", "go"},
		BlockLicenses:   []string{"GPL-3.0", "AGPL-3.0"},
		CacheTTLSeconds: 3600,
		Context7:        Context7{Enabled: true},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return cfg, nil // defaults + env only
			}
			return cfg, err
		}
		if err := json.Unmarshal(b, &cfg); err != nil {
			return cfg, err
		}
	}
	cfg.GitHubToken = os.Getenv("GITHUB_TOKEN")
	cfg.Context7APIKey = os.Getenv("CONTEXT7_API_KEY")
	return cfg, nil
}
