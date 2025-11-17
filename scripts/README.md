# ceye CI Monitoring Scripts

Scripts for monitoring and debugging the self-hosted GitHub Actions runner.

## Scripts

### debug-ci-runner.sh

Comprehensive CI debugging and monitoring tool.

**Commands:**
```bash
./Scripts/debug-ci-runner.sh status      # Show runner status, active jobs, processes
./Scripts/debug-ci-runner.sh tail        # Tail logs in real-time
./Scripts/debug-ci-runner.sh logs        # Dump all logs to tmp/ for analysis
./Scripts/debug-ci-runner.sh watch       # Watch test output with auto-refresh
./Scripts/debug-ci-runner.sh workflows   # Show recent workflow runs
./Scripts/debug-ci-runner.sh jobs        # Show active workflow jobs
./Scripts/debug-ci-runner.sh clean       # Clean up old logs
```

**What it shows:**
- Runner process status (PID, CPU, memory, uptime)
- Active worker processes
- Go build/test processes
- Workspace location and size
- Recent errors and test results
- Latest GitHub workflow status

**Example:**
```bash
$ ./Scripts/debug-ci-runner.sh status

🔍 ceye CI Runner Status
════════════════════════════════════════════════

📊 Latest GitHub Workflow Run:
  Run ID: 19418947923
  Workflow: Comprehensive Tests
  Status: in_progress
  Branch: main

🏃 Runner Process:
  ✓ Running (PID: 82195)
  CPU: 0.0% | Memory: 0.1% | Uptime: 01:42:24

⚙️  Active Workers:
  ✓ Worker PID: 8441
    CPU: 0.0% | Memory: 0.2% | Runtime: 03:21

🔨 Go Build/Test Processes:
  ✓ go test ./...
    CPU: 95.2% | Memory: 2.3% | Runtime: 00:45

📊 Test Results:
  ✓ Passed: 12 packages
  ✗ Failed: 0 packages
```

### monitor-ci-live.sh

Real-time monitoring of CI workflow execution.

**Usage:**
```bash
./Scripts/monitor-ci-live.sh
```

**What it does:**
- Monitors Go test output in real-time
- Detects and highlights errors immediately
- Shows test pass/fail counts
- Alerts when output stalls
- Logs errors to tmp/ci-errors-*.log

**Example output:**
```
🔍 Live CI Monitor - Checking every 5 seconds
Watching: ceye repository workflows
Press Ctrl+C to stop

✅ 12:34:56 - 3 test package(s) passed
[12:34:56] Lines: 450 → 478 (+28)
🚨 ERROR DETECTED at 12:35:01
    FAIL: TestStoreMerge (0.00s)
```

## Quick Reference

### Check if runner is working
```bash
./Scripts/debug-ci-runner.sh status
```

### Watch tests in real-time
```bash
./Scripts/debug-ci-runner.sh watch
```

### Dump logs for debugging
```bash
./Scripts/debug-ci-runner.sh logs
```

### View workflow history
```bash
./Scripts/debug-ci-runner.sh workflows
```

### Monitor live CI execution
```bash
./Scripts/monitor-ci-live.sh
```

## Log Locations

- **Runner logs:** `~/actions-runner-ceye/_diag/Runner_*.log`
- **Worker logs:** `~/actions-runner-ceye/_diag/Worker_*.log`
- **Test output:** `~/actions-runner-ceye/_work/ceye/ceye/tmp/go-test.log`
- **Build output:** `~/actions-runner-ceye/_work/ceye/ceye/tmp/go-build.log`
- **Dumped logs:** `./tmp/ci-debug-TIMESTAMP/`

## Troubleshooting

### Runner not running jobs?
```bash
# Check status
./Scripts/debug-ci-runner.sh status

# If not running, start it
cd ~/actions-runner-ceye && ./run.sh
```

### Tests failing but logs unclear?
```bash
# Dump all logs for analysis
./Scripts/debug-ci-runner.sh logs

# Check tmp/ci-debug-*/
```

### Want to see what's happening right now?
```bash
# Real-time test output
./Scripts/debug-ci-runner.sh watch

# Or monitor for errors
./Scripts/monitor-ci-live.sh
```

### Clean up old logs?
```bash
./Scripts/debug-ci-runner.sh clean
```

## Tips

1. **Use `watch` for long-running tests** - It shows only new output as it appears
2. **Use `monitor-ci-live` for error detection** - It highlights errors immediately
3. **Use `logs` when debugging** - Captures everything in one place
4. **Use `status` for quick checks** - See what's running at a glance

## Integration with ceye

These scripts are designed to work with the ceye self-hosted runner setup:
- Runner location: `~/actions-runner-ceye`
- Repository: `github.com/joelklabo/ceye`
- Workflows: All 8 workflows with 35 jobs
- Tests: Go test suite (175+ tests)

## Dependencies

- **Required:** bash, grep, awk
- **Optional:** jq (for pretty JSON output)
- **Optional:** gh CLI (for GitHub workflow status)

Install optional dependencies:
```bash
brew install jq gh
```

## Ported from BigTimer

These scripts were adapted from the BigTimer project's CI monitoring tools, converted from iOS/Xcode monitoring to Go/GitHub Actions monitoring.
