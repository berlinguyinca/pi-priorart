// Package rank merges, dedups, and scores prior-art candidates.
package rank

import (
	"sort"
	"strings"

	"github.com/berlinguyinca/priorart/internal/candidate"
	"github.com/berlinguyinca/priorart/internal/config"
	"github.com/berlinguyinca/priorart/internal/wheels"
)

// Rank dedups by Name, drops blocked licenses, scores, and sorts descending.
func Rank(cands []candidate.Candidate, query string, cfg config.Config) []candidate.Candidate {
	block := map[string]bool{}
	for _, l := range cfg.BlockLicenses {
		block[strings.ToLower(l)] = true
	}
	best := map[string]candidate.Candidate{}
	for _, c := range cands {
		if c.Name == "" {
			continue
		}
		if c.License != "" && block[strings.ToLower(c.License)] {
			continue
		}
		if prev, ok := best[c.Name]; !ok || score(c, query) > score(prev, query) {
			best[c.Name] = c
		}
	}
	out := make([]candidate.Candidate, 0, len(best))
	for _, c := range best {
		out = append(out, c)
	}
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := score(out[i], query), score(out[j], query)
		if si != sj {
			return si > sj
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func score(c candidate.Candidate, query string) float64 {
	s := 0.0
	if c.Source == "own" {
		s += 3.0
	}
	if q := strings.ToLower(query); q != "" && strings.Contains(strings.ToLower(c.Name), q) {
		s += 1.0
	}
	for _, w := range wheels.All() {
		for _, canon := range w.Canonical {
			if strings.EqualFold(canon, c.Name) {
				s += 5.0
			}
		}
	}
	if c.Activity.Stars > 0 {
		s += 1.0
	}
	return s
}
