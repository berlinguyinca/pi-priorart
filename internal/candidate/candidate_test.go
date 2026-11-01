package candidate

import (
	"encoding/json"
	"testing"
)

func TestCandidateJSONRoundTrip(t *testing.T) {
	in := Candidate{
		Source: "oss", Ecosystem: "js", Name: "left-pad", Version: "1.3.0",
		RepoURL: "https://github.com/x/left-pad", License: "MIT",
		Activity: Activity{LastRelease: "2026-01-01", Stars: 100},
		Snippet:  "pads", MatchReason: "name match", DocsURL: "https://docs",
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out Candidate
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Name != "left-pad" || out.Activity.Stars != 100 || out.Source != "oss" {
		t.Fatalf("round trip mismatch: %+v", out)
	}
}
