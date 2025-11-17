# ceye Development Plan

**Last Updated**: 2025-11-17 21:19 UTC  
**Status**: Phase 0.7 Critical Issues - 🟢 **12 of 20 COMPLETE**

## Current Status

**React Migration**: Complete ✅  
**Test Suite**: Go ✅ ALL PASSING | Playwright: 50/58 passing (8 failing)  
**Status**: ✅ **ONLINE - WebSocket FIXED!**

The React dashboard is now working! WebSocket connection was fixed by removing infinite reconnection loop.

**Stack**:
- React 19 + Vite + TypeScript
- Tailwind CSS v3 + Framer Motion
- Real-time WebSocket integration (WORKING ✅)
- 27 Playwright integration tests passing

**🚨 REMAINING ISSUES**:
- 🟡 Provider Health UI needs full-width redesign + webhook indicators
- 🟡 Can't view webhook payloads in UI
- 🟡 Activity Feed needs enhanced message details
- 🟡 UI flicker on updates (needs investigation)
- 🟡 Debug Panel - Event Timeline Visualization
- 🟡 Debug Panel - State Inspector
- 🟡 Debug Panel - Performance Profiler
- 🟡 Debug Panel - Webhook Simulator
- 🟡 Build Failure Notification Use Case

---

## 🚧 Active Tasks

### Phase 0.7: Critical Bug Fixes & Cleanup - 🔴 **IN PROGRESS**

**Goal**: Fix broken WebSocket connection and clean up codebase

**Priority**: CRITICAL - Validate webhook functionality!

**Tasks**:

#### 0.0.1. Fix All Test Failures (CRITICAL 🔴) - 🔄 **IN PROGRESS** - 4-6 hours (Commit: 0e1fdd3)

**Problem**: Tests are failing locally and in CI
- Go tests: ALL PASSING ✅ (fixed by earlier commits)
- Playwright tests: 8 failing, 50 passing, 14 skipped

**Root Causes**:
1. **Missing data-testid attributes** - React migration removed test IDs
2. **Old test selectors** - Tests written for static HTML, now using React
3. **Changed component structure** - Elements moved/renamed during migration

**Current Status**: 10 failing tests (increased from 8 after adding data-testid)

**Detailed Test-by-Test Analysis**:

**Group A: Old Static HTML Tests (3 tests) - DELETE/REWRITE**
These tests were written for the old static HTML dashboard and need complete rewrite:

1. **websocket-connection.spec.js:4** - `#connectionStatus`, `#lastUpdate` elements
   - **Problem**: Looking for `#connectionStatus` and `#lastUpdate` DOM IDs that don't exist in React
   - **Solution**: DELETE this file - functionality covered by websocket-critical.spec.ts
   
2. **websocket-connection.spec.js:28** - `.connection-indicator.connected` class
   - **Problem**: Looking for specific CSS classes from static HTML
   - **Solution**: DELETE - redundant with websocket-critical.spec.ts
   
3. **websocket-connection.spec.js:55** - `#activityToggle` element
   - **Problem**: Activity log was removed in React migration
   - **Solution**: DELETE - no equivalent in React app

**Group B: Provider Health Tests (3 tests) - FIX SELECTORS**
Tests are looking for wrong elements or have flaky selectors:

4. **provider-health-ui.spec.ts:10** - Full-width layout check
   - **Problem**: `[class*="space-y"]` selector too generic, finds multiple elements
   - **Solution**: Target specific ProviderCards container with data-testid
   
5. **provider-health-ui.spec.ts:33** - Provider with health indicator  
   - **Problem**: `text=github` fails because provider is "demo" in test mode
   - **Solution**: Use `text=demo` or generic provider selector
   
6. **provider-health-ui.spec.ts:107** - Match Activity feed width
   - **Problem**: `.locator('..')` parent selector is fragile
   - **Solution**: Add data-testid to both containers, compare directly

**Group C: React Dashboard Tests (2 tests) - FIX SVG SELECTOR**

