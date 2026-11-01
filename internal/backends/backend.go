// Package backends contains pluggable prior-art search sources.
package backends

import (
	"context"

	"github.com/berlinguyinca/priorart/internal/candidate"
)

// Backend searches one source for prior art matching a plain-English query.
type Backend interface {
	Name() string
	Search(ctx context.Context, query, ecosystem string, limit int) ([]candidate.Candidate, error)
}
