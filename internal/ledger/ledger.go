// Package ledger stores adopt-vs-build decisions as synced JSONL.
package ledger

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/berlinguyinca/priorart/internal/candidate"
)

func Normalize(query, scope string) string {
	s := strings.ToLower(strings.TrimSpace(query))
	s = strings.Join(strings.Fields(s), " ")
	if scope != "" {
		s = scope + "|" + s
	}
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func Lookup(path, key string) (*candidate.Decision, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var best *candidate.Decision
	for {
		var d candidate.Decision
		if err := dec.Decode(&d); err != nil {
			break
		}
		if d.Key == key {
			cp := d
			best = &cp
		}
	}
	if best != nil {
		return best, true
	}
	return nil, false
}

func Append(path string, d candidate.Decision) error {
	if d.Key == "" {
		d.Key = Normalize(d.Query, d.Scope)
	}
	if d.TS == "" {
		d.TS = time.Now().UTC().Format(time.RFC3339)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}