7. **dashboard-react.spec.ts:70** - Refresh button with SVG
   - **Problem**: `svg[data-lucide="refresh-cw"]` attribute doesn't exist
   - **Solution**: Use simpler selector or add data-testid to button

**Group D: Activity Feed Tests (1 test) - FIX ASSERTION**

8. **activity-feed-enhanced.spec.ts:36** - Duration display
   - **Problem**: Test expects "Duration:" but we hide it when 0 seconds
   - **Solution**: Update test to handle both cases OR ensure test data has duration

**Group E: WebSocket Critical Tests (2 tests) - FIXED BY data-testid**

9. **websocket-critical.spec.ts:60** - ✅ SHOULD BE FIXED (added data-testid="activity-feed")
10. **websocket-critical.spec.ts:79** - ✅ SHOULD BE FIXED (added data-testid="stat-*")

**Note**: Tests 9-10 should now pass with commit 0e1fdd3. Need to verify.

**Detailed Fix Plan** (6-8 hours total):

**Phase 1: Add Missing data-testid Attributes** ✅ DONE (Commit: 0e1fdd3)
- [x] ActivityFeed.tsx - Added `data-testid="activity-feed"`
- [x] StatsCards.tsx - Added test IDs to all stat cards
- [x] Should fix tests 9-10 (websocket-critical.spec.ts)

**Phase 2: Delete Obsolete Static HTML Tests** (30 minutes)
- [ ] Delete `e2e/websocket-connection.spec.js` entirely (fixes tests 1-3)
- [ ] Verify coverage exists in websocket-critical.spec.ts
- [ ] Commit: "test: Remove obsolete static HTML WebSocket tests"

**Phase 3: Fix Provider Health Test Selectors** (1.5 hours)
- [ ] **Test 4** - Add data-testid="provider-cards" to ProviderCards container
  - Update test to use `[data-testid="provider-cards"]` instead of `[class*="space-y"]`
  
- [ ] **Test 5** - Fix provider name selector
  - Change `text=github` to `text=demo` (demo mode provider name)
  - Or use `.capitalize()` to accept any provider name
  
- [ ] **Test 6** - Add container test IDs
  - Add data-testid="provider-cards-container" to ProviderCards
  - Add data-testid="activity-feed-container" to ActivityFeed
  - Compare widths using testid selectors
  
- [ ] Commit: "test: Fix Provider Health UI test selectors"

**Phase 4: Fix React Dashboard Refresh Button Test** (45 minutes)
- [ ] **Test 7** - Option A: Add data-testid="provider-refresh-button"
  - OR Option B: Simplify selector to find RefreshCw icon by className
  - Update test to use new selector
- [ ] Commit: "test: Fix provider health refresh button test"

**Phase 5: Fix Activity Feed Duration Test** (45 minutes)
- [ ] **Test 8** - Option A: Update test to handle "Duration: 0s" being hidden
  - Check if duration exists, if not, verify it's a queued/starting run
  - OR Option B: Ensure demo data always has duration > 0
- [ ] Commit: "test: Fix activity feed duration display test"

**Phase 6: Verification & Cleanup** (2 hours)
- [ ] Run all tests 3 times to check for flakiness
- [ ] Fix any flaky tests found
- [ ] Verify websocket-critical tests now pass (data-testid changes)
- [ ] Update test counts in plan.md
- [ ] Document any remaining skipped tests with reasons
- [ ] Commit: "test: All Playwright tests passing"

**Phase 7: CI Validation** (1 hour)
- [ ] Push all changes to trigger CI
- [ ] Monitor CI test runs
- [ ] Fix any CI-specific issues (timing, env differences)
- [ ] Verify green CI badge

**Success Criteria**:
- [x] All Go tests passing (already done ✅)
- [ ] All Playwright tests passing (currently 10 failing → target 0)
- [ ] No flaky tests (run 3 times, all pass)
- [ ] CI green (all checks passing)
- [ ] Tests document what they're testing
- [ ] No unnecessary skipped tests (currently 14 → review needed)

