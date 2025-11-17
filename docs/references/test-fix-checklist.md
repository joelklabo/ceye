# Test Fix Checklist - Task 0.0.1

**Created**: 2025-11-17 21:19 UTC
**Status**: 10 failing tests → Target: 0 failures
**Estimated Time**: 6-8 hours

## Quick Start

```bash
# Run specific failing tests
npx playwright test e2e/websocket-critical.spec.ts  # Should pass now!
npx playwright test e2e/provider-health-ui.spec.ts
npx playwright test e2e/dashboard-react.spec.ts:70
npx playwright test e2e/activity-feed-enhanced.spec.ts:36

# Run all tests
npx playwright test --reporter=list
```

## Phase-by-Phase Checklist

### Phase 1: ✅ DONE (Commit: 0e1fdd3)
- [x] Added data-testid="activity-feed" to ActivityFeed.tsx
- [x] Added data-testid attributes to StatsCards.tsx
  - stat-total-runs (hidden element)
  - stat-active-runs, stat-queued-runs, stat-success-runs, stat-failed-runs

### Phase 2: Delete Obsolete Tests (30 min)
- [ ] Delete `e2e/websocket-connection.spec.js`
- [ ] Verify coverage in websocket-critical.spec.ts
- [ ] Run: `npx playwright test e2e/websocket-critical.spec.ts`
- [ ] Commit: "test: Remove obsolete static HTML WebSocket tests"
- [ ] **Fixes**: Tests 1, 2, 3 (3 tests fixed)

### Phase 3: Fix Provider Health Selectors (1.5 hrs)
- [ ] **Test 4**: Add data-testid="provider-cards" to ProviderCards.tsx
  - Location: `web/src/components/dashboard/ProviderCards.tsx`
  - Update test line 16: `page.locator('[data-testid="provider-cards"]')`
  
- [ ] **Test 5**: Fix provider name selector
  - Update test line 35: Change `text=github` to `text=demo`
  - OR: Use more generic selector
  
- [ ] **Test 6**: Add container testids
  - ProviderCards: Add data-testid="provider-health-section"
  - ActivityFeed: Add data-testid="activity-section"
  - Update test to compare these directly
  
- [ ] Run: `npx playwright test e2e/provider-health-ui.spec.ts`
- [ ] Commit: "test: Fix Provider Health UI test selectors"
- [ ] **Fixes**: Tests 4, 5, 6 (3 tests fixed)

### Phase 4: Fix Refresh Button Test (45 min)
- [ ] **Test 7**: dashboard-react.spec.ts:70
  - Option A: Add data-testid="provider-refresh-button" to ProviderCards.tsx
  - Option B: Simplify selector to match RefreshCw component
  - Update test line 80 with new selector
  
- [ ] Run: `npx playwright test e2e/dashboard-react.spec.ts:70`
- [ ] Commit: "test: Fix provider health refresh button test"
- [ ] **Fixes**: Test 7 (1 test fixed)

### Phase 5: Fix Duration Test (45 min)
- [ ] **Test 8**: activity-feed-enhanced.spec.ts:36
  - Problem: We hide "Duration:" when 0 seconds
  - Solution: Update test to handle optional duration
  - Add: `if (await page.locator('text=/Duration:/').count() > 0)`
  
- [ ] Run: `npx playwright test e2e/activity-feed-enhanced.spec.ts:36`
- [ ] Commit: "test: Fix activity feed duration test for 0s case"
- [ ] **Fixes**: Test 8 (1 test fixed)

### Phase 6: Verification (2 hrs)
- [ ] Run all tests: `npx playwright test --reporter=list`
- [ ] Count failures (should be 0 or close)
- [ ] Run 3 times to check flakiness
- [ ] Fix any flaky tests found
- [ ] Update test counts in plan.md
- [ ] Commit: "test: All Playwright tests passing - comprehensive fix"

### Phase 7: CI Validation (1 hr)
- [ ] Push to origin/main
- [ ] Monitor GitHub Actions CI
- [ ] Check test results in CI
- [ ] Fix any CI-specific failures
- [ ] Verify green checkmark

## Expected Progress

| Phase | Tests Fixed | Tests Remaining |
|-------|-------------|-----------------|
| Start | 0 | 10 |
| Phase 1 | 2 (maybe) | 8-10 |
| Phase 2 | 3 | 5-7 |
| Phase 3 | 3 | 2-4 |
| Phase 4 | 1 | 1-3 |
| Phase 5 | 1 | 0-2 |
| Phase 6 | All | 0 |

## Test Details

### Group A: Obsolete Tests (DELETE)
1. websocket-connection.spec.js:4 - Looking for `#connectionStatus`, `#lastUpdate`
2. websocket-connection.spec.js:28 - Looking for `.connection-indicator.connected`
3. websocket-connection.spec.js:55 - Looking for `#activityToggle`

### Group B: Provider Health (FIX SELECTORS)
4. provider-health-ui.spec.ts:10 - `[class*="space-y"]` too generic
5. provider-health-ui.spec.ts:33 - `text=github` wrong provider name
6. provider-health-ui.spec.ts:107 - Fragile parent selector

### Group C: Dashboard (FIX SVG)
7. dashboard-react.spec.ts:70 - `svg[data-lucide="refresh-cw"]` attribute missing

### Group D: Activity Feed (FIX ASSERTION)
8. activity-feed-enhanced.spec.ts:36 - Duration hidden when 0s

### Group E: WebSocket Critical (SHOULD BE FIXED)
9. websocket-critical.spec.ts:60 - Added data-testid="activity-feed" ✅
10. websocket-critical.spec.ts:79 - Added data-testid="stat-*" ✅

## Files to Touch

**Delete:**
- e2e/websocket-connection.spec.js

**Modify (React components):**
- web/src/components/dashboard/ProviderCards.tsx
- web/src/components/dashboard/ActivityFeed.tsx

**Modify (Tests):**
- e2e/provider-health-ui.spec.ts
- e2e/dashboard-react.spec.ts
- e2e/activity-feed-enhanced.spec.ts

## Success Metrics

- [ ] 0 failing Playwright tests
- [ ] All Go tests still passing
- [ ] No flaky tests (3 consecutive runs pass)
- [ ] CI green
- [ ] Updated plan.md with final status
