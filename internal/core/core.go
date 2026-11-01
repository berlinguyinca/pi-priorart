// Package core orchestrates ledger lookup, backend fan-out, ranking, and caching.
package core

import (
	"context"
	"sync"
	"time"

	"github.com/berlinguyinca/priorart/internal/backends"
	"github.com/berlinguyinca/priorart/internal/candidate"
	"github.com/berlinguyinca/priorart/internal/config"
	"github.com/berlinguyinca/priorart/internal/ledger"
	"github.com/berlinguyinca/priorart/internal/rank"
)

type cacheEntry struct {
	cats []candidate.Candidate
	at   time.Time
}

type Core struct {
	cfg      config.Config
	backends []backends.Backend
	ledgerP  string
	mu       sync.Mutex
	cache    map[string]cacheEntry
	CacheTTL int
}

func New(cfg config.Config, bs []backends.Backend) *Core {
	return &Core{cfg: cfg, backends: bs, cache: map[string]cacheEntry{}, CacheTTL: cfg.CacheTTLSeconds}
}

func (c *Core) SetLedgerPath(p string) { c.ledgerP = p }

func (c *Core) PriorArt(ctx context.Context, query, ecosystem string, limit int) (candidate.PriorArtResult, error) {
	if limit <= 0 {
		limit = 5
	}
	res := candidate.PriorArtResult{Query: query, Searched: []string{}}
	key := ledger.Normalize(query, ecosystem)
	if d, ok := ledger.Lookup(c.ledgerP, key); ok {
		res.Decision = d
	}

	ckey := ecosystem + "|" + key
	if c.CacheTTL > 0 {
		c.mu.Lock()
		if ce, ok := c.cache[ckey]; ok && time.Since(ce.at) < time.Duration(c.CacheTTL)*time.Second {
			res.Candidates = rank.Rank(ce.cats, query, c.cfg)
			c.mu.Unlock()
			return res, nil
		}
		c.mu.Unlock()
	}

	var merged []candidate.Candidate
	for _, b := range c.backends {
		cats, err := b.Search(ctx, query, ecosystem, limit)
		if err != nil {
			res.Skipped = append(res.Skipped, b.Name()+": "+err.Error())
			continue
		}
		merged = append(merged, cats...)
		res.Searched = append(res.Searched, b.Name())
	}
	ranked := rank.Rank(merged, query, c.cfg)
	if c.CacheTTL > 0 {
		c.mu.Lock()
		c.cache[ckey] = cacheEntry{cats: ranked, at: time.Now()}
		c.mu.Unlock()
	}
	res.Candidates = ranked
	return res, nil
}

func (c *Core) SearchOwn(ctx context.Context, query string, limit int) ([]candidate.Candidate, error) {
	for _, b := range c.backends {
		if b.Name() == "own-local" {
			return b.Search(ctx, query, "", limit)
		}
	}
	return nil, nil
}

func (c *Core) RecordDecision(d candidate.Decision) error {
	return ledger.Append(c.ledgerP, d)
}
