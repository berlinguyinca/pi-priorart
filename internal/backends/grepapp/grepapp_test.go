package grepapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchParsesHits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hits":{"hits":[{"repo":{"raw":"avast/retry-go","http_clone_url":"https://github.com/avast/retry-go"},"repoPath":"retry.go","content":{"content":"func Retry(...)"}}],"total":1}}`))
	}))
	defer srv.Close()
	got, err := New(srv.URL, srv.Client()).Search(context.Background(), "Retry", "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "avast/retry-go" || got[0].RepoURL != "https://github.com/avast/retry-go" {
		t.Fatalf("bad: %+v", got)
	}
}
