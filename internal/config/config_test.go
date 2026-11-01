package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultsAndEnvSecrets(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "priorart.json")
	body := `{"own_roots":["/tmp/x"],"github":{"own_account":"berlinguyinca","use_own_layer":false},"github_token":"SHOULD_BE_IGNORED"}`
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GITHUB_TOKEN", "ghp_from_env")
	t.Setenv("CONTEXT7_API_KEY", "ctx7_from_env")

	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.OwnRoots) != 1 || cfg.OwnRoots[0] != "/tmp/x" {
		t.Fatalf("own_roots not loaded: %+v", cfg.OwnRoots)
	}
	if cfg.GitHub.OwnAccount != "berlinguyinca" {
		t.Fatalf("github not loaded: %+v", cfg.GitHub)
	}
	if cfg.GitHubToken != "ghp_from_env" {
		t.Fatalf("secret must come from env, got %q", cfg.GitHubToken)
	}
	if cfg.Context7APIKey != "ctx7_from_env" {
		t.Fatalf("context7 key must come from env")
	}
	// the file's bogus github_token must NOT have leaked in
	if cfg.GitHubToken == "SHOULD_BE_IGNORED" {
		t.Fatal("secret leaked from file")
	}
}

func TestLoadMissingFileUsesDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Ecosystems) == 0 || cfg.CacheTTLSeconds == 0 {
		t.Fatalf("defaults not applied: %+v", cfg)
	}
}
