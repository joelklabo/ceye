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
```

Alternatively, you can run `go run ./cmd/ci-dash --config path/to/config.yaml` once CLI flags are wired.

## Configuration
Create `ceye.yaml` (or point `CEYE_CONFIG` to a custom file). Example (`config.example.yaml` is checked into the repo for convenience):

```yaml
providers:
  - type: github
    repos:
      - owner: octocat
        repo: Hello-World
        workflows: ["CI", "Deploy"]
  - type: azure
    org: myorg
    project: MyProject
    pipelines: [42, 43]
```

Set provider credentials via environment variables (`GITHUB_TOKEN`, `AZURE_DEVOPS_PAT`, etc.). The config loader searches `./ceye.yaml` then `~/.config/ceye/ceye.yaml` by default, and respects `CEYE_*` env overrides.
The GitHub provider automatically reads `CEYE_GITHUB_TOKEN` (preferred) or `GITHUB_TOKEN` for authentication.

## Usage
Once running, the dashboard will start polling providers and updating the table automatically. Key bindings:
- `Tab`: Cycle provider filter (All → GitHub → Azure → ...)
- `r`: Force an immediate refresh of all providers
- `q` or `Ctrl+C`: Quit
- Highlight a run with the arrow keys (or `j`/`k`) to see its repo/branch/URL details in the pane beneath the table
- `f`: Cycle status filter (All → Running → Queued → Failed → Success)
- `o`: Open the selected run in your default browser
- Arrow keys / `j`, `k`: Navigate table rows (handled by Bubbles table component)

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
