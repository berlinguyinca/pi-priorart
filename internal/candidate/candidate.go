// Package candidate defines the shared data types returned by priorart.
package candidate

// Candidate is a single prior-art hit from any backend.
type Candidate struct {
	Source      string   `json:"source"` // "own" | "oss"
	Ecosystem   string   `json:"ecosystem"`
	Name        string   `json:"name"`
	Version     string   `json:"version,omitempty"`
	RepoURL     string   `json:"repo_url,omitempty"`
	License     string   `json:"license,omitempty"`
	Activity    Activity `json:"activity,omitempty"`
	Snippet     string   `json:"snippet,omitempty"`
	MatchReason string   `json:"match_reason,omitempty"`
	DocsURL     string   `json:"docs_url,omitempty"`
}

type Activity struct {
	LastRelease string `json:"last_release,omitempty"`
	Stars       int    `json:"stars,omitempty"`
}

// PriorArtResult is the return of the prior_art tool.
type PriorArtResult struct {
	Query      string      `json:"query"`
	Decision   *Decision   `json:"prior_decision,omitempty"`
	Candidates []Candidate `json:"candidates"`
	Searched   []string    `json:"searched"`
	Skipped    []string    `json:"skipped,omitempty"`
}

// Decision is a recorded adopt-vs-build choice (ledger line).
type Decision struct {
	Key                  string   `json:"key"`
	Query                string   `json:"query"`
	Scope                string   `json:"scope,omitempty"`
	Action               string   `json:"action"` // "adopt" | "build"
	Choice               string   `json:"choice"`
	Reason               string   `json:"reason"`
	CandidatesConsidered []string `json:"candidates_considered,omitempty"`
	TS                   string   `json:"ts"`
	Machine              string   `json:"machine,omitempty"`
}

// Wheel is a "known wheel" that must not be reimplemented.
type Wheel struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Canonical   []string `json:"canonical"`
}
