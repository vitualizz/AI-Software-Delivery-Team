.PHONY: build test lint clean hooks

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	go clean ./...

hooks:
	go install github.com/evilmartians/lefthook@latest
	lefthook install
