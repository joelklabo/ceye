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
42. - [x] **Step 42 – Provider palette UI** (`commit: b5932b9`, `push: yes`)
    - Implement an overlay list to toggle providers (e.g., press `p` to show checkboxes and Enter to apply).
43. - [x] **Step 43 – Layout & styling overhaul** (`commit: b5932b9`, `push: yes`)
    - Rework header/table/footer layout using Lip Gloss/Bubbles patterns (help model, badges, bordered sections) for a polished look.
44. - [x] **Step 44 – Inspiration-driven layout refresh** (`commit: 540102c`, `push: yes`)
    - Borrowed pane layout patterns from Glow/LazyGit, introduced a stats strip, provider panel, refreshed selection/log panes, and improved time/duration formatting.
45. - [x] **Step 45 – Dedicated help overlay** (`commit: 5fd9324`, `push: yes`)
    - Replace the inline help toggle with a modal overlay mirroring Glow/LazyGit reference patterns so shortcuts are easier to discover; ensure tests keep covering this interaction.
46. - [x] **Step 46 – Focus view toggle** (`commit: b78fa57`, `push: yes`)
    - Add a keybinding to switch between dashboard/panel layout and a full-width table focus mode so the UI adapts to different workflows.
47. - [x] **Step 47 – Sort cycling** (`commit: c1bd25c`, `push: yes`)
    - Allow switching between status, updated-time, and duration orderings so the most relevant runs float to the top.
48. - [x] **Step 48 – Color-coded activity log** (`commit: b3ef799`, `push: yes`)
    - Track log severity (info/warn/error) so the Activity pane highlights provider errors distinctly.
49. - [x] **Step 49 – Responsive layout** (`commit: 269abe3`, `push: yes`)
    - Capture terminal size and switch to a stacked mobile-style layout when the window is narrow so the dashboard stays legible.
50. - [x] **Step 50 – Copy run URL key** (`commit: 0fae62b`, `push: yes`)
    - Provide a keybinding to copy the selected run's URL to the clipboard (in addition to opening it) for workflows that prefer sharing links.
51. - [x] **Step 51 – Enter key opens run** (`commit: 8a8b48e`, `push: yes`)
    - Make the Enter key open the selected run URL (alongside `o`) so the UX matches expectations from other TUIs.
52. - [x] **Step 52 – Copy run summary** (`commit: 8cd0eb4`, `push: yes`)
    - Add a keybinding to copy a formatted summary of the selected run (provider, repo, branch, status, URL) for sharing in chat or tickets.
53. - [x] **Step 53 – Action confirmations** (`commit: 0d5dca3`, `push: yes`)
    - Show brief confirmation toasts (e.g., “Copied run URL”) so feedback is visible after key actions.
54. - [x] **Step 54 – Keyboard shortcut docs** (`commit: a4794c7`, `push: yes`)
    - Sync README/docs with the latest keybindings (focus mode, sorting, copy actions, etc.).
55. - [x] **Step 55 – Demo provider** (`commit: 2932655`, `push: yes`)
    - Add a built-in demo provider so we can verify the UI without real credentials.
56. - [x] **Step 56 – UI polish (status icons & branch colors)** (`commit: 68d7d4f`, `push: yes`)
    - Incorporate iconified statuses, colored branch badges, and table highlight styles inspired by Glow/LazyGit.
57. - [x] **Step 57 – Demo-only default config** (`commit: 45abce4`, `push: yes`)
    - Ship a demo-only `config.example.yaml` so the out-of-box experience shows data without credential errors.
58. - [x] **Step 58 – Demo CLI flag** (`commit: a3d1c4c`, `push: yes`)
    - Add `--demo` / `--demo-runs` flags so anyone can start the dashboard with synthetic runs without editing config.
59. - [x] **Step 59 – Demo make target** (`commit: 6c3594e`, `push: yes`)
    - Add `make demo` to run the CLI in demo mode (with docs) so validating the UI is one command.
60. - [x] **Step 60 – Demo diagnostics loop** (`commit: 6dcbc0d`, `push: yes`)
    - Provide `--demo-duration` / `--log-events` flags plus a `make snapshot` target that captures the TUI output for screenshots/logging.
61. - [x] **Step 61 – Provider health tracking** (`commit: 9cc5ca0`, `push: yes`)
    - Track last success/error + error counts per provider and display them in the status badges to make real providers easier to diagnose.
62. - [x] **Step 62 – Provider metrics/alerts** (`commit: 1ceb900`, `push: yes`)
    - Surface provider lag/failures in the UI (badges, notifications, or logs) and record metrics for slow polls.
63. - [x] **Step 63 – CI snapshots** (`commit: 25ef8b7`, `push: yes`)
    - Run `make snapshot` as part of CI so every push captures the latest demo screen + event log for regression evidence.
64. - [x] **Step 64 – Additional providers** (`commit: 3dd1b97`, `push: pending`)
    - Add support (config, factory, docs, tests) for a new provider such as GitLab so the dashboard can expand beyond GitHub/Azure.
65. - [x] **Step 65 – Accessibility/high-contrast theme** (`commit: 0b05288`, `push: yes`)
    - Add a high-contrast toggle, brighter colors, and new keyboard help so the UI stays readable on any background.
66. - [x] **Step 66 – Provider notifications** (`commit: d272644`, `push: pending`)
    - Show a transient alert when a provider errors so regressions are obvious without scanning logs.
