// Package ownlocal searches the user's own repos (ripgrep, with a Go fallback).
package ownlocal

import (
	"bufio"
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/berlinguyinca/priorart/internal/candidate"
)

type Client struct {
	Roots  []string
	RGPath string
}

func New(roots []string, rgPath string) *Client {
	if rgPath == "" {
		rgPath = "rg"
	}
	return &Client{Roots: roots, RGPath: rgPath}
}

func (c *Client) Name() string { return "own-local" }

// min length 5 drops common 4-char stopwords (with/that/this) that add search noise.
var tokenRe = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9_]{4,}`)

// Tokens extracts search tokens (alphanumeric, len>=4) from a query, lowercased, deduped.
func Tokens(query string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range tokenRe.FindAllString(strings.ToLower(query), -1) {
		if !seen[m] {
			seen[m] = true
			out = append(out, m)
		}
	}
	return out
}

const snippetMax = 300

func clip(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > snippetMax {
		return s[:snippetMax]
	}
	return s
}

type hit struct {
	path    string
	snippet string
	token   string
}

func (c *Client) Search(ctx context.Context, query, _ string, limit int) ([]candidate.Candidate, error) {
	if limit <= 0 {
		limit = 10
	}
	toks := Tokens(query)
	if len(toks) == 0 {
		return nil, nil
	}
	_, lookErr := exec.LookPath(c.RGPath)
	rgOK := lookErr == nil
	var hits []hit
	seen := map[string]bool{}
	add := func(h hit) {
		if !seen[h.path] {
			seen[h.path] = true
			hits = append(hits, h)
		}
	}
	for _, root := range c.Roots {
		if _, err := os.Stat(root); err != nil {
			continue
		}
		if rgOK {
			for _, h := range c.ripgrep(ctx, root, toks, limit) {
				add(h)
			}
			continue
		}
		for _, h := range c.walk(root, toks, limit) {
			add(h)
		}
	}
	out := make([]candidate.Candidate, 0, len(hits))
	for _, h := range hits {
		if len(out) >= limit {
			break
		}
		out = append(out, candidate.Candidate{
			Source: "own", Name: "own:" + h.path, Snippet: clip(h.snippet),
			MatchReason: "own repo match (token " + h.token + ")",
		})
	}
	return out, nil
}

func (c *Client) ripgrep(ctx context.Context, root string, toks []string, limit int) []hit {
	var out []hit
	for _, tok := range toks {
		args := []string{"-n", "--no-heading", "-F", "-m", "1", tok, root}
		cmd := exec.CommandContext(ctx, c.RGPath, args...)
		b, err := cmd.Output()
		if err != nil {
			continue // rg exits 1 on no match; treat as empty
		}
		sc := bufio.NewScanner(strings.NewReader(string(b)))
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			line := sc.Text()
			parts := strings.SplitN(line, ":", 3)
			if len(parts) < 3 {
				continue
			}
			out = append(out, hit{path: parts[0], snippet: parts[2], token: tok})
			if len(out) >= limit*len(toks) {
				return out
			}
		}
	}
	return out
}

var skipDir = map[string]bool{".git": true, "node_modules": true, "vendor": true, "dist": true, "build": true, ".venv": true, "venv": true}

func (c *Client) walk(root string, toks []string, limit int) []hit {
	var out []hit
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if skipDir[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			line := sc.Text()
			low := strings.ToLower(line)
			for _, tok := range toks {
				if strings.Contains(low, tok) {
					out = append(out, hit{path: p, snippet: line, token: tok})
					if len(out) >= limit*len(toks) {
						return nil
					}
					break
				}
			}
		}
		return nil
	})
	return out
}
