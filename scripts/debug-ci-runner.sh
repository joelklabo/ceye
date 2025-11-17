#!/bin/bash
# Debug CI Runner - Comprehensive local CI debugging tool for ceye
# Monitors self-hosted runner activity, logs, and workflow status

set -euo pipefail

RUNNER_DIR="$HOME/actions-runner-ceye"
WORKSPACE_DIR="$RUNNER_DIR/_work/ceye/ceye"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

usage() {
    cat << EOF
Usage: $(basename "$0") [COMMAND]

Commands:
    status      Show current CI run status and runner health
    tail        Tail runner and build logs in real-time
    logs        Dump recent logs to tmp/ for analysis
    watch       Watch CI progress with auto-refresh
    clean       Clean up stale log files
    workflows   Show recent workflow runs
    jobs        Show active jobs
    help        Show this help message

Examples:
    $(basename "$0") status       # Quick status check
    $(basename "$0") tail         # Follow logs live
    $(basename "$0") watch        # Auto-refreshing dashboard
    $(basename "$0") workflows    # Recent workflow history
EOF
}

get_latest_logs() {
    WORKER_LOG=$(ls -t "$RUNNER_DIR/_diag/Worker_"*.log 2>/dev/null | head -1)
    RUNNER_LOG=$(ls -t "$RUNNER_DIR/_diag/Runner_"*.log 2>/dev/null | head -1)
    GO_TEST_LOG="$WORKSPACE_DIR/tmp/go-test.log"
    GO_BUILD_LOG="$WORKSPACE_DIR/tmp/go-build.log"
}

