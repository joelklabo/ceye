GO ?= go
BINARY ?= bin/ci-dash

.PHONY: build run test fmt clean demo

build:
	@mkdir -p $$(dirname $(BINARY))
	$(GO) build -o $(BINARY) ./cmd/ci-dash

run: build
	$(BINARY)

demo:
	$(GO) run ./cmd/ci-dash --demo --demo-runs 4

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

clean:
	rm -rf bin
