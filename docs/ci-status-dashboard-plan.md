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
14. - [x] **Step 14 – Config Parsing Test** (`commit: 467678f`, `push: yes`)
    - Write failing test loading sample YAML via Viper into config structs.
15. - [x] **Step 15 – Config Parsing Implementation** (`commit: c97f43e`, `push: yes`)
    - Implement `LoadConfig` to satisfy Step 14 test.
16. - [x] **Step 16 – Main Wiring** (`commit: 72e7a4a`, `push: yes`)
    - Integrate config, provider factory, store, aggregator goroutine, and Bubble Tea bootstrap.
17. - [x] **Step 17 – Initial TUI Model** (`commit: ba368cd`, `push: yes`)
    - Create Bubble Tea model with table, store hooks, header/footer view skeleton.
18. - [x] **Step 18 – TUI Update Tests** (`commit: 1d36e0f`, `push: yes`)
    - Unit tests covering `RunUpdatedMsg` handling and provider filtering.
19. - [x] **Step 19 – Key Handling Tests** (`commit: f74060d`, `push: yes`)
    - Test quitting, provider cycling, and table integration for navigation keys.
20. - [x] **Step 20 – TUI Finalization** (`commit: 9568498`, `push: yes`)
    - Implement keybindings, Lip Gloss styling, refresh hooks, and ensure manual UX sanity.
21. - [x] **Step 21 – Makefile & Tooling** (`commit: 0ba55fd`, `push: yes`)
    - Add Makefile targets (`build`, `run`, `test`, optional `fmt`/`lint`).
22. - [x] **Step 22 – Documentation** (`commit: 53de48d`, `push: yes`)
    - Expand README with usage, config example, env vars, and architecture overview.
23. - [x] **Step 23 – Final QA** (`commit: a2b28da`, `push: yes`)
    - `go test ./...`, manual CI provider smoke test, fmt/vet, and final cleanup.
24. - [x] **Step 24 – CLI Flags** (`commit: da28875`, `push: yes`)
    - Add Cobra-based CLI entrypoint with `--config` flag and route execution through root command.
25. - [x] **Step 25 – Example Config** (`commit: 9e89cfd`, `push: yes`)
    - Add `config.example.yaml` plus README note so users can copy/edit quickly.
26. - [x] **Step 26 – UI Header & Styling** (`commit: ca0469d`, `push: yes`)
    - Introduce Lip Gloss-styled header/footer with last update timestamp and key hints.
27. - [x] **Step 27 – Refresh Key** (`commit: f76f887`, `push: yes`)
    - Handle `r` keybinding to trigger immediate refresh: aggregator should notify providers via a channel or context poke.
28. - [x] **Step 28 – GitHub HTTP client** (`commit: 5e0ff0a`, `push: yes`)
    - Implement real GitHub API client using `net/http`, injecting it into provider factory with auth token support (supports credentials from config/environment).
29. - [x] **Step 29 – Azure HTTP client** (`commit: 7196660`, `push: yes`)
    - Implement Azure DevOps REST client and wire PAT support via config/env.
30. - [x] **Step 30 – Provider errors & status** (`commit: 12951be`, `push: yes`)
    - Surface provider errors in UI header/footer and log warnings when polls fail.
31. - [x] **Step 31 – Table status coloring** (`commit: 362f36c`, `push: yes`)
    - Apply Lip Gloss styles to the status column (Success/Failed/Running colors) for better visibility.
32. - [x] **Step 32 – Status-aware sorting** (`commit: 001de78`, `push: yes`)
    - Sort runs by status buckets (running > queued > failed > success) before timestamp.
33. - [x] **Step 33 – Run details view** (`commit: 5cc6a37`, `push: yes`)
    - Add ability to select a row and show details (repo, branch, commit, URL) in a pane or popup.
34. - [x] **Step 34 – Table keybindings** (`commit: 8b8f8e6`, `push: yes`)
    - Explicitly handle arrow keys (↑/↓, j/k) and page navigation by delegating to the table component.
35. - [x] **Step 35 – Status filter** (`commit: 0a8a9d4`, `push: yes`)
    - Allow filtering runs by status (e.g., failures only) via a keybinding.
36. - [x] **Step 36 – URL opening** (`commit: 119ec6c`, `push: yes`)
    - Add keybinding (e.g., `o` or Enter) to open the selected run's URL in the default browser.
37. - [x] **Step 37 – Keyboard help** (`commit: 8fddbd9`, `push: yes`)
    - Provide an on-screen summary of keybindings or toggleable help dialog.
38. - [x] **Step 38 – Text filter** (`commit: 2115f57`, `push: yes`)
    - Add substring filter to show runs matching a query (e.g., `/` to enter search).
39. - [x] **Step 39 – Aggregated stats** (`commit: c80f212`, `push: yes`)
    - Display counts of running/failed/successful runs in header or summary line.
40. - [x] **Step 40 – Provider metrics** (`commit: ad31c36`, `push: yes`)
    - Track per-provider last update/error timestamps and show them alongside counts.
41. - [x] **Step 41 – Provider-level filtering** (`commit: d0dda0a`, `push: yes`)
    - Support toggling individual providers on/off or selecting them from a list beyond Tab cycling (e.g., `p` to open provider palette).
42. - [ ] **Step 42 – Provider palette UI** (`commit: pending`, `push: pending`)
    - Implement an overlay list to toggle providers (e.g., press `p` to show checkboxes and Enter to apply).
42. - [ ] **Step 42 – Provider palette UI** (`commit: pending`, `push: pending`)
    - Implement an overlay list to toggle providers (e.g., press `p` to show checkboxes and Enter to apply).
39. - [ ] **Step 39 – Aggregated stats** (`commit: pending`, `push: pending`)
    - Display counts of running/failed/successful runs in header or summary line.

## Tracking Updates
When a step finishes, update its checklist line to `[x]`, replace `commit: pending` with the actual hash (e.g., `commit: abc1234`), and mark `push: yes` (or justify `push: no` if absolutely necessary). Add brief notes inline or append short bullet points under the step if context is useful for future reference.
