BIN := priorart
GOFLAGS := CGO_ENABLED=0
LDFLAGS := -s -w -extldflags '-static'

.PHONY: build test lint clean

build:
	$(GOFLAGS) go build -ldflags "$(LDFLAGS)" -o dist/$(BIN) ./cmd/priorart

test:
	$(GOFLAGS) go test -count=1 ./...

lint:
	$(GOFLAGS) go vet ./...

clean:
	rm -rf dist
