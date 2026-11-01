// Package crates searches crates.io for Rust prior art.
package crates

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
		baseURL = "https://crates.io"
	}
	if httpc == nil {
		httpc = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{Base: baseURL, HTTP: httpc}
}

func (c *Client) Name() string { return "crates" }

type cratesResponse struct {
	Crates []struct {
		Name        string `json:"name"`
		MaxVersion  string `json:"max_version"`
		Description string `json:"description"`
		Repository  string `json:"repository"`
		UpdatedAt   string `json:"updated_at"`
	} `json:"crates"`
}

func (c *Client) Search(ctx context.Context, query, _ string, limit int) ([]candidate.Candidate, error) {
	if limit <= 0 {
		limit = 5
	}
	u := fmt.Sprintf("%s/api/v1/crates?q=%s&per_page=%d", c.Base, url.QueryEscape(query), limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "priorart/1.0")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crates: status %d", resp.StatusCode)
	}
	var r cratesResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	out := make([]candidate.Candidate, 0, len(r.Crates))
	for _, c2 := range r.Crates {
		out = append(out, candidate.Candidate{
			Source: "oss", Ecosystem: "rs",
			Name: c2.Name, Version: c2.MaxVersion, RepoURL: c2.Repository,
			Activity: candidate.Activity{LastRelease: c2.UpdatedAt},
			Snippet:  c2.Description, MatchReason: "crates.io search match",
		})
	}
	return out, nil
}
