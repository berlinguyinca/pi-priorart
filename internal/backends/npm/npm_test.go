package npm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchParsesCandidates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/-/v1/search" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"objects":[{"package":{"name":"got","version":"11.0.0","description":"simple HTTP","links":{"repository":"https://github.com/x/got","homepage":"https://github.com/x/got"}}}]}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client())
	got, err := c.Search(context.Background(), "http client", "js", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "got" || got[0].Source != "oss" || got[0].Ecosystem != "js" {
		t.Fatalf("bad result: %+v", got)
	}
	if got[0].RepoURL != "https://github.com/x/got" {
		t.Fatalf("repo url: %+v", got[0])
	}
}

func TestSearchNon200IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer srv.Close()
	if _, err := New(srv.URL, srv.Client()).Search(context.Background(), "x", "js", 5); err == nil {
		t.Fatal("expected error on 500")
	}
}
