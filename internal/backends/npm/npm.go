// Package npm searches the npm registry for JS/TS prior art.
package npm

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
		baseURL = "https://registry.npmjs.org"
	}
	if httpc == nil {
		httpc = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{Base: baseURL, HTTP: httpc}
}

func (c *Client) Name() string { return "npm" }

type npmSearchResponse struct {
	Objects []struct {
		Package struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
			Links       struct {
				Repository string `json:"repository"`
				Homepage   string `json:"homepage"`
			} `json:"links"`
		} `json:"package"`
	} `json:"objects"`
}

func (c *Client) Search(ctx context.Context, query, _ string, limit int) ([]candidate.Candidate, error) {
	if limit <= 0 {
		limit = 5
	}
	u := fmt.Sprintf("%s/-/v1/search?text=%s&size=%d", c.Base, url.QueryEscape(query), limit)
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
		return nil, fmt.Errorf("npm: status %d", resp.StatusCode)
	}
	var r npmSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	out := make([]candidate.Candidate, 0, len(r.Objects))
	for _, o := range r.Objects {
		out = append(out, candidate.Candidate{
			Source: "oss", Ecosystem: "js",
			Name: o.Package.Name, Version: o.Package.Version,
			RepoURL: o.Package.Links.Repository, DocsURL: o.Package.Links.Homepage,
			Snippet: o.Package.Description, MatchReason: "npm search match",
		})
	}
	return out, nil
}
