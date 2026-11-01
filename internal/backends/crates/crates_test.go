package crates

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/crates" {
			t.Errorf("path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"crates":[{"id":"serde","name":"serde","max_version":"1.0.200","description":"Serialization","repository":"https://github.com/serde-rs/serde","updated_at":"2026-01-01T00:00:00Z"}]}`))
	}))
	defer srv.Close()
	got, err := New(srv.URL, srv.Client()).Search(context.Background(), "serde", "rs", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "serde" || got[0].Ecosystem != "rs" || got[0].Version != "1.0.200" {
		t.Fatalf("bad: %+v", got)
	}
}