**Time Estimate**: 6-8 hours total (updated after detailed analysis)

**Files to Modify**:
```
DELETE:
- e2e/websocket-connection.spec.js (obsolete static HTML tests)

MODIFY (add data-testid):
- web/src/components/dashboard/ProviderCards.tsx
- web/src/components/dashboard/ActivityFeed.tsx (already has testid)
- web/src/components/dashboard/StatsCards.tsx (already has testid)

MODIFY (fix test selectors):
- e2e/provider-health-ui.spec.ts (3 tests)
- e2e/dashboard-react.spec.ts (1 test)  
- e2e/activity-feed-enhanced.spec.ts (1 test)

VERIFY (should pass now):
- e2e/websocket-critical.spec.ts (2 tests)
```

**Quick Win Check** (10 minutes):
Run `npx playwright test e2e/websocket-critical.spec.ts` to verify tests 9-10 now pass with data-testid changes. If they pass, we're down to 8 failing tests immediately.

#### 0.7. Provider Health UI Redesign - Full Width & Details (HIGH 🟡) - 🔄 **IN PROGRESS** - 3-4 hours

**Problem**: Provider cards don't match Activity feed styling
- Cards use grid layout (doesn't fill width)
- Missing webhook activity indicators
- No refresh button
- Can't see webhook payloads
- Inconsistent with Activity feed

**Design Goals**:
1. **Full-width cards** - Match Activity feed width exactly
2. **Webhook indicators** - Flash/animation when webhook received
3. **Refresh button** - Manual refresh in header
4. **Payload viewing** - Click to see full webhook JSON
5. **Message preview** - Show last webhook event type

**Implementation**:

**Phase 1: Full-width Layout** (1 hour)
**Phase 2: Webhook Animation** (1 hour)
**Phase 3: Backend Support** (1 hour)
**Phase 4: WebSocket Updates** (30 min)

**Success Criteria**:
- [x] Provider cards full-width
- [x] Refresh button in header works
- [ ] Webhook flash animation on receipt
- [x] Can view last webhook payload
- [x] Message count accurate
- [x] Matches Activity feed styling

**Time**: 3-4 hours
**Progress**: 5/6 complete (Commit: f12a2f0)

#### 4. Add Workflow Source Links (HIGH 🟡) - ✅ **COMPLETE** (Commit: fa33c2e) - 1 hour
**Problem**: No way to open workflow run at source (GitHub/Azure DevOps)
- Users can't navigate to original CI system
- Missing Run.URL field usage in UI
- Hard to investigate failures without source link

**Requirements**:
1. **Each workflow card** - Add external link icon → Run.URL
2. **Activity feed items** - Clickable link to source
3. **Run details modal** - Prominent "View on GitHub" button
4. **Provider-aware icons** - GitHub logo for GitHub, Azure logo for Azure

**Steps**:
1. [ ] Add external link icon to RunRow component
2. [ ] Add onClick handler to open Run.URL in new tab
3. [ ] Update ActivityFeed to show clickable links
4. [ ] Test with GitHub runs (verify URL format)
5. [ ] Add Azure DevOps URL format (future-proof)
6. [ ] Write test verifying links are clickable
7. [ ] Commit + push

**Success Criteria**:
- [ ] Can click to open run on GitHub
- [ ] External link icon visible on every run
- [ ] Opens in new tab
- [ ] Works for both completed and in-progress runs
- [ ] Ready for Azure DevOps URLs

#### 6. Fix UI Flicker (LOW 🟢) - ✅ **COMPLETED** (Commit: 9205aa8) - 1-2 hours
**Problem**: Possible flicker during WebSocket updates
- Component re-renders on every message
- No memoization
- Animations re-trigger unnecessarily

**Steps**:
1. [ ] Write test that captures flicker (video recording)
2. [ ] Identify which components re-render
3. [ ] Add React.memo() where needed
4. [ ] Use layout animations instead of mount animations
5. [ ] Verify smooth updates
6. [ ] Commit + push

**Success Criteria**:
- [ ] No visible flicker on updates
- [ ] 60fps smooth animations
- [ ] Test proves improvement

#### 9. Fix Activity Feed Duration Display (HIGH 🟡) - ✅ **COMPLETE** (Commit: 2cd74e7) - 30 minutes
**Problem**: Duration in Activity Feed always shows "0s" - not calculating correctly

**Root Cause**: Need to investigate:
- Is duration field populated in backend?
- Is it being calculated from StartedAt/UpdatedAt?
- Is formatting correct?

**Steps**:
1. [ ] Check backend Run struct - verify duration field
2. [ ] Check if duration is calculated in provider
3. [ ] Check ActivityFeed.tsx formatting
4. [ ] Fix calculation/display
5. [ ] Test with real runs
6. [ ] Commit + push

**Success Criteria**:
- [ ] Duration shows actual run time (e.g., "2m 34s", "45s")
- [ ] Works for in-progress runs (shows elapsed time)
- [ ] Works for completed runs (shows total duration)
- [ ] Format is human-readable

#### 10. Enhanced Debug Panel - Unified Log Stream (HIGH 🟡) - ✅ **COMPLETE** (Commit: 4395dce) - 3 hours
**Problem**: No way to see backend logs in UI, have to switch between terminal and browser

**Vision**: One place to see EVERYTHING happening in the system
- Backend Go logs (color-coded by level) ✅
- Frontend console.log() messages (Future enhancement)
- WebSocket frames (sent/received) (Already in WebSocket tab)
- Webhook deliveries (incoming HTTP requests) (Future enhancement)
- Provider polling cycles (Future enhancement)
- Store updates (Future enhancement)
- All timestamped, searchable, filterable

**Implementation**:
1. [x] Backend: Add `/debug/logs` WebSocket endpoint
2. [x] Backend: Stream log lines to connected clients
3. [x] Backend: Include log level, component, message
4. [x] Frontend: Add "Logs" tab to Debug Panel
5. [x] Frontend: Display backend logs in real-time
6. [ ] Add filters: level (debug/info/warn/error), component, search (Future)
7. [x] Add clear button
8. [ ] Add export to file (Future)
9. [x] Write tests
10. [x] Commit + push

**Success Criteria**:
- [x] Can see backend logs in browser
- [x] Color-coded by log level
- [x] Shows timestamp, level, component, message
- [x] Real-time WebSocket updates
- [x] Connection status indicator
- [x] Clear button
- [x] All tests passing

#### 11. Debug Panel - Event Timeline Visualization (MEDIUM 🟡) - ✅ **COMPLETED** (Commit: 6fc906d) - 4-5 hours
**Problem**: Hard to understand event flow and timing relationships

**Vision**: Visual timeline showing all system events with timing and relationships

**Features**:
- Horizontal timeline (last 5 minutes visible)
- Lanes for: Polling, Webhooks, WebSocket, UI Updates, Errors
- Click event to see details
- Zoom in/out
- Highlight correlated events (poll → store update → UI update)
- Show gaps (missed polls, slow responses)

**Success Criteria**:
- [ ] Shows all major system events
- [ ] Visual representation of timing
- [ ] Can identify bottlenecks
- [ ] Can correlate related events
- [ ] Helps debug timing issues

#### 12. Debug Panel - State Inspector (MEDIUM 🟡) - 2-3 hours
**Problem**: Can't easily see current Store state or DashboardContext

**Features**:
- View all runs in Store (raw JSON)
- View DashboardContext state
- View provider health
- Compare: WebSocket message vs current state
- Export state to JSON file
- Refresh button

**Success Criteria**:
- [ ] Can view Store state
- [ ] Can view Context state
- [ ] Can export to JSON
- [ ] Can compare message vs state
- [ ] Helps debug state issues

#### 13. Debug Panel - Performance Profiler (LOW 🟢) - 3-4 hours
**Problem**: Need to identify performance bottlenecks and memory leaks

**Features**:
- Component render time tracking
- WebSocket message processing time
- Store update latency
- Memory usage over time
- FPS tracker
- Network request timing

**Success Criteria**:
- [ ] Tracks key performance metrics
- [ ] Identifies slow components
- [ ] Detects memory leaks
- [ ] Helps optimize performance

#### 14. Debug Panel - Webhook Simulator (LOW 🟢) - 2-3 hours
**Problem**: Hard to test webhook handling without triggering real CI runs

**Features**:
- Send fake webhook payloads
- Pre-set templates (success, failure, in_progress)
- Custom JSON editor
- Replay captured webhooks
- Test error scenarios (malformed, invalid signature)

**Success Criteria**:
- [ ] Can send test webhooks
- [ ] Can use templates
- [ ] Can edit JSON
- [ ] Can test error cases
- [ ] Helps test webhook handling

#### 15. Build Failure Notification Use Case (LOW 🟢) - 1-2 hours
**Problem**: Need to design for external tool sending build failure signals

**Requirements**:
- External tool will send signal when build fails
- UI should show immediate visual feedback
- Need robust testing strategy

**Success Criteria**:
- [ ] External tool can POST build failure
- [ ] UI shows immediate feedback
- [ ] Tests verify end-to-end flow
- [ ] API documented

---

## ✅ Completed Tasks

- ✅ **0.0. Investigate and Fix CI Failures (Comprehensive Tests)** (Commit: 83c5864)
- ✅ **0. Webhook vs Polling Validation Test** (Commit: edfc9fb)
- ✅ **0.5. Webhook Integration Testing - Trigger & Verify** (Commit: 940a953)
- ✅ **0.6. Automatic ngrok Setup & Webhook Documentation** (Commit: aa98180)
- ✅ **1. WebSocket Connection Fix** (Commit: 2bde007)
- ✅ **2. Remove Orphaned Code** (Commit: 9ad241b)
- ✅ **3. Startup Performance - Add Timing Metrics** (Commit: ab83713)
- ✅ **4. Add Workflow Source Links** (Commit: fa33c2e)
- ✅ **5. Enhanced Activity Feed with Message Details** (Commit: 932bb48)
- ✅ **7. GitHub Logo Investigation** (Commit: 07adf02)
- ✅ **8. Developer Debugging Dashboard** (Commit: 7f5986f)
- ✅ **9. Fix Activity Feed Duration Display** (Commit: 2cd74e7)
- ✅ **10. Enhanced Debug Panel - Unified Log Stream** (Commit: 4395dce)
- ✅ **Phase 0.6: UI Polish & Bug Fixes** (All sub-tasks completed)
- ✅ **Phase -1: Testing & Screenshots** (All sub-tasks completed)
- ✅ **Phase 0: React Migration** (All sub-tasks completed)

---

## Build & Run

```bash
# Development
make web-dev              # Start Vite dev server (localhost:5173)

# Production
make build                # Build web + Go binary
./bin/ceye --port 8080    # Run dashboard

# Testing
npx playwright test       # Run all tests
npx playwright test --ui  # Interactive test UI
```

---

## Future Options (Post-React Migration)

### Option 1: Enhanced Features
- Historical data storage (SQLite)
- Trends & analytics
- Alerting (Slack/Email)
- Performance metrics

### Option 2: Azure DevOps Provider
- Complete provider implementation
- Feature parity with GitHub
- Multi-provider testing

### Option 3: User Experience
- Keyboard shortcuts
- Theme system
- Dashboard customization
- Advanced filtering

### Option 4: Enterprise
- Authentication & RBAC
- Audit logs
- Multi-tenancy
- SSO integration

---

**For detailed task breakdowns and historical context, see git history and commit messages.**