.PHONY: test build lint fmt

test:
	go test -v

build:
	go build -o searchall

lint:
	golangci-lint run

fmt:
	go fmt ./...
