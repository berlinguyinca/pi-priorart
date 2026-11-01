package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/berlinguyinca/priorart/internal/backends"
	"github.com/berlinguyinca/priorart/internal/candidate"
	"github.com/berlinguyinca/priorart/internal/config"
	"github.com/berlinguyinca/priorart/internal/core"
)

type fb struct{ n string }

func (f fb) Name() string { return f.n }
func (f fb) Search(ctx context.Context, q, e string, l int) ([]candidate.Candidate, error) {
	return []candidate.Candidate{{Source: "oss", Name: "hit-" + f.n}}, nil
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	c := core.New(config.Default(), []backends.Backend{fb{n: "b1"}})
	c.SetLedgerPath(t.TempDir() + "/d.jsonl")
	c.CacheTTL = 0
	return New(c)
}

func TestToolsList(t *testing.T) {
	res, rpcErr := newTestServer(t).Dispatch("tools/list", map[string]any{})
	if rpcErr != nil {
		t.Fatal(rpcErr)
	}
	names := res.(map[string]any)["tools"].([]toolDef)
	if len(names) != 5 {
		t.Fatalf("expected 5 tools, got %d", len(names))
	}
}

func TestToolsCallPriorArtReturnsJSONText(t *testing.T) {
	res, rpcErr := newTestServer(t).Dispatch("tools/call", map[string]any{
		"name":      "prior_art",
		"arguments": map[string]any{"query": "retry", "ecosystem": "go"},
	})
	if rpcErr != nil {
		t.Fatal(rpcErr)
	}
	content := res.(map[string]any)["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("content text must be JSON: %q err=%v", text, err)
	}
	if _, ok := parsed["query"]; !ok {
		t.Fatalf("expected query in result: %s", text)
	}
}

func TestToolsCallRecordDecision(t *testing.T) {
	s := newTestServer(t)
	if _, rpcErr := s.Dispatch("tools/call", map[string]any{
		"name":      "record_decision",
		"arguments": map[string]any{"query": "retry", "action": "adopt", "choice": "x", "reason": "r"},
	}); rpcErr != nil {
		t.Fatal(rpcErr)
	}
}
