package core

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/berlinguyinca/priorart/internal/backends"
	"github.com/berlinguyinca/priorart/internal/candidate"
	"github.com/berlinguyinca/priorart/internal/config"
)

type fake struct {
	name  string
	calls int32
	cands []candidate.Candidate
	err   error
}

func (f *fake) Name() string { return f.name }
func (f *fake) Search(ctx context.Context, q, e string, l int) ([]candidate.Candidate, error) {
	atomic.AddInt32(&f.calls, 1)
	return f.cands, f.err
}

var _ backends.Backend = (*fake)(nil)

func TestPriorArtMergesAndRecordsSkip(t *testing.T) {
	ok := &fake{name: "ok", cands: []candidate.Candidate{{Source: "oss", Name: "lib"}}}
	bad := &fake{name: "bad", err: errors.New("boom")}
	c := New(config.Default(), []backends.Backend{ok, bad})
	c.SetLedgerPath(filepath.Join(t.TempDir(), "d.jsonl"))
	c.CacheTTL = 0

	res, err := c.PriorArt(context.Background(), "thing", "go", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Candidates) != 1 || res.Candidates[0].Name != "lib" {
		t.Fatalf("candidates: %+v", res.Candidates)
	}
	if len(res.Skipped) != 1 {
		t.Fatalf("expected one skipped, got %v", res.Skipped)
	}
	if len(res.Searched) != 1 || res.Searched[0] != "ok" {
		t.Fatalf("searched: %v", res.Searched)
	}
}

func TestPriorArtReturnsPriorDecision(t *testing.T) {
	c := New(config.Default(), nil)
	c.SetLedgerPath(filepath.Join(t.TempDir(), "d.jsonl"))
	c.CacheTTL = 0
	_ = c.RecordDecision(candidate.Decision{Query: "retry", Scope: "go", Action: "adopt", Choice: "avast/retry-go", Reason: "r"})
	res, _ := c.PriorArt(context.Background(), "retry", "go", 5)
	if res.Decision == nil || res.Decision.Choice != "avast/retry-go" {
		t.Fatalf("expected prior decision, got %+v", res.Decision)
	}
}

func TestCacheAvoidsSecondCall(t *testing.T) {
	ok := &fake{name: "ok", cands: []candidate.Candidate{{Source: "oss", Name: "lib"}}}
	c := New(config.Default(), []backends.Backend{ok})
	c.SetLedgerPath(filepath.Join(t.TempDir(), "d.jsonl"))
	c.CacheTTL = 3600
	_, _ = c.PriorArt(context.Background(), "same query", "go", 5)
	_, _ = c.PriorArt(context.Background(), "same query", "go", 5)
	if got := atomic.LoadInt32(&ok.calls); got != 1 {
		t.Fatalf("expected backend called once (cache), got %d", got)
	}
}
