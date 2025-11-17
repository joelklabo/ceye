GO ?= go
BINARY ?= bin/ceye
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -X 'main.Version=$(VERSION)' -X 'main.GitCommit=$(COMMIT)' -X 'main.BuildTime=$(BUILD_TIME)'

.PHONY: build run test fmt clean demo snapshot install

build:
	@echo "🔨 Building ceye..."
	@mkdir -p $$(dirname $(BINARY))
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/ceye
	@echo "✅ Built: $(BINARY)"
	@./$(BINARY) --version

install: build
	@echo "📦 Installing to /usr/local/bin/ceye..."
	sudo install -m 755 $(BINARY) /usr/local/bin/ceye
	@sudo xattr -c /usr/local/bin/ceye 2>/dev/null || true
	@echo "✅ Installed!"
	@ceye --version

run: build
	$(BINARY)

demo:
	$(GO) run ./cmd/ceye --demo --demo-runs 4

SNAPSHOT_SECONDS ?= 5
snapshot:
	tmux new-session -d -s ceye-snapshot "cd $(PWD) && $(GO) run ./cmd/ceye --demo --demo-runs 4 --demo-duration $(SNAPSHOT_SECONDS)s --log-events docs/demo-events.jsonl"; \
	sleep $(SNAPSHOT_SECONDS); \
	tmux capture-pane -t ceye-snapshot -p > docs/ui-demo.txt; \
	tmux kill-session -t ceye-snapshot

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

clean:
	rm -rf bin
