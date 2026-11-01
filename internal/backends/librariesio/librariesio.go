// Package librariesio searches libraries.io (cross-registry) for prior art.
package librariesio

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
		baseURL = "https://libraries.io"
	}
	if httpc == nil {
		httpc = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{Base: baseURL, HTTP: httpc}
}

func (c *Client) Name() string { return "librariesio" }

type liResponse struct {
	Packages []struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		RepoURL  string `json:"repo_url"`
		License  string `json:"license"`
		Version  string `json:"version"`
		Platform string `json:"platform"`
		Stats    struct {
			LastUpdated string `json:"last_updated"`
			Stars       int    `json:"stars"`
		} `json:"stats"`
	} `json:"packages"`
}

func (c *Client) Search(ctx context.Context, query, ecosystem string, limit int) ([]candidate.Candidate, error) {
	if limit <= 0 {
		limit = 5
	}
	u := fmt.Sprintf("%s/api/search/%s?per_page=%d", c.Base, url.PathEscape(query), limit)
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
		return nil, fmt.Errorf("librariesio: status %d", resp.StatusCode)
	}
	var r liResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	out := make([]candidate.Candidate, 0, len(r.Packages))
	for _, p := range r.Packages {
		out = append(out, candidate.Candidate{
			Source: "oss", Ecosystem: eco(p.Platform),
			Name: p.Name, Version: p.Version, RepoURL: p.RepoURL, License: p.License,
			Activity: candidate.Activity{LastRelease: p.Stats.LastUpdated, Stars: p.Stats.Stars},
			Snippet:  p.URL, MatchReason: "libraries.io search match", DocsURL: p.URL,
		})
	}
	return out, nil
}

func eco(platform string) string {
	switch platform {
	case "crates.io":
		return "rs"
	case "go", "golang":
		return "go"
	case "npm":
		return "js"
	default:
		return "py"
	}
}
