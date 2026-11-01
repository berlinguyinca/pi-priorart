package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCapturesReadmeAndFiles(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "mytool")
	_ = os.MkdirAll(filepath.Join(repo, "cmd"), 0o755)
	_ = os.WriteFile(filepath.Join(repo, "README.md"), []byte("# My Tool\n\nDoes the thing.\n"), 0o644)
	_ = os.WriteFile(filepath.Join(repo, "main.go"), []byte("package main\n"), 0o644)

	out := t.TempDir()
	entries, err := Build(root, out)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(entries))
	}
	e := entries[0]
	if e.Repo != "mytool" || e.Title != "My Tool" {
		t.Fatalf("bad entry: %+v", e)
	}
	found := false
	for _, f := range e.Files {
		if f == "main.go" {
			found = true
		}
	}
	if !found {
		t.Fatalf("main.go not indexed: %+v", e.Files)
	}
}
