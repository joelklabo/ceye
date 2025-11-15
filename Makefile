GO ?= go
BINARY ?= bin/ci-dash

.PHONY: build run test fmt clean

build:
	@mkdir -p $$(dirname $(BINARY))
	$(GO) build -o $(BINARY) ./cmd/ci-dash

run: build
	$(BINARY)

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

clean:
	rm -rf bin
