## Agent Installation Notes

To keep the CLI on your `PATH` up to date, rebuild and reinstall after any code changes:

```bash
cd /Users/honk/code/ceye
go build -o bin/ci-dash ./cmd/ci-dash
sudo cp bin/ci-dash /usr/local/bin/ci-dash
```

This ensures the command you run from `~/code` (or anywhere else) executes the latest binary that includes config discovery, provider store overlays, and all the recent work.

Make this the first step after pulling or editing the repo so `ci-dash` on your `PATH` never lags behind the code inside `~/code/ceye`. Repeat the commands above each time you pull or modify the CLI.

Always verify in the workspace root (e.g. `~/code`) before declaring the work done—attach to tmux if needed and capture a screenshot showing the real provider data alongside the updated version info. Never claim completion until you've personally confirmed the latest build behaves as expected.
