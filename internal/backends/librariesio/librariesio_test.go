package librariesio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/search/pypi" {
			t.Errorf("path %s", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "requests" {
			t.Errorf("q = %q, want requests", r.URL.Query().Get("q"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"packages":[{"name":"requests","pretty_name":"Requests","url":"https://libraries.io/pypi/requests","repo_url":"https://github.com/psf/requests","license":"Apache-2.0","version":"2.31.0","platform":"pypi","stats":{"last_updated":"2026-01-01T00:00:00Z","stars":50000}}]}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "", srv.Client())
	got, err := c.Search(context.Background(), "requests", "py", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "requests" || got[0].License != "Apache-2.0" || got[0].Activity.Stars != 50000 {
		t.Fatalf("bad: %+v", got)
	}
}

func TestSearchSendsAPIKeyAndEcosystem(t *testing.T) {
	var gotPath, gotKey, gotQ string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.URL.Query().Get("api_key")
		gotQ = r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"packages":[]}`))
	}))
	defer srv.Close()
	c := New(srv.URL, "SECRET", srv.Client())
	if _, err := c.Search(context.Background(), "backoff", "go", 3); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/search/go" {
		t.Fatalf("path = %q, want /api/search/go", gotPath)
	}
	if gotKey != "SECRET" {
		t.Fatalf("api_key = %q, want SECRET", gotKey)
	}
	if gotQ != "backoff" {
		t.Fatalf("q = %q, want backoff", gotQ)
	}
}
