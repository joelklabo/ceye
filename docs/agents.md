## Agent Installation Notes

To keep the CLI on your `PATH` up to date, rebuild and reinstall after any code changes:

```bash
cd /Users/honk/code/ceye
go build -o bin/ci-dash ./cmd/ci-dash
sudo cp bin/ci-dash /usr/local/bin/ci-dash
```

This ensures the command you run from `~/code` (or anywhere else) executes the latest binary that includes config discovery, provider store overlays, and all the recent work.
