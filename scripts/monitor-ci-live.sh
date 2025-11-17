#!/bin/bash
# Real-time CI monitoring for ceye self-hosted runner
# Monitors Go test output, build logs, and workflow progress

RUNNER_DIR="$HOME/actions-runner-ceye"
WORKSPACE_DIR="$RUNNER_DIR/_work/ceye/ceye"
GO_TEST_LOG="$WORKSPACE_DIR/tmp/go-test.log"
BUILD_LOG="$WORKSPACE_DIR/tmp/go-build.log"
ERROR_LOG="$(pwd)/tmp/ci-errors-$(date +%s).log"

echo "🔍 Live CI Monitor - Checking every 5 seconds"
echo "Watching: ceye repository workflows"
echo "Error log: $ERROR_LOG"
echo "Press Ctrl+C to stop"
echo ""

LAST_LINE_COUNT=0
LAST_CHECK_TIME=$(date +%s)

while true; do
    CURRENT_TIME=$(date +%s)
    
    # Check Go test log
    if [ -f "$GO_TEST_LOG" ]; then
        CURRENT_LINE_COUNT=$(wc -l < "$GO_TEST_LOG" 2>/dev/null || echo "0")
        
        # Check if new lines appeared
        if [ $CURRENT_LINE_COUNT -gt $LAST_LINE_COUNT ]; then
            # Get new lines since last check
            NEW_LINES=$((CURRENT_LINE_COUNT - LAST_LINE_COUNT))
            tail -n $NEW_LINES "$GO_TEST_LOG" > /tmp/new_output.txt
            
            # Check for errors in new output
            if grep -qiE "FAIL|panic|fatal|error:" /tmp/new_output.txt; then
                echo "🚨 ERROR DETECTED at $(date '+%H:%M:%S')"
                grep -iE "FAIL|panic|fatal|error:" /tmp/new_output.txt | tee -a "$ERROR_LOG"
                echo ""
            fi
            
            # Check for test passes
            if grep -q "^PASS" /tmp/new_output.txt; then
                PASSED=$(grep "^PASS" /tmp/new_output.txt | wc -l)
                echo "✅ $(date '+%H:%M:%S') - $PASSED test package(s) passed"
            fi
            
            # Show progress
            echo "[$(date '+%H:%M:%S')] Lines: $LAST_LINE_COUNT → $CURRENT_LINE_COUNT (+$NEW_LINES)"
            
            LAST_LINE_COUNT=$CURRENT_LINE_COUNT
            LAST_CHECK_TIME=$CURRENT_TIME
        else
            # No new output - check if stuck
            IDLE_TIME=$((CURRENT_TIME - LAST_CHECK_TIME))
            if [ $IDLE_TIME -gt 120 ]; then  # 2 minutes no output
                echo "[$(date '+%H:%M:%S')] ⚠️  No new output for ${IDLE_TIME}s (line count: $CURRENT_LINE_COUNT)"
                
                # Check if go test is still running
                if ! pgrep -f "go test" > /dev/null; then
                    echo "⚠️  go test process has exited!"
                    break
                fi
            fi
        fi
    else
        # Check workspace activity
        if [ -d "$WORKSPACE_DIR" ]; then
            echo "[$(date '+%H:%M:%S')] Workspace exists, waiting for test output..."
            
            # Show what's currently running
            if pgrep -f "go build" > /dev/null; then
                echo "  → go build is running"
            fi
            if pgrep -f "go test" > /dev/null; then
                echo "  → go test is running"
            fi
            if pgrep -f "golangci-lint" > /dev/null; then
                echo "  → golangci-lint is running"
            fi
        else
            echo "[$(date '+%H:%M:%S')] Waiting for workflow to start..."
        fi
    fi
    
    sleep 5
done

echo ""
echo "Monitor stopped. Check errors in: $ERROR_LOG"
