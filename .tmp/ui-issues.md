# UI Issues & Improvements - Analysis

**Date**: 2025-11-16
**Status**: Issues Identified

## TUI (Terminal UI) Issues

### 1. **Line Wrapping/Truncation in Activity Panel**
- Activity messages are wrapped poorly: "invalid character '<' looking for beginning" shows on multiple lines
- Makes the activity panel hard to read
- **Fix**: Truncate long messages with ellipsis or improve wrapping logic

### 2. **Provider Errors Spam Activity Panel**  
- Every repo without workflows shows "gh run list: exit status 1"
- Azure errors repeat every poll cycle
- Creates noise, makes real issues hard to spot
- **Fix**: 
  - Rate-limit duplicate errors (show once every N minutes)
  - Or add error suppression for expected failures (repos without workflows)
  - Consider a separate "Errors" tab/panel

### 3. **"ci-dash" Still in Logs**
- Startup logs show "ci-dash: using gh CLI-based GitHub client"
- Should be "ceye:" after rename
- **Fix**: Update log prefix in cmd/ceye/main.go

### 4. **No Run Data Visible**
- Despite "172 runs refreshed", no runs shown in table
- Provider Health shows "healthy" but Activity shows errors
- Unclear if this is a filter issue or data issue
- **Fix**: Need to investigate why runs aren't displaying

### 5. **Activity Panel Size**
- Only shows 2-3 messages at a time
- Could be taller or scrollable
- **Fix**: Make activity panel scrollable or show more lines

### 6. **Status Bar at Bottom**
- Very long status bar text wraps awkwardly
- "tab cycle providers • f cycle status • t cycle sort..." 
- Consider multi-line or grouped shortcuts
- **Fix**: Format help text better, maybe show only most important shortcuts

## Web UI Issues

### 1. **No /api/runs Endpoint**
- API returns 404 for `/api/runs`
- Web UI expects this but it doesn't exist
- Only `/ws` (WebSocket) and `/api/analytics/trends` exist
- **Fix**: Either:
  - Add REST API endpoint for runs
  - Or document that web UI is WebSocket-only

### 2. **WebSocket Connection Status Unknown**
- Couldn't verify WebSocket is working (browser not accessible)
- Need to test:  
  - Does WS connect?
  - Do runs update in real-time?
  - Does provider health show?

### 3. **Title Still "ceye" Not "CI Eye"**
- HTML title is just "ceye"  
- Could be more descriptive: "CI Eye - CI/CD Dashboard"
- **Fix**: Update title tag

### 4. **Error Logging Too Verbose**
- Web server logs every repo error
- Same as TUI issue - too noisy
- **Fix**: Reduce logging verbosity for expected errors

## Positive Observations

✅ **TUI Layout** - Clean, organized panels
✅ **Provider Health Panel** - Clear status indicators
✅ **Trends Panel** - Good info presentation (when data available)
✅ **Web UI HTML** - Clean, semantic structure
✅ **Stats Cards** - Nice visual summary
✅ **Responsive Design** - Web UI has proper viewport meta tag

## Priority Fixes

**High Priority:**
1. Fix "ci-dash" → "ceye" in logs
2. Reduce error spam (rate limit duplicates)
3. Verify why runs aren't displaying despite being fetched
4. Test WebSocket connection in web UI

**Medium Priority:**
5. Improve Activity panel message formatting
6. Add /api/runs endpoint or clarify WS-only
7. Make Activity panel taller/scrollable

**Low Priority:**
8. Update HTML title
9. Improve status bar formatting
10. Consider error suppression options in config

## Next Steps

1. Fix log prefix (quick win)
2. Debug why runs aren't showing in TUI table
3. Test web UI with actual browser access
4. Implement error rate limiting
5. Improve activity message truncation
