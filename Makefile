GO ?= go
BINARY ?= bin/ceye

.PHONY: build run test fmt clean demo snapshot

build:
	@mkdir -p $$(dirname $(BINARY))
	$(GO) build -o $(BINARY) ./cmd/ceye

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