show_status() {
    echo -e "${BLUE}🔍 ceye CI Runner Status${NC}"
    echo "════════════════════════════════════════════════"
    echo ""
    
    # GitHub CI Status
    echo -e "${YELLOW}📊 Latest GitHub Workflow Run:${NC}"
    if command -v gh &> /dev/null; then
        cd "$PROJECT_DIR"
        if command -v jq &> /dev/null; then
            gh run list --limit 1 --json conclusion,status,name,databaseId,createdAt,event,headBranch 2>/dev/null | \
                jq -r '.[] | "  Run ID: \(.databaseId)\n  Workflow: \(.name)\n  Status: \(.status)\n  Conclusion: \(.conclusion // "N/A")\n  Branch: \(.headBranch)\n  Event: \(.event)\n  Created: \(.createdAt)"' || echo "  No recent runs"
        else
            gh run list --limit 1 2>/dev/null || echo "  No recent runs"
        fi
    else
        echo "  gh CLI not installed"
    fi
    echo ""
    
    # Runner Process Status
    echo -e "${YELLOW}🏃 Runner Process:${NC}"
    if pgrep -f "Runner.Listener" > /dev/null; then
        RUNNER_PID=$(pgrep -f "Runner.Listener" | head -1)
        echo -e "  ${GREEN}✓${NC} Running (PID: $RUNNER_PID)"
        ps -p "$RUNNER_PID" -o %cpu,%mem,etime 2>/dev/null | tail -1 | awk '{print "  CPU: " $1 "% | Memory: " $2 "% | Uptime: " $3}'
    else
        echo -e "  ${RED}✗${NC} Not running"
        echo "  Start with: cd ~/actions-runner-ceye && ./run.sh"
    fi
    echo ""
    
    # Worker Status
    echo -e "${YELLOW}⚙️  Active Workers:${NC}"
    WORKER_PIDS=$(pgrep -f "Runner.Worker" 2>/dev/null || true)
    if [ -n "$WORKER_PIDS" ]; then
        echo "$WORKER_PIDS" | while read pid; do
            echo -e "  ${GREEN}✓${NC} Worker PID: $pid"
            ps -p "$pid" -o %cpu,%mem,etime 2>/dev/null | tail -1 | awk '{print "    CPU: " $1 "% | Memory: " $2 "% | Runtime: " $3}'
        done
    else
        echo "  No active workers"
    fi
    echo ""
    
    # Go Process Status
    echo -e "${YELLOW}🔨 Go Build/Test Processes:${NC}"
    if pgrep -f "go test\|go build\|golangci-lint" > /dev/null 2>&1; then
        GO_PIDS=$(pgrep -f "go test\|go build\|golangci-lint" 2>/dev/null || true)
        if [ -n "$GO_PIDS" ]; then
            echo "$GO_PIDS" | while read pid; do
                CMD=$(ps -p "$pid" -o command= 2>/dev/null | head -c 60)
                echo -e "  ${GREEN}✓${NC} $CMD"
                ps -p "$pid" -o %cpu,%mem,etime 2>/dev/null | tail -1 | awk '{print "    CPU: " $1 "% | Memory: " $2 "% | Runtime: " $3}'
            done
        fi
    else
        echo "  No Go processes running"
    fi
    echo ""
    
    # Workspace Status
    echo -e "${YELLOW}📂 Workspace:${NC}"
    if [ -d "$WORKSPACE_DIR" ]; then
        echo -e "  ${GREEN}✓${NC} $WORKSPACE_DIR"
        WORKSPACE_SIZE=$(du -sh "$WORKSPACE_DIR" 2>/dev/null | cut -f1)
        echo "  Size: $WORKSPACE_SIZE"
    else
        echo "  No active workspace"
    fi
    echo ""
    
    # Log Files
    get_latest_logs
    echo -e "${YELLOW}📋 Log Files:${NC}"
    [ -f "$WORKER_LOG" ] && echo "  Worker: $(basename "$WORKER_LOG") ($(du -h "$WORKER_LOG" | cut -f1))"
    [ -f "$RUNNER_LOG" ] && echo "  Runner: $(basename "$RUNNER_LOG") ($(du -h "$RUNNER_LOG" | cut -f1))"
    [ -f "$GO_TEST_LOG" ] && echo "  Tests:  $(basename "$GO_TEST_LOG") ($(du -h "$GO_TEST_LOG" | cut -f1))"
    [ -f "$GO_BUILD_LOG" ] && echo "  Build:  $(basename "$GO_BUILD_LOG") ($(du -h "$GO_BUILD_LOG" | cut -f1))"
    echo ""
    
    # Recent Errors
    if [ -f "$WORKER_LOG" ]; then
        ERROR_COUNT=$(grep -c "ERR\|ERROR\|FAIL" "$WORKER_LOG" 2>/dev/null || echo "0")
        if [ "$ERROR_COUNT" -gt 0 ]; then
            echo -e "${RED}⚠️  Recent Errors: $ERROR_COUNT${NC}"
            echo "  Run '$(basename "$0") logs' to see details"
            echo ""
        fi
    fi
    
    # Test Results Summary
    if [ -f "$GO_TEST_LOG" ]; then
        PASS_COUNT=$(grep -c "^PASS" "$GO_TEST_LOG" 2>/dev/null || echo "0")
        FAIL_COUNT=$(grep -c "^FAIL" "$GO_TEST_LOG" 2>/dev/null || echo "0")
        if [ "$PASS_COUNT" -gt 0 ] || [ "$FAIL_COUNT" -gt 0 ]; then
            echo -e "${CYAN}📊 Test Results:${NC}"
            [ "$PASS_COUNT" -gt 0 ] && echo -e "  ${GREEN}✓${NC} Passed: $PASS_COUNT packages"
            [ "$FAIL_COUNT" -gt 0 ] && echo -e "  ${RED}✗${NC} Failed: $FAIL_COUNT packages"
            echo ""
        fi
    fi
}

