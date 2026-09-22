BINARY ?= bin/whoop

.PHONY: build test race vet vuln lint ci clean

build:
	mkdir -p bin
	go build -trimpath -o $(BINARY) ./cmd/whoop

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

vuln:
	GOTOOLCHAIN=go1.25.13 go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...

lint:
	@test -z "$$(gofmt -l cmd internal whoop)" || (echo 'gofmt required:'; gofmt -l cmd internal whoop; exit 1)

ci: test vet lint build

clean:
	rm -rf bin
