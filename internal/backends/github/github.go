// Package github searches GitHub repositories for prior art.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/berlinguyinca/priorart/internal/candidate"
)

type Client struct {
	Base  string
	Token string
	HTTP  *http.Client
}

func New(baseURL, token string, httpc *http.Client) *Client {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	if httpc == nil {
		httpc = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{Base: baseURL, Token: token, HTTP: httpc}
}

func (c *Client) Name() string { return "github" }

type repoSearchResponse struct {
	Items []struct {
		FullName        string `json:"full_name"`
		HTMLURL         string `json:"html_url"`
		Description     string `json:"description"`
		StargazersCount int    `json:"stargazers_count"`
		License         *struct {
			SpdxID string `json:"spdx_id"`
		} `json:"license"`
		Language string `json:"language"`
		PushedAt string `json:"pushed_at"`
	} `json:"items"`
}

func (c *Client) do(ctx context.Context, q string, limit int) (*repoSearchResponse, error) {
	if c.Token == "" {
		return nil, errors.New("github: GITHUB_TOKEN not set")
	}
	u := fmt.Sprintf("%s/search/repositories?q=%s&per_page=%d", c.Base, url.QueryEscape(q), limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: status %d", resp.StatusCode)
	}
	var r repoSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) toCandidates(r *repoSearchResponse, eco string) []candidate.Candidate {
	out := make([]candidate.Candidate, 0, len(r.Items))
	for _, it := range r.Items {
		lic := ""
		if it.License != nil {
			lic = it.License.SpdxID
		}
		out = append(out, candidate.Candidate{
			Source: "oss", Ecosystem: eco,
			Name: it.FullName, RepoURL: it.HTMLURL, License: lic,
			Activity: candidate.Activity{LastRelease: it.PushedAt, Stars: it.StargazersCount},
			Snippet:  it.Description, MatchReason: "github repo search match", DocsURL: it.HTMLURL,
		})
	}
	return out
}

// Search runs a repository search; when ecosystem is "go" it scopes to Go.
func (c *Client) Search(ctx context.Context, query, ecosystem string, limit int) ([]candidate.Candidate, error) {
	if limit <= 0 {
		limit = 5
	}
	q, eco := query, "oss"
	if ecosystem == "go" {
		q = query + " language:Go"
		eco = "go"
	}
	r, err := c.do(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	return c.toCandidates(r, eco), nil
}

// SearchGo is an explicit Go-only repository search.
func (c *Client) SearchGo(ctx context.Context, query string, limit int) ([]candidate.Candidate, error) {
	if limit <= 0 {
		limit = 5
	}
	r, err := c.do(ctx, query+" language:Go", limit)
	if err != nil {
		return nil, err
	}
	return c.toCandidates(r, "go"), nil
}
