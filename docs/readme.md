# CI Status Dashboard

CI Status Dashboard is a terminal UI written in Go that aggregates workflow/build runs from multiple CI providers into a single Bubble Tea interface. It currently targets GitHub Actions and Azure DevOps, polling each provider concurrently, normalizing run metadata, and rendering a unified table with filtering and keyboard navigation.

## Features
- **Provider abstraction:** GitHub and Azure DevOps implementations share a `core.Provider` interface, enabling easy expansion to other CI backends.
- **Thread-safe store:** A central store merges provider events into a normalized run map, providing sorted slices to the TUI.
- **Adaptive polling loops:** Providers adjust their polling interval based on active runs to balance responsiveness and rate limits.
- **Bubble Tea UI:** Runs are displayed in a stylized table with provider tabs, real-time updates, and keybindings (`Tab` to cycle providers, `r` to force refresh, `q`/Ctrl+C to quit).
- **Cobra/Viper-based config:** Configuration is loaded from `ceye.yaml` (or a path passed via `CEYE_CONFIG`) and supports environment overrides.
- **Missing-config onboarding:** When running from your workspace root (or the directory named by `--config-dir`/`CEYE_CONFIG_ROOT`), `ci-dash` automatically scans all git repositories for missing `ceye.*` files, surfaces them in the “Missing configs” panel, and lets you press `n`/`a` to highlight each repo and scaffold a template config—after creation the list refreshes so the repo disappears immediately.

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

### Automatic CLI discovery

When no `ceye.*` file is found, `ci-dash` now tries to build a configuration by calling `gh repo list <org>` and `az pipelines list --org <org> --project <project>`. This happens transparently whenever `gh`/`az` are installed and credentialed for your GitHub and Azure DevOps orgs. The defaults (`joelklabo`, `joelklabo`, `Big Timer`) can be overridden with `--github-org`, `--azure-org`, `--azure-project` (or the `CEYE_GITHUB_ORG`, `CEYE_AZURE_ORG`, `CEYE_AZURE_PROJECT` environment variables). If discovery fails it still falls back to the demo provider so the UI can start.

### Global configuration generator

If you want one config that works regardless of which repo you run from, use the CLI-based generator in this repo:

```bash
cd ~/code/ceye
python3 scripts/gen-global-ceye-config.py
```

`gh` must be logged into the `joelklabo` org, and `az` must have access to `https://dev.azure.com/joelklabo` with the `Big Timer` project; the script lists every GitHub repo in that org and every pipeline under the Azure project, then writes `~/.config/ceye/ceye.yaml`. Pass `--github-org`, `--azure-org`, `--azure-project`, or `--output` if your targets differ.

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

The header now displays the build version/commit, and you can run `ci-dash --version` to see the same information from the CLI (useful when teams compare binaries).

### Runtime provider store
Dynamic providers added at runtime are kept in `~/.config/ceye/providers.json` by default (or override via `CEYE_PROVIDER_STORE`/`--provider-store`). Use the `ci-dash provider` subcommands to inspect or mutate the runtime list without editing the main config:
- `ci-dash provider list`: show stored entries with their IDs, types, and optional `display_name`.
- `ci-dash provider add --config provider.yaml`: add a new entry defined in a YAML/JSON snippet. Include a `display_name` if you want a friendly label. Example snippet:

  ```yaml
  type: github
  display_name: frontend-ci
  repos:
    - owner: octocat
      repo: hello-world
      workflows:
        - CI
  ```

- `ci-dash provider update --id <id> --config ...`: replace the stored config.
- `ci-dash provider enable|disable --id <id>`: toggle whether a stored provider participates in polling.
- `ci-dash provider remove --id <id>`: delete the stored entry.

Stored providers are merged with your static `ceye.yaml` providers on startup, and their friendly names appear as provider tabs in the UI. Adjust `--provider-store` to point at another file when sharing dynamic lists across machines.

- Press `P` while the dashboard is running to view a provider store overlay (showing each stored entry and its enabled/disabled state) without leaving the TUI; press `Space` to toggle an entry’s enabled flag while the overlay is active, `d` to remove it, `e` to duplicate the configuration, or `E` to edit its key fields (owner/repo for GitHub, org/project[:pipelines] for Azure, or `gitlab_project` for GitLab).
 - `ci-dash provider export --file providers.json`: dump stored entries so you can share them as JSON.
- `ci-dash provider import --file providers.json`: append entries from JSON (use `--replace` to overwrite the current store).
- When `ci-dash` detects repos without `ceye.*`, the UI shows a “Missing configs” panel; press `n` to highlight each repository and `a` to scaffold a default `ceye.yaml` inside it.
  After scaffolding, `ci-dash` reruns the scan immediately so the repo vanishes from the list as soon as the new file exists.

### Demo/diagnostic flags
- `--demo` / `--demo-runs`: start with synthetic runs only.
- `--demo-duration=5s`: auto-exit after the duration (useful for automated screenshots/tests).
- `--log-events=out.jsonl`: write provider events to a JSON-lines file for debugging.
- `--notify`: emit a desktop notification whenever a provider reports an error (macOS/Linux only).
- `--history-path=<path>`: write the recent run history for each provider to a JSON file (default `~/.config/ceye/run-history.json`).
- `--webhook-url=<url>`: POST provider errors to this webhook endpoint (e.g., Slack/Teams) so you can hook alerts into other systems.
- `--config-dir=<path>`: walk the directory tree rooted at the provided path and auto-discover every `ceye.*` config file (defaults to the nearest ancestor named `code` or the current directory if none is found). Run it from `~/code` to monitor multiple repos at once.
- `D`: toggle the detail pane that shows durations/log summaries for the selected run.
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
