// Package wheels lists "known wheels" that must not be reimplemented.
package wheels

import "github.com/berlinguyinca/priorart/internal/candidate"

func All() []candidate.Wheel {
	return []candidate.Wheel{
		{Name: "Retry / backoff", Description: "retry with exponential backoff + jitter",
			Canonical: []string{"avast/retry-go", "tenacity", "p-queue", "backon"}},
		{Name: "CSV parsing", Description: "read/write CSV",
			Canonical: []string{"csv", "pandas", "csv-parse"}},
		{Name: "JSON parsing", Description: "marshal/unmarshal JSON",
			Canonical: []string{"encoding/json", "serde_json"}},
		{Name: "Date/time", Description: "parse/format timestamps, timezones",
			Canonical: []string{"time", "python-dateutil", "chrono"}},
		{Name: "Slugify", Description: "slug from string",
			Canonical: []string{"python-slugify", "slugify"}},
		{Name: "Hashing", Description: "md5/sha/crc",
			Canonical: []string{"crypto/md5", "crypto/sha256"}},
		{Name: "Logging setup", Description: "structured logging",
			Canonical: []string{"log/slog", "structlog", "tracing"}},
		{Name: "HTTP client", Description: "HTTP requests with retries/timeouts",
			Canonical: []string{"net/http", "requests", "reqwest"}},
		{Name: "CLI scaffolding", Description: "command-line argument parsing",
			Canonical: []string{"flag", "click", "clap"}},
		{Name: "ID generation", Description: "uuid/ulid/nanoid",
			Canonical: []string{"google/uuid", "uuid6", "ulid"}},
	}
}