67. - [x] **Step 67 – Desktop alerts** (`commit: a0e2a71`, `push: yes`)
    - Emit a desktop notification (via `osascript`/`notify-send`) whenever a provider reports an error so failures surface even if a terminal isn’t visible.
68. - [x] **Step 68 – Run history panel** (`commit: 733063d`, `push: pending`)
    - Track each provider’s last few run summaries and expose them in a dedicated sidebar panel for quick inspection.
69. - [x] **Step 69 – Persistent run history** (`commit: c9cdeac`, `push: pending`)
    - Persist recent run summaries to disk so history survives restarts and can be used for diagnostics later.
70. - [x] **Step 70 – Alerts channel** (`commit: 61c6e65`, `push: pending`)
    - Push provider failures to an external webhook (Slack/Teams/HTTP) so downstream systems can react automatically.
71. - [x] **Step 71 – Run filtering & tabs** (`commit: fba3d51`, `push: pending`)
    - Introduce per-provider tabs and status filters so users can focus on a single provider or status subset without scrolling.
72. - [x] **Step 72 – Provider detail view** (`commit: 629c4ff`, `push: pending`)
    - Allow expanding a provider or run row to reveal detailed information (timings, logs, quick actions) without leaving the UI.
73. - [x] **Step 73 – Alert log** (`commit: 21691fa`, `push: yes`)
    - Record each provider alert/webhook event in an “Alert log” panel so historical notifications are easy to review.
    - Introduce per-provider tabs and status filters so users can focus on a single provider or status subset without scrolling.
74. - [x] **Step 74 – Runtime provider management** (`commit: 564645b`, `push: yes`)
    - Add UI controls or CLI helpers so operators can add, update, or disable providers without editing the primary config file.
    - Persist provider metadata (credentials, filters, notification hooks) to disk so the dashboard can reload the dynamic list across restarts.
    - Write tests for the manager that ensure the in-memory provider registry reflects added/removed entries and the UI updates when the list changes.
75. - [x] **Step 75 – Provider store overlay** (`commit: f304034`, `push: yes`)
    - Surface runtime provider store data in a Bubble Tea overlay (key `P`), listing each stored entry with ID, friendly name, and enabled state.
    - Keep the overlay entries synchronized with the manager via `RunUpdatedMsg` payloads and ensure it handles keyboard navigation/closing.
    - Add tests that verify the overlay toggle key and message handling so the UI stays in sync even as new entries appear.
76. - [x] **Step 76 – Store overlay actions** (`commit: cd0ac1e`, `push: yes`)
    - Allow users to toggle a stored provider’s enabled state (`Space` in the overlay) without leaving the TUI.
    - Propagate manager-led changes back to the Bubble Tea model through `RunUpdatedMsg.Store` so the overlay stays current.
    - Expand tests and docs to describe the new keybinding and overlay behavior.
77. - [x] **Step 77 – Store overlay removals** (`commit: 99a0672`, `push: yes`)
    - Enable deleting stored providers directly from the overlay with a dedicated key (`d`), and refresh the overlay when removal completes.
    - Surface removal success/errors in the UI header/footer so operators know the store state changed.
    - Add tests covering the new keybinding and confirm the overlay rerenders after removal via `RunUpdatedMsg`.
78. - [x] **Step 78 – Store overlay edits** (`commit: d39509b`, `push: yes`)
    - Add an inline “edit” or “duplicate” shortcut in the overlay (e.g., press `e` on a stored entry) to spawn a temporary YAML snippet for editing, or trigger `ci-dash provider update` with the current config.
    - Ensure edits respect validation (provider type + required fields) and reuse the existing provider factory/state so changes immediately refresh the list.
    - Write tests that simulate the edit action and confirm the overlay/list updates after `RunUpdatedMsg` events reflecting the edited entry.
79. - [x] **Step 79 – Store overlay editing workflow** (`commit: adfe6e6`, `push: yes`)
    - Allow the overlay to edit a stored provider’s metadata (display name, etc.) inline and persist the change via the manager.
    - Validate edits and rerender the overlay with the new data so the UI and persisted store stay in sync.
    - Add tests that cover the edit UI and confirm `RunUpdatedMsg` updates refresh the overlay.
80. - [x] **Step 80 – Store overlay details** (`commit: 258a017`, `push: pending`)
    - Render provider-specific details (owner/repo, org/project, pipelines) beneath each entry in the overlay so operators can recall configuration at a glance.
    - Share helpers that build those detail strings so the overlay, logs, or future UI components can reuse the same summary logic.
    - Write tests for the helper to ensure detail strings include the expected data for each provider type.
79. - [x] **Step 79 – Store overlay editing workflow** (`commit: adfe6e6`, `push: yes`)
    - Allow the overlay to edit a stored provider’s fields inline (e.g., show the YAML snippet beside the list or open a modal that lets you update owner/repo/pipeline IDs).
    - Validate edits (type-specific required fields) using the same factory logic to prevent invalid configs, and reload the dashboard state once the change commits.
    - Add tests that simulate invoking the edit UI, feeding updates to the store via RunUpdatedMsg, and ensuring the overlay refreshes with the new values.

## Tracking Updates
When a step finishes, update its checklist line to `[x]`, replace `commit: pending` with the actual hash (e.g., `commit: abc1234`), and mark `push: yes` (or justify `push: no` if absolutely necessary). Add brief notes inline or append short bullet points under the step if context is useful for future reference.
