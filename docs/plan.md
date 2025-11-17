# ceye Development Plan

**Last Updated**: 2025-11-17 19:01 UTC  
**Status**: Phase 0.7 Critical Issues - 🟢 **11 of 20 COMPLETE**

## Current Status

**React Migration**: Complete ✅  
**Test Suite**: 27/29 passing ✅  
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
- 🟡 GitHub logo may not be correct (needs clarification)
- 🟡 Build failure notification use case needs design
- 🟡 Fix Activity Feed Duration Display (currently shows "0s")
- 🟡 Enhanced Debug Panel - Unified Log Stream
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
- [ ] Provider cards full-width
- [ ] Refresh button in header works
- [ ] Webhook flash animation on receipt
- [ ] Can view last webhook payload
- [ ] Message count accurate
- [ ] Matches Activity feed styling

**Time**: 3-4 hours

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

#### 5. Enhanced Activity Feed with Message Details (HIGH 🟡) - 🔄 **IN PROGRESS** - 2-3 hours
**Problem**: Activity feed shows minimal info - just workflow name and status
- No message content details
- Can't see what changed
- No drill-down capability
- Missing context for debugging

**Message Details to Show**:
1. **Event type** - `workflow_run.completed`, `check_suite.requested`, etc.
2. **Duration** - How long the run took
3. **Commit info** - SHA + first line of commit message
4. **Changed files** - Count of files changed (if available)
5. **Actor** - Who triggered the run
6. **Conclusion** - success/failure/cancelled
7. **Message body** - First 100 chars of relevant message

**Expandable Details**:
- Click to expand full message JSON
- Show complete webhook payload
- Link to GitHub/Azure/GitLab run URL (see Task 4)

**Steps**:
1. [ ] Extend ActivityItem type with rich data
2. [ ] Update DashboardContext to pass full run details
3. [ ] Redesign ActivityFeed layout for more info
4. [ ] Add expand/collapse for full details
5. [ ] Add external link button
6. [ ] Style improvements (better icons, colors, spacing)
7. [ ] Write tests for new layout
8. [ ] Commit + push

**Success Criteria**:
- [ ] Activity shows meaningful details
- [ ] Can expand for full JSON
- [ ] External links work
- [ ] Easy to scan and understand
- [ ] Helps with debugging

#### 6. Fix UI Flicker (LOW 🟢) - 🔄 **IN PROGRESS** - 1-2 hours
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

#### 10. Enhanced Debug Panel - Unified Log Stream (HIGH 🟡) - 3-4 hours
**Problem**: No way to see backend logs in UI, have to switch between terminal and browser

**Vision**: One place to see EVERYTHING happening in the system
- Backend Go logs (color-coded by level)
- Frontend console.log() messages
- WebSocket frames (sent/received)
- Webhook deliveries (incoming HTTP requests)
- Provider polling cycles
- Store updates
- All timestamped, searchable, filterable

**Implementation**:
1. [ ] Backend: Add `/debug/logs` WebSocket endpoint
2. [ ] Backend: Stream log lines to connected clients
3. [ ] Backend: Include log level, component, message
4. [ ] Frontend: Add "Logs" tab to Debug Panel
5. [ ] Frontend: Merge backend + console logs in single stream
6. [ ] Add filters: level (debug/info/warn/error), component, search
7. [ ] Add auto-scroll toggle + clear button
8. [ ] Add export to file
9. [ ] Write tests
10. [ ] Commit + push

**Success Criteria**:
- [ ] Can see backend logs in browser
- [ ] Can see frontend console in same stream
- [ ] Can see WebSocket traffic
- [ ] Can see webhook deliveries
- [ ] Can filter by level/component
- [ ] Can search full history
- [ ] Auto-scrolls to new messages
- [ ] Can export to file for sharing

#### 11. Debug Panel - Event Timeline Visualization (MEDIUM 🟡) - 4-5 hours
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
- ✅ **5. Enhanced Activity Feed with Message Details** (Commit: 932bb48)
- ✅ **7. GitHub Logo Investigation** (Commit: 07adf02)
- ✅ **4. Add Workflow Source Links** (Commit: fa33c2e)
- ✅ **8. Developer Debugging Dashboard** (Commit: 7f5986f)
- ✅ **9. Fix Activity Feed Duration Display** (Commit: 2cd74e7)
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