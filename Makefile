.PHONY: build test lint fmt run clean check

build:
	go build -o bin/harvest-tui ./cmd/harvest-tui

test:
	go test -v ./...

lint:
	go vet ./...

fmt:
	go fmt ./...

check: fmt lint test

run: build
	./bin/harvest-tui

clean:
	rm -rf bin/