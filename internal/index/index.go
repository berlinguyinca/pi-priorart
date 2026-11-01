// Package index builds a deterministic per-repo capability index.
package index

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type RepoEntry struct {
	Repo    string   `json:"repo"`
	Title   string   `json:"title,omitempty"`
	Summary string   `json:"summary,omitempty"`
	Files   []string `json:"files"`
}

const maxFiles = 200

// Build indexes every immediate subdirectory of root that looks like a repo.
func Build(root, outDir string) ([]RepoEntry, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	dirs, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []RepoEntry
	for _, d := range dirs {
		if !d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			continue
		}
		e := buildOne(d.Name(), filepath.Join(root, d.Name()))
		b, _ := json.Marshal(e)
		_ = os.WriteFile(filepath.Join(outDir, d.Name()+".json"), b, 0o644)
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Repo < out[j].Repo })
	return out, nil
}

func buildOne(repo, repoPath string) RepoEntry {
	e := RepoEntry{Repo: repo}
	if readme := readReadme(repoPath); readme != "" {
		e.Title, e.Summary = parseReadme(readme)
	}
	var files []string
	_ = filepath.WalkDir(repoPath, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor", "dist", "build", ".venv", "venv":
				return filepath.SkipDir
			}
			return nil
		}
		if len(files) < maxFiles && isSource(d.Name()) {
			rel, _ := filepath.Rel(repoPath, p)
			files = append(files, rel)
		}
		return nil
	})
	sort.Strings(files)
	e.Files = files
	return e
}

func readReadme(repoPath string) string {
	for _, n := range []string{"README.md", "README.rst", "README"} {
		if b, err := os.ReadFile(filepath.Join(repoPath, n)); err == nil {
			return string(b)
		}
	}
	return ""
}

func parseReadme(s string) (title, summary string) {
	sc := bufio.NewScanner(strings.NewReader(s))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if title == "" && strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			continue
		}
		if title != "" && summary == "" && line != "" && !strings.HasPrefix(line, "#") {
			summary = line
			break
		}
	}
	return
}

func isSource(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".go", ".py", ".ts", ".js", ".mjs", ".rs", ".java", ".scala", ".rb", ".c", ".cpp", ".cs", ".sh":
		return true
	}
	return false
}

// Load reads a previously built entry.
func Load(outDir, repo string) (RepoEntry, bool) {
	var e RepoEntry
	if b, err := os.ReadFile(filepath.Join(outDir, repo+".json")); err == nil {
		_ = json.Unmarshal(b, &e)
		return e, true
	}
	return e, false
}