tail_logs() {
    get_latest_logs
    
    echo -e "${BLUE}📜 Tailing CI Logs (Ctrl+C to exit)${NC}"
    echo "════════════════════════════════════════════════"
    echo ""
    
    # Collect available logs
    LOGS_TO_TAIL=()
    [ -f "$WORKER_LOG" ] && LOGS_TO_TAIL+=("$WORKER_LOG")
    [ -f "$GO_TEST_LOG" ] && LOGS_TO_TAIL+=("$GO_TEST_LOG")
    [ -f "$GO_BUILD_LOG" ] && LOGS_TO_TAIL+=("$GO_BUILD_LOG")
    
    if [ ${#LOGS_TO_TAIL[@]} -eq 0 ]; then
        echo -e "${YELLOW}⏳ No logs found yet, waiting...${NC}"
        # Wait for logs to appear
        for i in {1..30}; do
            get_latest_logs
            [ -f "$WORKER_LOG" ] && LOGS_TO_TAIL+=("$WORKER_LOG") && break
            sleep 1
        done
    fi
    
    if [ ${#LOGS_TO_TAIL[@]} -gt 0 ]; then
        tail -f "${LOGS_TO_TAIL[@]}"
    else
        echo -e "${RED}No log files found${NC}"
        exit 1
    fi
}

dump_logs() {
    get_latest_logs
    
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    OUTPUT_DIR="$PROJECT_DIR/tmp/ci-debug-$TIMESTAMP"
    mkdir -p "$OUTPUT_DIR"
    
    echo -e "${BLUE}💾 Dumping CI Logs to tmp/${NC}"
    echo "════════════════════════════════════════════════"
    echo ""
    
    # Copy runner logs
    if [ -f "$WORKER_LOG" ]; then
        cp "$WORKER_LOG" "$OUTPUT_DIR/worker.log"
        echo "✓ Copied worker log"
    fi
    
    if [ -f "$RUNNER_LOG" ]; then
        cp "$RUNNER_LOG" "$OUTPUT_DIR/runner.log"
        echo "✓ Copied runner log"
    fi
    
    # Copy Go logs
    if [ -f "$GO_TEST_LOG" ]; then
        cp "$GO_TEST_LOG" "$OUTPUT_DIR/go-test.log"
        echo "✓ Copied Go test log"
    fi
    
    if [ -f "$GO_BUILD_LOG" ]; then
        cp "$GO_BUILD_LOG" "$OUTPUT_DIR/go-build.log"
        echo "✓ Copied Go build log"
    fi
    
    # Extract errors
    if [ -f "$WORKER_LOG" ]; then
        grep -i "ERR\|ERROR\|FAIL\|WARN" "$WORKER_LOG" > "$OUTPUT_DIR/errors.log" 2>/dev/null || true
        echo "✓ Extracted errors"
    fi
    
    if [ -f "$GO_TEST_LOG" ]; then
        grep -E "FAIL|panic|fatal" "$GO_TEST_LOG" > "$OUTPUT_DIR/test-failures.log" 2>/dev/null || true
        echo "✓ Extracted test failures"
    fi
    
    # Get GitHub workflow status
    cd "$PROJECT_DIR"
    gh run list --limit 5 --json conclusion,status,name,databaseId,createdAt 2>/dev/null > "$OUTPUT_DIR/gh-runs.json" || true
    echo "✓ Saved GitHub run status"
    
    # Process list
    ps aux | grep -E "Runner|go test|go build|golangci" | grep -v grep > "$OUTPUT_DIR/processes.txt" || true
    echo "✓ Saved process list"
    
    # System info
    {
        echo "=== System Info ==="
        uname -a
        echo ""
        echo "=== Go Version ==="
        go version 2>/dev/null || echo "N/A"
        echo ""
        echo "=== Disk Space ==="
        df -h "$RUNNER_DIR"
        echo ""
        echo "=== Runner Config ==="
        cat "$RUNNER_DIR/.runner" 2>/dev/null || echo "N/A"
        echo ""
        echo "=== Go Env ==="
        go env 2>/dev/null || echo "N/A"
    } > "$OUTPUT_DIR/system-info.txt"
    echo "✓ Saved system info"
    
    echo ""
    echo -e "${GREEN}Logs dumped to: $OUTPUT_DIR${NC}"
    echo ""
    echo "Files created:"
    ls -lh "$OUTPUT_DIR" | tail -n +2
}

watch_ci() {
    trap "echo ''; exit 0" INT TERM
    
    echo "🔍 CI Monitor Started - $(date '+%Y-%m-%d %H:%M:%S')"
    echo "Watching Go test/build output (new lines appear below)"
    echo "Press Ctrl+C to exit"
    echo "════════════════════════════════════════════════════════════════"
    echo ""
    
    get_latest_logs
    LAST_LINE_COUNT=0
    
    if [ -f "$GO_TEST_LOG" ]; then
        LAST_LINE_COUNT=$(wc -l < "$GO_TEST_LOG")
        echo "[$(date '+%H:%M:%S')] Starting from line $LAST_LINE_COUNT"
    fi
    
    while true; do
        get_latest_logs
        
        if [ -f "$GO_TEST_LOG" ]; then
            CURRENT_LINE_COUNT=$(wc -l < "$GO_TEST_LOG")
            
            if [ $CURRENT_LINE_COUNT -gt $LAST_LINE_COUNT ]; then
                # New output detected - show only new lines
                NEW_LINES=$((CURRENT_LINE_COUNT - LAST_LINE_COUNT))
                echo ""
                echo -e "${GREEN}[$(date '+%H:%M:%S')] +$NEW_LINES new lines${NC}"
                echo "────────────────────────────────────────────────────────────────"
                tail -n $NEW_LINES "$GO_TEST_LOG"
                echo "────────────────────────────────────────────────────────────────"
                LAST_LINE_COUNT=$CURRENT_LINE_COUNT
            fi
        else
            if [ $LAST_LINE_COUNT -eq 0 ]; then
                echo "[$(date '+%H:%M:%S')] Waiting for test output..."
                LAST_LINE_COUNT=-1  # Mark as waiting
            fi
        fi
        
        sleep 5
    done
}

show_workflows() {
    echo -e "${BLUE}📋 Recent Workflow Runs${NC}"
    echo "════════════════════════════════════════════════"
    echo ""
    
    cd "$PROJECT_DIR"
    gh run list --limit 10 2>/dev/null || echo "No workflow runs found"
}

show_jobs() {
    echo -e "${BLUE}🔧 Active Workflow Jobs${NC}"
    echo "════════════════════════════════════════════════"
    echo ""
    
    cd "$PROJECT_DIR"
    
    # Get latest run
    LATEST_RUN=$(gh run list --limit 1 --json databaseId --jq '.[0].databaseId' 2>/dev/null)
    
    if [ -n "$LATEST_RUN" ]; then
        echo "Latest run: $LATEST_RUN"
        echo ""
        gh run view "$LATEST_RUN" 2>/dev/null || echo "Cannot fetch job details"
    else
        echo "No active runs"
    fi
}

clean_logs() {
    echo -e "${BLUE}🧹 Cleaning Stale Logs${NC}"
    echo "════════════════════════════════════════════════"
    echo ""
    
    # Clean tmp directory logs
    find "$PROJECT_DIR/tmp" -name "ci-debug-*" -mtime +7 -exec rm -rf {} \; 2>/dev/null || true
    find "$PROJECT_DIR/tmp" -name "*.log" -mtime +7 -delete 2>/dev/null || true
    echo "✓ Cleaned tmp/ logs older than 7 days"
    
    # Show current tmp usage
    echo ""
    echo "Current tmp/ usage:"
    du -sh "$PROJECT_DIR/tmp" 2>/dev/null || echo "No tmp directory"
}

# Main
case "${1:-status}" in
    status)
        show_status
        ;;
    tail)
        tail_logs
        ;;
    logs)
        dump_logs
        ;;
    watch)
        watch_ci
        ;;
    workflows)
        show_workflows
        ;;
    jobs)
        show_jobs
        ;;
    clean)
        clean_logs
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        echo ""
        usage
        exit 1
        ;;
esac
