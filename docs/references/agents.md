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

## UI Testing Strategy

When making UI/TUI changes (layout, rendering, styling), **always verify the changes yourself** before completing the task. Never ask the user to confirm something you can verify independently.

### Required Testing Loop

1. **Build the binary**
   ```bash
   cd /Users/honk/code/ceye
   go build -o bin/ci-dash ./cmd/ci-dash
   ```

2. **Start in tmux** (allows capture without user interaction)
   ```bash
   tmux kill-session -t ci-dash-live 2>/dev/null
   tmux new-session -d -s ci-dash-live "cd /Users/honk/code/ceye && ./bin/ci-dash"
   sleep 3  # Give it time to initialize
   ```

3. **Capture and verify output**
   ```bash
   # Capture visible pane
   tmux capture-pane -t ci-dash-live -p | head -50
   
   # Capture with more history if needed
   tmux capture-pane -t ci-dash-live -p -S -40 | head -50
   
   # Check terminal dimensions
   tmux display-message -t ci-dash-live -p '#{pane_width}x#{pane_height}'
   ```

4. **Test with demo mode** (when no real data needed)
   ```bash
   ./bin/ci-dash --demo --demo-duration 5s 2>&1 | cat
   ```

5. **Verify specific issues**
   - Check for text overlapping/wrapping incorrectly
   - Verify column alignment in tables
   - Confirm truncation works (ellipsis appears correctly)
   - Test in standard 80x24 terminal size
   - Test in wider terminals (150+ columns) if responsive layout matters
   - Ensure ANSI color codes don't break width calculations

### Common UI Issues to Check

- **Text overflow**: Does long text get truncated with "…"?
- **Column alignment**: Do table columns line up properly?
- **Status rendering**: Are status indicators (✓ ✗ ▸ …) displaying correctly?
- **Width calculation**: Are ANSI escape codes causing width issues? (Remove styling from cells if needed)
- **Terminal size**: Test in 80-column terminal (default tmux) and wider
- **Border rendering**: Are boxes/panels aligned and not broken?

### Providing Results to User

After testing, provide:
1. The tmux connection command: `tmux attach -t ci-dash-live`
2. Confirmation of what was verified (specific issues fixed)
3. Only commit and push after successful verification

### Example Test Session

```bash
# Build
cd /Users/honk/code/ceye && go build -o bin/ci-dash ./cmd/ci-dash

# Start and verify
tmux new-session -d -s ci-dash-live "cd /Users/honk/code/ceye && ./bin/ci-dash"
sleep 3
tmux capture-pane -t ci-dash-live -p | head -40

# If output looks wrong, fix code and repeat
# If output looks correct, commit and provide connection command
```

**Key principle**: If you can observe it programmatically, do not ask the user to verify it. Use the tools available.
