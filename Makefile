.PHONY: test build install clean cross

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -s -w -X github.com/knowblazer/knowblazer/internal/version.Version=$(VERSION) -X github.com/knowblazer/knowblazer/internal/version.Commit=$(COMMIT) -X github.com/knowblazer/knowblazer/internal/version.Date=$(DATE)

test:
	go test ./...

build:
	go build -ldflags "$(LDFLAGS)" -o bin/knowblazer ./cmd/knowblazer

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/knowblazer

cross:
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/knowblazer-linux-amd64 ./cmd/knowblazer
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/knowblazer-darwin-amd64 ./cmd/knowblazer
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/knowblazer-darwin-arm64 ./cmd/knowblazer
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/knowblazer-windows-amd64.exe ./cmd/knowblazer

clean:
	rm -rf bin
