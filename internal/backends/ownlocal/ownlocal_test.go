package ownlocal

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestTokens(t *testing.T) {
	got := Tokens("retry with exponential backoff")
	if len(got) != 3 {
		t.Fatalf("tokens: %v", got)
	}
	for _, w := range []string{"retry", "exponential", "backoff"} {
		found := false
		for _, g := range got {
			if g == w {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing %s in %v", w, got)
		}
	}
}

func TestSearchFindsOwnFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "sub/retry.go", "func retryWithBackoff(n int) {}\n")
	c := New([]string{dir}, "rg") // rg is on PATH in this env
	got, err := c.Search(context.Background(), "retry backoff", "go", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatal("expected at least one own hit")
	}
	if got[0].Source != "own" {
		t.Fatalf("expected source=own, got %+v", got[0])
	}
}

func TestGoFallbackWhenRgMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.py", "def backoff_schedule():\n    return 1\n")
	c := New([]string{dir}, "/nonexistent/rg") // force fallback
	got, err := c.Search(context.Background(), "backoff schedule", "py", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || got[0].Name == "" {
		t.Fatalf("fallback found nothing: %+v", got)
	}
}
