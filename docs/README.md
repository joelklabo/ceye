# CI Status Dashboard

CI Status Dashboard is a terminal UI written in Go that aggregates workflow/build runs from multiple CI providers into a single Bubble Tea interface. It currently targets GitHub Actions and Azure DevOps, polling each provider concurrently, normalizing run metadata, and rendering a unified table with filtering and keyboard navigation.

## Features
- **Provider abstraction:** GitHub and Azure DevOps implementations share a `core.Provider` interface, enabling easy expansion to other CI backends.
- **Thread-safe store:** A central store merges provider events into a normalized run map, providing sorted slices to the TUI.
- **Adaptive polling loops:** Providers adjust their polling interval based on active runs to balance responsiveness and rate limits.
- **Bubble Tea UI:** Runs are displayed in a stylized table with provider tabs, real-time updates, and keybindings (`Tab` to cycle providers, `r` to force refresh, `q`/Ctrl+C to quit).
- **Cobra/Viper-based config:** Configuration is loaded from `ceye.yaml` (or a path passed via `CEYE_CONFIG`) and supports environment overrides.

## Getting Started

### Prerequisites
- Go 1.21+
- Valid credentials for the providers you plan to monitor (e.g., `GITHUB_TOKEN`/`CEYE_GITHUB_TOKEN` for GitHub, Azure DevOps PAT for Azure).

### Installation & Build
```bash
# Clone the repo
$ git clone https://github.com/joelklabo/ceye
$ cd ceye

# Run tests
$ make test

# Build the binary into bin/ci-dash
$ make build

# Run directly
$ make run -- --config path/to/ceye.yaml

# Demo mode (no config/credentials required)
$ go run ./cmd/ci-dash --demo --demo-runs 4
$ make demo  # equivalent helper target
```

Alternatively, you can run `go run ./cmd/ci-dash --config path/to/config.yaml` once CLI flags are wired.

## Configuration
Create `ceye.yaml` (or point `CEYE_CONFIG` to a custom file). Example (`config.example.yaml` is checked into the repo for convenience):

```yaml
providers:
  - type: demo
    runs: 4
  # Optional GitLab config (no auth required for this demo implementation)
  # - type: gitlab
  #   gitlab_project: example-org/project
```

Set provider credentials via environment variables (`GITHUB_TOKEN`, `AZURE_DEVOPS_PAT`, etc.). The config loader searches `./ceye.yaml` then `~/.config/ceye/ceye.yaml` by default, and respects `CEYE_*` env overrides.
The GitHub provider automatically reads `CEYE_GITHUB_TOKEN` (preferred) or `GITHUB_TOKEN` for authentication.

### Demo provider
The optional `demo` provider emits synthetic runs so you can verify the UI without real credentials. Include it in your config (as shown above) and it will stream Build/Test/Deploy runs that cycle through queued/running/success/failure states.

## Usage
Once running, the dashboard polls providers and refreshes the table automatically. Key bindings:
- `Tab`: Cycle provider filter (All → GitHub → Azure → ...)
- `f`: Cycle status filter (All → Running → Queued → Failed → Success)
- `t`: Cycle sort mode (status, updated time, duration)
- `p`: Toggle provider palette (space toggles visibility, Enter/Esc closes)
- `/`: Start a substring filter; type to filter, Enter/Esc to exit
- Arrow keys / `j`, `k`: Navigate rows; PageUp/PageDown handled by the table component
- `Enter` or `o`: Open the selected run in your default browser
- `y`: Copy the run URL to the clipboard
- `c`: Copy a summary (provider • repo • branch • workflow/status • URL)
- `v`: Toggle focus mode (full-width table vs. paneled view)
- `r`: Force immediate provider refresh
- `?`: Toggle the help overlay
- `q` or `Ctrl+C`: Quit

A flash message appears beneath the header after copy operations or other actions so you know the key press succeeded.

### Demo/diagnostic flags
- `--demo` / `--demo-runs`: start with synthetic runs only.
- `--demo-duration=5s`: auto-exit after the duration (useful for automated screenshots/tests).
- `--log-events=out.jsonl`: write provider events to a JSON-lines file for debugging.
- `--notify`: emit a desktop notification whenever a provider reports an error (macOS/Linux only).
- `run history`: the right sidebar shows each provider’s last few run summaries for quick inspection.
- `make demo`: convenience wrapper for `go run ./cmd/ci-dash --demo --demo-runs 4`.
- `make snapshot`: launches demo mode in tmux, waits a few seconds, and writes the current TUI into `docs/ui-demo.txt`.

## Architecture Overview
- `cmd/ci-dash`: Entrypoint wiring config, providers, store, and the Bubble Tea program.
- `internal/core`: Core types (`Run`, `RunEvent`, `RunStatus`), provider interface, and thread-safe store.
- `internal/providers`: Provider-specific clients and polling loops (`github`, `azure`) plus a factory for config-driven instantiation.
- `internal/ui`: Bubble Tea model, table rendering, and RunUpdated message handling.
- `internal/config`: Viper-based loader with sane defaults and env overrides.
- `docs/ci-status-dashboard-plan.md`: The working implementation notebook tracking progress via commits and pushes.

## Development Workflow
1. Update `docs/ci-status-dashboard-plan.md` as steps complete, recording commit hashes and push state.
2. Follow TDD for new components: add failing tests, implement functionality, run `go test ./...` before every commit.
3. Use `make fmt` to run `go fmt ./...` and `make test` before pushing.
4. Push every commit to trigger CI and keep the remote plan synchronized.

## Future Work
- Implement `r` refresh key with provider pokes.
- Extend UI with Lip Gloss styling, detailed run views, and status coloring.
- Add additional providers (GitLab, Jenkins, CircleCI, etc.).
- Introduce CLI flags via Cobra and support live configuration reloads.
- Add end-to-end tests and sample configs under `examples/`.
