.PHONY: test build install clean cross

test:
	go test ./...

build:
	go build -o bin/knowblazer ./cmd/knowblazer

install:
	go install ./cmd/knowblazer

cross:
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/knowblazer-linux-amd64 ./cmd/knowblazer
	GOOS=darwin GOARCH=amd64 go build -o bin/knowblazer-darwin-amd64 ./cmd/knowblazer
	GOOS=darwin GOARCH=arm64 go build -o bin/knowblazer-darwin-arm64 ./cmd/knowblazer
	GOOS=windows GOARCH=amd64 go build -o bin/knowblazer-windows-amd64.exe ./cmd/knowblazer

clean:
	rm -rf bin
