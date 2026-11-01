package ledger

import (
	"path/filepath"
	"testing"

	"github.com/berlinguyinca/priorart/internal/candidate"
)

func TestAppendAndLookupLastWriteWins(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "decisions.jsonl")
	k := Normalize("retry with backoff", "go")
	if err := Append(p, candidate.Decision{Key: k, Query: "retry with backoff", Scope: "go", Action: "build", Choice: "own", Reason: "v1"}); err != nil {
		t.Fatal(err)
	}
	if err := Append(p, candidate.Decision{Key: k, Query: "retry with backoff", Scope: "go", Action: "adopt", Choice: "avast/retry-go", Reason: "v2"}); err != nil {
		t.Fatal(err)
	}
	d, ok := Lookup(p, k)
	if !ok || d.Action != "adopt" || d.Choice != "avast/retry-go" {
		t.Fatalf("expected last-write-wins adopt, got ok=%v d=%+v", ok, d)
	}
	if _, ok := Lookup(p, "missing-key"); ok {
		t.Fatal("expected no hit for missing key")
	}
}

func TestNormalizeStable(t *testing.T) {
	if Normalize("Retry  With Backoff", "go") != Normalize("retry with backoff", "go") {
		t.Fatal("normalize must be case/whitespace-insensitive")
	}
	if Normalize("x", "go") == Normalize("x", "py") {
		t.Fatal("scope must affect the key")
	}
}
