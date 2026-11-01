// Package mcp implements a minimal MCP stdio server (JSON-RPC 2.0).
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/berlinguyinca/priorart/internal/candidate"
	"github.com/berlinguyinca/priorart/internal/core"
	"github.com/berlinguyinca/priorart/internal/wheels"
)

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string { return fmt.Sprintf("rpc %d: %s", e.Code, e.Message) }

type Server struct{ core *core.Core }

func New(c *core.Core) *Server { return &Server{core: c} }

type toolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

func schema(props map[string]any, req ...string) map[string]any {
	return map[string]any{"type": "object", "properties": props, "required": req}
}

func (s *Server) tools() []toolDef {
	return []toolDef{
		{Name: "prior_art", Description: "Search own repos + public OSS for prior art before writing code.",
			InputSchema: schema(map[string]any{
				"query":     map[string]any{"type": "string"},
				"ecosystem": map[string]any{"type": "string"},
				"limit":     map[string]any{"type": "integer"},
			}, "query")},
		{Name: "search_own", Description: "Search only your own repos.",
			InputSchema: schema(map[string]any{"query": map[string]any{"type": "string"}, "limit": map[string]any{"type": "integer"}}, "query")},
		{Name: "search_registry", Description: "Search one package registry.",
			InputSchema: schema(map[string]any{
				"ecosystem": map[string]any{"type": "string", "enum": []string{"py", "js", "rs", "go"}},
				"query":     map[string]any{"type": "string"},
				"limit":     map[string]any{"type": "integer"},
			}, "ecosystem", "query")},
		{Name: "record_decision", Description: "Record an adopt-vs-build decision in the ledger.",
			InputSchema: schema(map[string]any{
				"query":  map[string]any{"type": "string"},
				"action": map[string]any{"type": "string", "enum": []string{"adopt", "build"}},
				"choice": map[string]any{"type": "string"},
				"reason": map[string]any{"type": "string"},
			}, "query", "action", "choice", "reason")},
		{Name: "wheels", Description: "List known wheels that must not be reimplemented.",
			InputSchema: schema(map[string]any{})},
	}
}

func (s *Server) Dispatch(method string, params map[string]any) (any, *RPCError) {
	switch method {
	case "initialize":
		return map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo":      map[string]any{"name": "priorart", "version": "0.1.0"},
			"capabilities":    map[string]any{"tools": map[string]any{}},
		}, nil
	case "tools/list":
		return map[string]any{"tools": s.tools()}, nil
	case "tools/call":
		return s.call(params)
	default:
		return nil, &RPCError{Code: -32601, Message: "method not found: " + method}
	}
}

func (s *Server) call(params map[string]any) (any, *RPCError) {
	name, _ := params["name"].(string)
	args, _ := params["arguments"].(map[string]any)
	ctx := context.Background()
	var payload any
	var err error
	switch name {
	case "prior_art":
		var r candidate.PriorArtResult
		r, err = s.core.PriorArt(ctx, str(args, "query"), str(args, "ecosystem"), intv(args, "limit", 5))
		payload = r
	case "search_own":
		payload, err = s.core.SearchOwn(ctx, str(args, "query"), intv(args, "limit", 10))
	case "search_registry":
		var r candidate.PriorArtResult
		r, err = s.core.PriorArt(ctx, str(args, "query"), str(args, "ecosystem"), intv(args, "limit", 5))
		payload = r
	case "record_decision":
		err = s.core.RecordDecision(candidate.Decision{
			Query: str(args, "query"), Scope: str(args, "ecosystem"),
			Action: str(args, "action"), Choice: str(args, "choice"), Reason: str(args, "reason"),
		})
		payload = map[string]any{"ok": true}
	case "wheels":
		payload = wheels.All()
	default:
		return nil, &RPCError{Code: -32602, Message: "unknown tool: " + name}
	}
	if err != nil {
		return contentResult(fmt.Sprintf("priorart error: %v", err)), nil
	}
	return contentResult(toJSON(payload)), nil
}

func contentResult(text string) map[string]any {
	return map[string]any{"content": []any{map[string]any{"type": "text", "text": text}}}
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func str(m map[string]any, k string) string {
	if m == nil {
		return ""
	}
	s, _ := m[k].(string)
	return s
}

func intv(m map[string]any, k string, def int) int {
	if m == nil {
		return def
	}
	if f, ok := m[k].(float64); ok {
		return int(f)
	}
	return def
}

type rpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// Serve reads newline-delimited JSON-RPC messages from r and writes responses to w.
func (s *Server) Serve(r io.Reader, w io.Writer) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var msg rpcMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}
		var params map[string]any
		if len(msg.Params) > 0 {
			_ = json.Unmarshal(msg.Params, &params)
		}
		if msg.Method == "notifications/initialized" || msg.Method == "initialized" {
			continue // notification: no response
		}
		res, rpcErr := s.Dispatch(msg.Method, params)
		out := rpcMessage{JSONRPC: "2.0", ID: msg.ID, Result: res, Error: rpcErr}
		b, _ := json.Marshal(out)
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	}
	return sc.Err()
}
