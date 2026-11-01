package wheels

import (
	"strings"
	"testing"
)

func TestAllHasCoreWheels(t *testing.T) {
	all := All()
	if len(all) < 6 {
		t.Fatalf("too few wheels: %d", len(all))
	}
	joined := ""
	for _, w := range all {
		joined += w.Name + " "
	}
	for _, want := range []string{"retry", "csv", "date"} {
		if !strings.Contains(strings.ToLower(joined), want) {
			t.Fatalf("missing wheel %q in %q", want, joined)
		}
	}
}
