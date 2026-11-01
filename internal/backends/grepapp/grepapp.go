// Package grepapp searches public code via grep.app (no auth).
package grepapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/berlinguyinca/priorart/internal/candidate"
)

type Client struct {
	Base string
	HTTP *http.Client
}

func New(baseURL string, httpc *http.Client) *Client {
	if baseURL == "" {
		baseURL = "https://grep.app"
	}
	if httpc == nil {
		httpc = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{Base: baseURL, HTTP: httpc}
}

func (c *Client) Name() string { return "grepapp" }

type grepappResponse struct {
	Hits struct {
		Hits []struct {
			Repo struct {
				Raw          string `json:"raw"`
				HTTPCloneURL string `json:"http_clone_url"`
			} `json:"repo"`
			RepoPath string `json:"repoPath"`
			Content  struct {
				Content string `json:"content"`
			} `json:"content"`
		} `json:"hits"`
	} `json:"hits"`
}

func (c *Client) Search(ctx context.Context, query, _ string, limit int) ([]candidate.Candidate, error) {
	if limit <= 0 {
		limit = 5
	}
	u := fmt.Sprintf("%s/api/search?q=%s", c.Base, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("grepapp: status %d", resp.StatusCode)
	}
	var r grepappResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	out := make([]candidate.Candidate, 0, len(r.Hits.Hits))
	for _, h := range r.Hits.Hits {
		if len(out) >= limit {
			break
		}
		out = append(out, candidate.Candidate{
			Source:  "oss",
			Name:    h.Repo.Raw,
			RepoURL: h.Repo.HTTPCloneURL,
			Snippet: h.Content.Content, MatchReason: "grep.app code match: " + h.RepoPath,
		})
	}
	return out, nil
}
