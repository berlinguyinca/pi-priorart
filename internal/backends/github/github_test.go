package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRepoSearchParsesAndSendsToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/search/repositories" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"full_name":"avast/retry-go","html_url":"https://github.com/avast/retry-go","description":"retry","stargazers_count":2000,"license":{"spdx_id":"MIT"},"language":"Go","pushed_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "ghp_test", srv.Client())
	got, err := c.Search(context.Background(), "retry", "go", 5)
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer ghp_test" {
		t.Fatalf("auth header %q", gotAuth)
	}
	if len(got) != 1 || got[0].Name != "avast/retry-go" || got[0].License != "MIT" || got[0].Activity.Stars != 2000 {
		t.Fatalf("bad: %+v", got)
	}
}

func TestNoTokenIsError(t *testing.T) {
	c := New("https://api.github.com", "", nil)
	if _, err := c.Search(context.Background(), "x", "go", 5); err == nil {
		t.Fatal("expected error when token empty")
	}
}
