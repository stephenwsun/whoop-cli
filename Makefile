BINARY ?= bin/whoop

.PHONY: build test vet lint ci clean

build:
	mkdir -p bin
	go build -trimpath -o $(BINARY) ./cmd/whoop

test:
	go test ./...

vet:
	go vet ./...

lint:
	@test -z "$$(gofmt -l cmd internal whoop)" || (echo 'gofmt required:'; gofmt -l cmd internal whoop; exit 1)

ci: test vet lint build

clean:
	rm -rf bin
