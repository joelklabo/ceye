# CI Status Dashboard Implementation Notebook

This document is my working plan for building the CI Status Dashboard TUI in Go. It outlines the architecture, critical design decisions, and a detailed, checkable execution list so every commit and push can be tracked explicitly.

## Vision & Goals
- Deliver a Bubble Tea-based terminal dashboard that aggregates GitHub Actions and Azure DevOps workflow runs into a single, responsive view.
- Keep the system extensible via a provider abstraction plus a thread-safe store that normalizes run data.
- Drive development with TDD and a steady cadence of small commits; every checklist item includes a slot to record the resulting commit hash and whether it was pushed (it always should be, unless otherwise justified).

## Architecture Notes
- **Tech Stack:** Go 1.21+, Cobra + Viper for CLI/config, Bubble Tea + Bubbles (table) + Lip Gloss for the UI layer, net/http clients for provider APIs.
- **Provider Model:** Each CI provider (GitHub, Azure DevOps, later more) implements a shared `Provider` interface with `Name()` and `Start(ctx, chan<- RunEvent)`; providers run in independent goroutines and stream `RunEvent` batches through a central channel.
- **Normalized Data:** A `Run` struct abstracts workflow/build data (provider, repo/project, workflow name, branch, commit, status, conclusion, timing metadata, and URL). `RunEvent` carries provider metadata plus slices of `Run` updates.
- **Aggregator / Store:** A synchronized store merges incoming `RunEvent`s, keyed by provider + run ID, and serves sorted slices to the UI. It is the single source of truth for Bubble Tea.
- **Adaptive Polling:** Providers poll quickly (≈10–15s) when any run is active and back off (≈60–120s, possibly exponential) when idle, resetting on new activity.
- **UI:** Bubble Tea model owns the store handle, provider filter state, table component, and status header/footer via Lip Gloss styling. Key bindings include navigation (arrows/j/k), tab for provider cycling, `r` for refresh, `q`/Ctrl+C to quit.
- **Configuration:** YAML/TOML config parsed by Viper (with Cobra flag `--config`). Config lists providers plus repo/pipeline selections; auth comes from env vars (e.g., `GITHUB_TOKEN`, Azure PAT).
- **Execution Flow:** Main loads config, instantiates providers, starts them with a shared event channel, runs a store-merging goroutine that emits `RunUpdatedMsg` into Bubble Tea, and runs until the user exits.

## Implementation Checklist
Each item is actionable, test-first when applicable, and records the eventual commit hash + push confirmation once completed.

1. - [x] **Step 1 – Project Initialization** (`commit: bb90f85`, `push: yes`)
   - Go module setup (`ci-dash`), git init, baseline directories (`cmd/ci-dash`, `internal/...`).
   - Add module deps (Bubble Tea, Lip Gloss, Bubbles, Cobra, Viper) and a trivial sanity test to prove the test harness.
2. - [x] **Step 2 – Core Data Types** (`commit: 3be7e8f`, `push: yes`)
   - Define `RunStatus` enum, `Run`, `RunEvent`, and `Provider` interface under `internal/core`.
3. - [x] **Step 3 – Store Tests** (`commit: 1959a79`, `push: yes`)
   - Write failing tests covering merge (new + updated runs) and `ListRuns` sorting/filtering.
4. - [x] **Step 4 – Store Implementation** (`commit: 9a01be6`, `push: yes`)
   - Implement thread-safe store with merge + list logic to satisfy Step 3 tests.
5. - [x] **Step 5 – GitHub Parser Tests** (`commit: 0ec64cd`, `push: yes`)
   - Add failing tests for parsing GitHub workflow run JSON into normalized `Run`s.
6. - [x] **Step 6 – GitHub Parser Implementation** (`commit: 9d22057`, `push: yes`)
   - Implement `ParseGitHubRuns` to satisfy Step 5 tests.
7. - [x] **Step 7 – Azure Parser Tests** (`commit: c38012e`, `push: yes`)
   - Add failing tests for Azure DevOps builds API -> `Run`s transformation.
8. - [x] **Step 8 – Azure Parser Implementation** (`commit: 0c8dcf8`, `push: yes`)
   - Implement `ParseAzureRuns` to satisfy Step 7 tests.
9. - [x] **Step 9 – GitHub Provider Start Test** (`commit: 8c8eee7`, `push: yes`)
   - Test provider polling loop with stub client + context cancellation.
10. - [x] **Step 10 – GitHub Provider Start Implementation** (`commit: cb00000`, `push: yes`)
    - Implement adaptive polling loop, auth, and event emission for GitHub.
11. - [x] **Step 11 – Azure Provider Start Test** (`commit: 36965b8`, `push: yes`)
    - Stub-client test mirroring Step 9 for Azure pipelines.
12. - [x] **Step 12 – Azure Provider Start Implementation** (`commit: 5d89830`, `push: yes`)
    - Implement Azure polling with adaptive intervals and context handling.
13. - [x] **Step 13 – Provider Factory** (`commit: fadb214`, `push: yes`)
    - Introduce registry/factory to instantiate providers from config + unit tests.
14. - [ ] **Step 14 – Config Parsing Test** (`commit: pending`, `push: pending`)
    - Write failing test loading sample YAML via Viper into config structs.
15. - [ ] **Step 15 – Config Parsing Implementation** (`commit: pending`, `push: pending`)
    - Implement `LoadConfig` to satisfy Step 14 test.
16. - [ ] **Step 16 – Main Wiring** (`commit: pending`, `push: pending`)
    - Integrate config, provider factory, store, aggregator goroutine, and Bubble Tea bootstrap.
17. - [ ] **Step 17 – Initial TUI Model** (`commit: pending`, `push: pending`)
    - Create Bubble Tea model with table, store hooks, header/footer view skeleton.
18. - [ ] **Step 18 – TUI Update Tests** (`commit: pending`, `push: pending`)
    - Unit tests covering `RunUpdatedMsg` handling and provider filtering.
19. - [ ] **Step 19 – Key Handling Tests** (`commit: pending`, `push: pending`)
    - Test quitting, provider cycling, and table integration for navigation keys.
20. - [ ] **Step 20 – TUI Finalization** (`commit: pending`, `push: pending`)
    - Implement keybindings, Lip Gloss styling, refresh hooks, and ensure manual UX sanity.
21. - [ ] **Step 21 – Makefile & Tooling** (`commit: pending`, `push: pending`)
    - Add Makefile targets (`build`, `run`, `test`, optional `fmt`/`lint`).
22. - [ ] **Step 22 – Documentation** (`commit: pending`, `push: pending`)
    - Expand README with usage, config example, env vars, and architecture overview.
23. - [ ] **Step 23 – Final QA** (`commit: pending`, `push: pending`)
    - `go test ./...`, manual CI provider smoke test, fmt/vet, and final cleanup.

## Tracking Updates
When a step finishes, update its checklist line to `[x]`, replace `commit: pending` with the actual hash (e.g., `commit: abc1234`), and mark `push: yes` (or justify `push: no` if absolutely necessary). Add brief notes inline or append short bullet points under the step if context is useful for future reference.
