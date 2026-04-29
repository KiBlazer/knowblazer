.PHONY: test build install clean

test:
	go test ./...

build:
	go build -o bin/knowblazer ./cmd/knowblazer

install:
	go install ./cmd/knowblazer

clean:
	rm -rf bin
