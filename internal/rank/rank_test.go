package rank

import (
	"testing"

	"github.com/berlinguyinca/priorart/internal/candidate"
	"github.com/berlinguyinca/priorart/internal/config"
)

func TestOwnFavoredAndDedup(t *testing.T) {
	cands := []candidate.Candidate{
		{Source: "oss", Name: "libx", Ecosystem: "go"},
		{Source: "own", Name: "own:repo/libx.go", Ecosystem: "go"},
		{Source: "oss", Name: "libx", Ecosystem: "go"}, // dup of first
	}
	out := Rank(cands, "x", config.Default())
	if len(out) != 2 {
		t.Fatalf("expected dedup to 2, got %d: %+v", len(out), out)
	}
	if out[0].Source != "own" {
		t.Fatalf("own should rank first: %+v", out[0])
	}
}

func TestBlockedLicenseFiltered(t *testing.T) {
	cfg := config.Default()
	cfg.BlockLicenses = []string{"GPL-3.0"}
	out := Rank([]candidate.Candidate{
		{Source: "oss", Name: "a", License: "MIT"},
		{Source: "oss", Name: "b", License: "GPL-3.0"},
	}, "x", cfg)
	if len(out) != 1 || out[0].Name != "a" {
		t.Fatalf("expected only MIT to survive: %+v", out)
	}
}
