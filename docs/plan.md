# ceye Development Plan

**Last Updated**: 2025-11-17 10:19 UTC  
**Status**: React Migration - Building Stunning Dashboard 🚀

## Current Status

**React Migration Approved - Priority #1**

Based on research and user approval, we're migrating to a modern React stack to achieve a stunning, production-quality dashboard that rivals Vercel, Linear, and Railway.

**New Direction**:
- 🚀 **React + Vite + Shadcn/ui** - Modern, beautiful components
- 🎨 **Professional Design** - Dark-first, animated, polished
- 📦 **TypeScript** - Type-safe, excellent DX
- ⚡ **Framer Motion** - Buttery smooth animations
- ♿ **Accessibility-first** - ARIA compliant, keyboard nav

**Total Work**: ~22 hours across 4 phases

---

## 🚧 Active Development Tasks

### Phase -1: Comprehensive Testing & Screenshots (6 hours) - **PRIORITY -1** 🔥🔥🔥🔥

**Goal**: Implement robust browser automation testing with Playwright and generate marketing-grade screenshots for README

**Why Playwright?**
Based on 2024 research, Playwright is superior for this project:
- ✅ **Fastest execution** (beats Cypress & Puppeteer in benchmarks)
- ✅ **Cross-browser** (Chromium, Firefox, WebKit)
- ✅ **TypeScript-first** (matches our stack)
- ✅ **Built-in screenshot** (fullPage, element, viewport options)
- ✅ **Auto-waiting** (less flaky tests)
- ✅ **Multi-tab support** (Cypress can't do this)
- ✅ **Network mocking** (for testing without backend)
- ✅ **Parallel testing** (fast CI/CD)

**Already Installed**: Playwright is already in package.json from previous E2E tests

**Test Strategy**:
1. **Integration Tests** - Full user journeys with real WebSocket
2. **Visual Regression** - Screenshot comparison between releases
3. **Marketing Screenshots** - Automated, consistent, high-quality
4. **Cross-browser** - Test on Chrome, Firefox, Safari (WebKit)

---

#### Phase -1.1: Integration Test Suite (3 hours) - ✅ **COMPLETE**

**Goal**: Comprehensive Playwright tests covering all dashboard functionality

**Commit**: 221c45b

**Tests Implemented**:

1. **Dashboard Loading & Connection** (30min)
   - [✅] Page loads successfully
   - [✅] WebSocket connects within 5s
   - [✅] Connection indicator shows green
   - [✅] Initial data loads (runs, stats, providers)
   - [✅] Error handling if WebSocket fails

2. **Stats Cards** (30min)
   - [✅] All 4 cards render (Running, Queued, Success, Failed)
   - [✅] Counters display numbers from WebSocket
   - [✅] Icons display correctly
   - [✅] Responsive layout (mobile, tablet, desktop)

3. **Runs Table** (1 hour)
   - [✅] Table renders with runs or empty state
   - [✅] Search box is present
   - [✅] Table headers are correct
   - [✅] Search/filter works correctly
   - [✅] Sorting works

4. **Provider Health Cards** (30min)
   - [✅] Section exists
   - [✅] Cards render for each provider
   - [✅] Health indicator visible

5. **Activity Feed** (30min)
   - [✅] Feed renders
   - [✅] Expand/collapse works
   - [✅] Items or empty state displays

6. **Real-Time Updates** (30min)
   - [✅] Dashboard receives WebSocket messages
   - [✅] Connection indicator updates on disconnect

**Test Structure**:
```
e2e/
├── dashboard.spec.ts          # Main dashboard tests
├── websocket.spec.ts          # WebSocket connectivity
├── components/
│   ├── stats-cards.spec.ts
│   ├── runs-table.spec.ts
│   ├── provider-cards.spec.ts
│   └── activity-feed.spec.ts
├── visual/
│   └── screenshot-regression.spec.ts
└── fixtures/
    └── mock-websocket-data.ts  # Test data
```

**Test Results**: 22/22 passing (100% ✅)

**Success Criteria**: ✅ ALL MET
- ✅ All tests pass (Chrome tested, cross-browser ready)
- ✅ Tests run in ~8 seconds
- ✅ Comprehensive component coverage
- ✅ No flaky tests (robust selectors + waits)
- ✅ CI/CD integration ready

**Key Features**:
- TypeScript for type safety
- WebSocket logging injection
- Helper functions for waits
- Proper error handling
- Screenshot/video/trace on failure

---

#### Phase -1.2: Marketing Screenshots (3 hours) - **IN PROGRESS** 🚧

**Goal**: Automated, professional screenshots for README and documentation

**Screenshots Needed**:

1. **Hero Screenshot** (Full dashboard view)
   - Desktop (1920x1080) - Dark mode
   - Shows all components with live data
   - 25+ runs in table
   - 3+ providers healthy
   - Activity feed populated
   - Connection indicator green

2. **Component Highlights** (Individual features)
   - Stats cards with animated counters
   - Runs table with sorting/filtering
   - Provider health cards
   - Activity feed expanded
   - Mobile responsive view

3. **Feature Demonstrations**
   - Real-time update sequence (before/after)
   - WebSocket reconnection
   - Error states (provider down)
   - Empty states

4. **Cross-Platform**
   - Chrome desktop
   - Firefox desktop  
   - Safari desktop
   - Mobile (iPhone 14, Pixel 7)

**Implementation**:
```typescript
// e2e/screenshots/generate-marketing.spec.ts
test('Generate marketing screenshots', async ({ page }) => {
  await page.goto('http://localhost:8080');
  await page.waitForSelector('.stats-cards');
  
  // Full dashboard hero shot
  await page.screenshot({
    path: 'docs/screenshots/dashboard-hero.png',
    fullPage: true,
  });
  
  // Highlight specific component
  await page.evaluate(() => {
    const element = document.querySelector('.stats-cards');
    element.style.border = '2px solid #22c55e';
  });
  await page.screenshot({
    path: 'docs/screenshots/stats-cards-highlight.png',
    clip: { x: 0, y: 100, width: 1200, height: 400 },
  });
});
```

**Automated Workflow**:
1. Start ceye server with demo data
2. Run screenshot generation script
3. Compress images (80% quality JPEG)
4. Auto-update README with new screenshots
5. Git commit with descriptive message

**Directory Structure**:
```
docs/
└── screenshots/
    ├── hero/
    │   ├── dashboard-dark.png
    │   └── dashboard-mobile.png
    ├── components/
    │   ├── stats-cards.png
    │   ├── runs-table.png
    │   ├── provider-cards.png
    │   └── activity-feed.png
    ├── features/
    │   ├── real-time-update.gif
    │   ├── sorting-filtering.png
    │   └── websocket-reconnect.png
    └── cross-platform/
        ├── chrome.png
        ├── firefox.png
        ├── safari.png
        └── mobile.png
```

**Success Criteria**:
- ✅ All screenshots are 4K-ready
- ✅ Consistent styling (dark mode)
- ✅ No placeholder/loading states visible
- ✅ Marketing-quality composition
- ✅ README auto-updated with images
- ✅ CI generates screenshots on release

---

**Commit Strategy**:
- Commit 1: Integration test suite
- Commit 2: Screenshot generation script
- Commit 3: README with screenshots

**Next**: After Phase -1 complete, continue with Phase 0.4 (Polish & Excellence)

---

### Phase 0: React Migration (22 hours) - **PRIORITY 0** 🔥🔥🔥

**Goal**: Migrate to React + Vite + Shadcn/ui for a stunning, production-quality dashboard

**Why React?**
- ✅ Component-based architecture (vs manual DOM manipulation)
- ✅ Modern animations (Framer Motion)
- ✅ Type-safe (TypeScript)
- ✅ Excellent DX (hot reload, fast builds)
- ✅ Production-ready UI components (Shadcn/ui)
- ✅ Industry standard stack (easy hiring, huge ecosystem)
- ✅ Still embeddable in Go backend

**Tech Stack**:
- **Vite** - Lightning fast build tool, instant HMR
- **React 19** - Modern UI library
- **TypeScript** - Type safety
- **Shadcn/ui** - Accessible components on Radix UI
- **Tailwind CSS v3** - Utility-first styling
- **Framer Motion** - Production-grade animations
- **ApexCharts** - Real-time chart animations
- **React Query (TanStack Query)** - Server state management
- **Lucide Icons** - Modern, beautiful icons

**Design Inspiration**:
- **Vercel** - Clean, professional, modern
- **Linear** - Minimal, functional, beautiful typography
- **Railway** - Bold colors, clear status indicators

**Color Palette**:
```
Dark Mode (Primary):
- Background: #0a0a0a (deep black)
- Surface: #18181b (elevated cards)
- Border: #27272a (subtle dividers)
- Text: #fafafa (high contrast)
- Muted: #71717a (secondary text)

Status Colors:
- Success: #22c55e (green-500)
- Error: #ef4444 (red-500)
- Warning: #f59e0b (amber-500)
- Info: #3b82f6 (blue-500)
- Running: #8b5cf6 (purple-500)
```

**Project Structure**:
```
ceye/
├── web/                          # NEW: React app
│   ├── src/
│   │   ├── components/
│   │   │   ├── ui/              # Shadcn components
│   │   │   ├── dashboard/
│   │   │   ├── runs/
│   │   │   └── providers/
│   │   ├── hooks/
│   │   ├── lib/
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── public/
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.ts
└── internal/server/
    └── server.go                 # Serves embedded React build
```

---

#### Phase 0.1: Foundation Setup (4 hours) - ✅ **COMPLETE**

**Goal**: Initialize React project and verify Go embedding works

**Commit**: 2d76988

**Tasks**:
- [✅] Initialize Vite + React + TypeScript project in `/web`
  - `npm create vite@latest web -- --template react-ts`
- [✅] Install dependencies:
  - `npm install -D tailwindcss postcss autoprefixer`
  - `npm install @radix-ui/react-* (Shadcn dependencies)`
  - `npm install framer-motion apexcharts`
  - `npm install @tanstack/react-query`
  - `npm install lucide-react`
- [✅] Initialize Shadcn/ui:
  - Manual setup (cli had path issues)
  - Configured Tailwind v3 with design tokens
  - Created utils.ts with cn() helper
- [✅] Configure Vite build:
  - Output to `dist/` folder
  - Set base path for embedded serving
  - Added path alias support (@/)
- [✅] Update Go server:
  - Created cmd/ceye/web.go with embed directive
  - Modified Server.New() to accept fs.FS parameter
  - Build copies web/dist to cmd/ceye/web/dist for embedding
- [✅] Create basic App shell:
  - Simple header + main layout
  - Test component rendering with Tailwind
- [✅] Test build & embed:
  - `npm run build` (creates dist/)
  - `go build` (embeds dist/)
  - Verified Go serves React app at http://localhost:8080
- [✅] Add npm scripts to Makefile:
  - `make web-dev` (starts Vite dev server)
  - `make web-build` (builds for production)
  - `make web-install` (installs dependencies)
  - `make build` (builds web + Go automatically)

**Success Criteria**: ✅ ALL MET
- ✅ Vite dev server runs (localhost:5173)
- ✅ React app renders with Tailwind styles
- ✅ Go embeds and serves React build
- ✅ No build errors
- ✅ Hot reload works in dev mode

---

#### Phase 0.2: Core Components (8 hours) - ✅ **COMPLETE**

**Goal**: Build main dashboard components with animations

**Commit**: e873ba3

**Tasks**:
- [✅] **Stats Cards** (2 hours)
  - 4 cards: Running, Queued, Success, Failed
  - Animated counters (count up/down)
  - Pulse effect on value changes
  - Gradient backgrounds
  - Lucide icons
  
- [✅] **Runs Table** (3 hours)
  - Full table implementation
  - Columns: Provider, Repo, Workflow, Status, Duration, Time
  - Sorting (click headers with arrows)
  - Filtering (search box with icon)
  - Status badges with colors
  - Row hover effects
  - Smooth entrance animations with stagger
  - Empty state handling
  
- [✅] **Provider Health Cards** (2 hours)
  - Card per provider
  - Health indicator with pulse animation
  - Last update timestamp
  - Error states with count display
  - Hover effect with scale
  - Status colors with smooth transitions
  
- [✅] **Activity Feed** (1 hour)
  - Collapsible timeline design
  - Slide-in animation for entries
  - Lucide icons for event types
  - Timestamp formatting
  - Expand/collapse behavior
  - Empty state handling

**Success Criteria**: ✅ ALL MET
- ✅ All components render correctly
- ✅ Animations smooth at 60fps (Framer Motion)
- ✅ Responsive layout (grid-based, mobile-friendly)
- ✅ Accessible (semantic HTML, hover states)

---

#### Phase 0.3: Real-Time Integration (4 hours) - ✅ **COMPLETE**

**Goal**: Connect React to Go WebSocket for live updates

**Commit**: 5e2f2e0

**Tasks**:
- [✅] **WebSocket Hook** (2 hours)
  - Custom `useWebSocket` hook
  - Auto-reconnect on disconnect (10 attempts, 3s interval)
  - Error handling and logging
  - Connection status tracking
  - Message parsing
  
- [✅] **Dashboard Context** (1 hour)
  - Created DashboardContext with provider
  - useDashboard() hook for components
  - Connects to WebSocket on mount
  - Calculates derived state (stats from runs)
  - Exposes connection status
  
- [✅] **Real-Time Updates** (1 hour)
  - All components use real WebSocket data
  - Connection indicator in header (green/red dot)
  - Last update timestamp displayed
  - Smooth state transitions with Framer Motion
  - Activity feed generated from recent runs

**Success Criteria**: ✅ ALL MET
- ✅ WebSocket connects on page load
- ✅ Real-time updates appear instantly
- ✅ Reconnection works seamlessly
- ✅ No memory leaks (proper cleanup)
- ✅ Smooth, no jarring updates

---

#### Phase 0.4: Polish & Excellence (6 hours)

**Goal**: Make it stunning and production-ready

**Tasks**:
- [ ] **Animations & Micro-interactions** (2 hours)
  - Framer Motion on all transitions
  - Hover effects on interactive elements
  - Loading states (skeleton screens)
  - Success/error animations
  - Page transition effects
  - Ripple effect on clicks
  
- [ ] **Responsive Design** (1 hour)
  - Mobile layout (stack cards)
  - Tablet layout (2-column grid)
  - Desktop layout (3-column grid)
  - Test on 320px, 768px, 1024px, 1920px
  
- [ ] **Accessibility** (1 hour)
  - Keyboard navigation (Tab, Enter, Esc)
  - Screen reader labels
  - Focus indicators
  - Color contrast audit (WCAG AA)
  - Skip links
  
- [ ] **Performance Optimization** (1 hour)
  - Code splitting (React.lazy)
  - Memoization (React.memo, useMemo)
  - Debounce WebSocket updates
  - Virtual scrolling for large lists
  - Bundle size analysis
  
- [ ] **Error Boundaries** (0.5 hour)
  - Catch component errors
  - Graceful fallback UI
  - Error reporting
  
- [ ] **Dark Mode** (0.5 hour)
  - Respect system preference
  - Smooth theme transitions
  - Persisted user choice

**Success Criteria**:
- ✅ Animations are buttery smooth (60fps)
- ✅ Responsive on all screen sizes
- ✅ Accessible (keyboard + screen reader)
- ✅ Bundle size < 400KB gzipped
- ✅ Lighthouse score > 90
- ✅ No console errors/warnings
- ✅ Looks stunning! 🤩

**Commit**: After Phase 0.4 complete

---

## Previous Completed Phases

### Phase 1: Remove TUI (2 hours) - **PRIORITY 1**

**Goal**: Remove terminal UI completely and focus on web-only

**Commits**: 
- fd035ee - Package deletion and code cleanup
- 142f9e9 - Documentation updates

**Tasks**:
- [✅] Delete `internal/ui/` package entirely
- [✅] Remove Bubble Tea dependencies (`go.mod`, `go.sum`)
  - `github.com/charmbracelet/bubbletea`
  - `github.com/charmbracelet/bubbles`
  - `github.com/charmbracelet/lipgloss`
- [✅] Update `cmd/ceye/main.go`:
  - Remove TUI mode flags
  - Remove TUI initialization code
  - Make `--web` the default mode
  - Keep webhook and demo flags
- [✅] Remove TUI tests from test suite (no separate TUI tests existed)
- [✅] Update `docs/agents.md` (remove all TUI sections)
- [✅] Update `docs/readme.md` (remove TUI examples)
- [✅] Run `go mod tidy` to clean dependencies
- [✅] Verify: `make build && make test` passes

**Success Criteria**: ✅ COMPLETE
- ✅ No TUI code remains
- ✅ Binary size: 20MB
- ✅ All tests pass
- ✅ Documentation updated

---

### Phase 1.5: Remove "Web UI" Terminology (15 minutes) - **PRIORITY 1**

**Goal**: Remove "Web UI" references - it's just the UI now

Since there's no longer a terminal UI, we shouldn't call it "Web UI" anymore. It's just the interface.

**Commit**: f6b718c

**Tasks**:
- [✅] Update `docs/agents.md`: "Web UI" → "UI" or "dashboard"
- [✅] Update `docs/readme.md`: "Web UI" → "dashboard" or remove qualifier
- [✅] Update `cmd/ceye/main.go`: Command description "Web UI" → "Dashboard"
- [✅] Search codebase for remaining "Web UI" / "web ui" references
- [✅] Update any remaining user-facing strings

**Success Criteria**: ✅ COMPLETE
- ✅ No "Web UI" terminology remains
- Just "dashboard" or "UI" or "interface"
- Clearer, simpler language

---

### Phase 2: Fix Critical Bugs (1 hour) - **PRIORITY 1**

**Goal**: Fix bugs found by E2E tests

**Root Cause Analysis**:
The E2E test failures were caused by TWO bugs:

1. **Activity log DOM manipulation race condition** (`app.js`)
   - Symptom: JavaScript error `Cannot read properties of undefined (reading 'contains')`
   - Cause: Checking/removing "waiting" message AFTER appending new items
   - This caused DOM to be in inconsistent state during checks
   - Side effect: Error prevented WebSocket onmessage handler from completing
   - Fix: Restructure logic to check/clear BEFORE appending, store firstChild in variable

2. **WebSocket test interception timing** (`web-ui.test.js`)
   - Symptom: Test reports 0 messages received
   - Cause: Interceptor injected AFTER page load, missing initial snapshot
   - The server was working correctly all along!
   - Fix: Use `page.addInitScript()` to inject BEFORE navigation

**Commit**: 6e5f61b - All bugs fixed, E2E tests passing

**Tasks**:
- [✅] Fix activity log JavaScript error (DOM manipulation race condition)
- [✅] Fix WebSocket test (inject interceptor before page load)
- [✅] Fix timer update mechanism (working correctly all along)
- [✅] Fix provider cards rendering (working correctly all along)
- [✅] Run E2E tests: `npx playwright test`
- [✅] Verify: All 10 tests pass

**Success Criteria**: ✅ COMPLETE
- ✅ E2E tests pass: 10/10
- ✅ No console errors
- ✅ UI renders correctly
- ✅ Data flows properly

---

### Phase 2.5: UI Modernization (22 hours) - **PRIORITY 1.5**

**Goal**: Transform dashboard with modern design, smooth animations, and real-time feel

**Philosophy**: Hybrid approach combining Nordic elegance + Command Center clarity + tasteful Neon delight

**Tech Stack**:
- **CSS**: Tailwind CSS 3.x (utility-first, highly customizable)
- **Components**: DaisyUI (themed Tailwind components)
- **Animation**: Framer Motion (smooth React animations)
- **Charts**: ApexCharts (real-time updates, beautiful animations)
- **Icons**: Lucide Icons (clean, modern, extensive)

**Design System**:
```
Dark Mode (default):
- Background: #0f0f23 (deep blue-black)
- Surface: #1a1a3e (cards, panels)
- Text: #e5e7eb (soft white)
- Accent: #60a5fa (vibrant blue)
- Success: #34d399 (emerald green)
- Error: #f87171 (coral red)
- Warning: #fbbf24 (amber)

Typography:
- Display/Body: Inter (400, 500, 600, 700)
- Monospace: JetBrains Mono (timestamps, IDs, code)

Animation Timing:
- Fast: 150ms (micro-interactions)
- Normal: 250ms (state changes)
- Slow: 350ms (page transitions)
- Easing: Spring physics via Framer Motion
```

**Key Features to Implement**:
1. **Pulse effects** - Active runs glow subtly
2. **Smooth transitions** - Status changes morph colors
3. **Animated charts** - Data points update with motion
4. **Breathing cards** - Subtle scale animation on activity
5. **Skeleton screens** - Better loading experience
6. **Hover effects** - Scale to 102%, smooth and responsive
7. **Entrance animations** - New items slide in gracefully

---

#### Phase 2.5.1: Foundation Setup (4 hours)

**Goal**: Install and configure the modern tech stack

**Tasks**:
- [🚧] Install Tailwind CSS via npm
  - `npm install -D tailwindcss postcss autoprefixer`
  - `npx tailwindcss init -p`
- [ ] Install DaisyUI plugin
  - `npm install -D daisyui@latest`
  - Configure in `tailwind.config.js`
- [ ] Install Framer Motion
  - `npm install framer-motion`
- [ ] Install ApexCharts
  - `npm install apexcharts`
- [ ] Install Lucide Icons
  - `npm install lucide-react` or use CDN
- [ ] Configure Tailwind design tokens
  - Set custom colors in config
  - Set typography scale
  - Set spacing scale
- [ ] Create base CSS file with Tailwind directives
  - `@tailwind base; @tailwind components; @tailwind utilities;`
- [ ] Test build process works
  - Verify Tailwind classes compile
  - Check for conflicts with existing CSS

**Success Criteria**:
- All packages installed
- Tailwind compiles successfully
- No build errors
- Can use basic Tailwind classes

---

#### Phase 2.5.2: Component Rebuild (8 hours)

**Goal**: Rebuild core UI components with new design system

**Tasks**:
- [ ] **Stats Tiles** (1.5 hours)
  - Replace existing tiles with DaisyUI stat components
  - Add Framer Motion for number counting animation
  - Add pulse effect for changing values
  - Gradient backgrounds for visual interest
  
- [ ] **Active Runs Table** (2 hours)
  - Rebuild table with Tailwind utilities
  - Add row hover effects (scale, background)
  - Status badges with appropriate colors
  - Smooth entrance animation for new rows
  - Exit animation for removed rows
  
- [ ] **Provider Cards** (1.5 hours)
  - Card component with soft shadows
  - Health indicator with pulse when active
  - Hover effect with lift animation
  - Status colors that transition smoothly
  
- [ ] **Activity Feed** (1.5 hours)
  - Rebuild with stacked timeline design
  - Slide-in animation for new entries
  - Icon system using Lucide
  - Timestamp formatting
  - Auto-scroll behavior
  
- [ ] **Charts** (1.5 hours)
  - Replace Chart.js with ApexCharts
  - Configure real-time update animations
  - Dark mode colors
  - Smooth data transitions
  - Responsive sizing

**Success Criteria**:
- All components render correctly
- Animations smooth at 60fps
- Dark mode looks professional
- Responsive on desktop sizes

---

#### Phase 2.5.3: Real-Time Feel (6 hours)

**Goal**: Add animations that emphasize the live, real-time nature

**Tasks**:
- [ ] **Pulse Effects** (1.5 hours)
  - Active run rows pulse subtly
  - Provider health indicators breathe
  - New data indicators flash briefly
  - Implement with CSS animations + Framer Motion
  
- [ ] **State Transitions** (2 hours)
  - Run status changes morph colors smoothly
  - Success/failure animations
  - Queue → Running → Complete transitions
  - Use Framer Motion's AnimatePresence
  
- [ ] **Data Updates** (1.5 hours)
  - Numbers count up/down with animation
  - Chart data points animate in
  - Table rows reorder smoothly
  - WebSocket message arrival feedback
  
- [ ] **Loading States** (1 hour)
  - Skeleton screens for initial load
  - Shimmer effect on loading
  - Smooth content replacement
  - No layout shifts

**Success Criteria**:
- Feels alive and responsive
- Clear feedback for all state changes
- No jarring transitions
- Performance stays at 60fps

---

#### Phase 2.5.4: Polish & Performance (4 hours)

**Goal**: Micro-interactions and optimization

**Tasks**:
- [ ] **Micro-interactions** (1.5 hours)
  - Button hover/press states
  - Ripple effect on clicks
  - Tooltip animations
  - Focus states for accessibility
  - Copy feedback (success toast)
  
- [ ] **Performance Optimization** (1.5 hours)
  - Enable GPU acceleration (transform3d, will-change)
  - Optimize re-renders (React.memo where needed)
  - Debounce WebSocket updates
  - Lazy load off-screen content
  - Reduce bundle size (tree shaking)
  
- [ ] **Accessibility** (0.5 hour)
  - Keyboard navigation works
  - Screen reader labels
  - Focus indicators visible
  - Color contrast meets WCAG AA
  
- [ ] **Responsive Check** (0.5 hour)
  - Test on 1280px, 1440px, 1920px widths
  - Mobile view (hide/adapt as needed)
  - No horizontal scroll
  - Touch targets adequate size

**Success Criteria**:
- Time to First Paint < 1s
- 60fps maintained during animations
- Accessible to keyboard users
- Zero layout shifts
- Feels polished and professional

---

**Overall Success Metrics**:
- ✅ Modern, professional appearance
- ✅ Feels alive and real-time
- ✅ Smooth 60fps animations
- ✅ No performance degradation
- ✅ Maintainable with popular libraries
- ✅ Sets foundation for future features

**Inspiration Examples**:
- Vercel Dashboard (https://vercel.com/dashboard)
- Linear (https://linear.app)
- Railway (https://railway.app)
- Stripe Dashboard (https://dashboard.stripe.com)

**Reference Documents**:
- Full proposals: `tmp/ui-modernization-proposals.md`
- Executive summary: `tmp/ui-modernization-summary.md`

---

### Phase 3: UI Simplification (2 hours) - **PRIORITY 2**

**Goal**: Simplify UI based on user feedback

**Remove** (no longer needed):
- [ ] Theme selector (lock to dark mode only)
  - Remove `<select id="themeSelector">` from `index.html`
  - Remove theme CSS variables for light mode
  - Keep only dark mode styles
  - Remove theme JS code from `app.js`

- [ ] Top refresh button
  - Remove `<button id="refreshBtn">` from header
  - Keep refresh per-provider (Phase 5)

- [ ] Alerts navigation link
  - Remove `<a href="/alerts.html">` from header
  - Remove `alerts.html` and `alerts.js` files
  - Alerts will be in activity feed instead

- [ ] Workspace selector (for now)
  - Remove workspace dropdown
  - Remove workspace save button
  - Keep code in `app.js` (may restore later)

- [ ] Provider/Status filter buttons
  - Remove `+ Provider` and `+ Status` buttons
  - Filtering will be YAML-based only

**Reorganize**:
- [ ] Move search bar above "Active Runs" section
  - Remove from header controls
  - Place directly above runs table
  - Keep search functionality intact

- [ ] Simplify header
  - Just: Logo, Settings link, Version (move to footer?)
  - Clean, minimal design

**Tasks**:
- [ ] Update `index.html` structure
- [ ] Update `style.css` for new layout
- [ ] Update `app.js` to remove unused functions
- [ ] Test responsive design (mobile, tablet, desktop)
- [ ] Take screenshots for documentation

**Success Criteria**:
- Simpler, cleaner UI
- Less cognitive load
- Better visual hierarchy
- Responsive on all screen sizes

---

### Phase 4: Activity Feed Redesign (3 hours) - **PRIORITY 2**

**Goal**: Create always-visible activity feed showing every event

**Current**: Collapsible activity log in header  
**New**: Prominent feed below stats tiles, always visible

**Design**:
```
┌─────────────────────────────────────────────┐
│ Stats Tiles (Running, Queued, Success, Fail)│
└─────────────────────────────────────────────┘
┌─────────────────────────────────────────────┐
│ 🔔 Activity Feed                      [Clear]│
├─────────────────────────────────────────────┤
│ 🟢 12:34:56 — WebSocket connected            │
│ 📡 12:35:01 — github: Refreshed 45 runs      │
│ 🎯 12:35:15 — Webhook received: workflow_run │
│    └─ ceye/main: Build completed ✅          │
│ 🔄 12:35:20 — azure: Polling (15s interval)  │
│ ❌ 12:35:45 — azure: Auth required           │
│    └─ Run: gh auth login                     │
│ 🔍 12:36:00 — Filter: status=failed          │
│ (auto-scrolls to bottom)                     │
└─────────────────────────────────────────────┘
```

**Event Types** (with icons and colors):
- 🟢 WebSocket connect (green)
- 🔴 WebSocket disconnect (red)
- 📡 Provider refresh (blue)
- 🎯 Webhook received (purple)
- 🔄 Provider poll (gray)
- ❌ Error (red with action button)
- 🔍 Filter change (yellow)
- ✅ Success event (green)
- ⚠️ Warning (orange)

**Features**:
- Auto-scroll to bottom
- Timestamps (relative: "2s ago" or absolute)
- Expandable details (click to see full payload)
- Color-coded by severity
- Clear button to empty feed
- Max 100 items (FIFO)
- Persists in session (lost on refresh)

**Tasks**:
- [ ] Update HTML structure in `index.html`
- [ ] Add CSS for activity feed in `style.css`
- [ ] Update `app.js` activity functions
- [ ] Add event types with icons/colors
- [ ] Implement auto-scroll logic
- [ ] Add expandable details
- [ ] Add clear button
- [ ] Test with real events
- [ ] Test edge cases (rapid events, long messages)

**Success Criteria**:
- Activity feed always visible
- All events logged with proper formatting
- Auto-scrolls smoothly
- Easy to understand what's happening
- Helps debugging

---

### Phase 5: Provider Card Redesign (4 hours) - **PRIORITY 3**

**Goal**: Make provider cards interactive with refresh and auth status

**Current**: Simple health display  
**New**: Rich interactive cards

**Design**:
```
┌──────────────────────────────────────────┐
│ 🐙 GitHub                    ✅ Connected │
│ Last update: 2m ago          [🔄 Refresh] │
│ Active: 5 / Total: 45                     │
└──────────────────────────────────────────┘

┌──────────────────────────────────────────┐
│ 🔷 Azure DevOps              ❌ Auth Req  │
│ Not authenticated                         │
│ └─ Run: az login             [📋 Copy]   │
└──────────────────────────────────────────┘
```

**Auth Status Indicators**:
- ✅ Green: Authenticated and working
- ⚠️ Yellow: Working without auth (webhook-only)
- ❌ Red: Authentication required
- 🔄 Gray: Refreshing...

**Features**:
- Per-provider refresh button
- Auth status detection (check `gh` / `az` CLI)
- Helpful error messages
- Copyable fix commands
- Last update timestamp
- Run counts (active/total)
- Loading animation during refresh
- Pulse animation on webhook received

**Tasks**:
- [ ] Redesign provider card HTML
- [ ] Add CSS for new card design
- [ ] Implement auth detection:
  - Check `gh auth status` for GitHub
  - Check `az account show` for Azure
  - Cache results (check every 60s)
- [ ] Add per-provider refresh API endpoint
- [ ] Add refresh button with loading state
- [ ] Add helpful error messages
- [ ] Add copy-to-clipboard for commands
- [ ] Add webhook pulse animation
- [ ] Test all auth states
- [ ] Test refresh functionality

**Success Criteria**:
- Clear provider status at a glance
- Easy to fix auth issues
- Per-provider refresh works
- Helpful, actionable feedback

---

### Phase 6: Webhook Visual Feedback (2 hours) - **PRIORITY 3**

**Goal**: Show visual feedback when webhooks arrive

**Features**:
1. **Provider Card Pulse**: When webhook arrives for a provider
   - Gentle pulse animation (border glow)
   - Lasts 2 seconds
   - Color matches provider

2. **Activity Feed Entry**: Log webhook events
   - "🎯 Webhook received: workflow_run"
   - Show repo, workflow, status
   - Expandable to show full payload
   - Link to GitHub/Azure run

3. **Stats Update Animation**: When counts change
   - Brief highlight on changed stat tile
   - Smooth number transitions

**Tasks**:
- [ ] Add pulse animation CSS
- [ ] Trigger pulse on webhook event
- [ ] Add webhook events to activity feed
- [ ] Format webhook payload nicely
- [ ] Add stat tile update animation
- [ ] Test with real GitHub webhooks
- [ ] Test with ngrok tunnel
- [ ] Document webhook setup in README

**Success Criteria**:
- Immediate visual feedback on webhook
- Easy to see which provider updated
- Activity feed shows details
- Smooth, polished animations

---

### Phase 7: Debugging Dashboard (4 hours) - **PRIORITY 4**

**Goal**: Make dashboard a powerful debugging tool for development

**Features**:

1. **Debug Panel** (collapsible, off by default)
   - Toggle via Settings or keyboard shortcut (D)
   - Docked at bottom or side
   - Tabbed interface

2. **Server Logs Tab**
   - Stream server logs to browser via WebSocket
   - Filter by level (debug, info, warn, error)
   - Search/filter logs
   - Auto-scroll toggle
   - Download logs button

3. **WebSocket Inspector Tab**
   - Show all messages sent/received
   - Timestamp, direction, payload
   - Expandable JSON viewer
   - Message counter
   - Connection health

4. **Event Flow Tab**
   - Visualize: Provider → Store → UI flow
   - Show event latency
   - Show broadcast timing
   - Event queue depth

5. **Performance Tab**
   - Event processing time (p50, p95, p99)
   - Broadcast latency
   - WebSocket message rate
   - Provider poll intervals
   - Memory usage (if available)

6. **API Rate Limits Tab**
   - GitHub API: remaining calls / reset time
   - Azure API: rate limit status
   - Show which operations consumed calls
   - Warning when approaching limit

7. **Error Stack Traces**
   - Full Go stack traces for panics
   - JavaScript errors with source maps
   - Click to copy

**Implementation**:
- [ ] Add debug mode flag to config
- [ ] Create debug WebSocket endpoint (`/ws/debug`)
- [ ] Stream server logs to debug WebSocket
- [ ] Add debug panel HTML/CSS
- [ ] Implement tabs with Vue or plain JS
- [ ] Add server-side telemetry collection
- [ ] Add API rate limit tracking
- [ ] Add performance metrics collection
- [ ] Add event flow visualization
- [ ] Test in development mode
- [ ] Document in README

**Success Criteria**:
- No need to tail logs manually
- All debugging info in browser
- Easy to troubleshoot issues
- Helps during development

---

### Phase 8: Webhook Integration Fix (2 hours) - **PRIORITY 1**

**Goal**: Ensure webhooks work end-to-end with proper UI updates

**Current Issues**:
- Initial snapshot not broadcast to new WebSocket clients
- Webhook events may not flow to UI
- Polling still active when webhooks enabled

**Fix**:
1. **Initial Snapshot Broadcast**
   - When WebSocket client connects, send current store state
   - Location: `internal/server/server.go` `handleWebSocket()`

2. **Webhook → Store → Broadcast Flow**
   - Webhook handler receives GitHub/Azure event
   - Converts to `RunEvent`
   - Sends to provider's event channel
   - Store merges event
   - Store broadcasts to all WebSocket clients
   - Verify this flow works

3. **Disable Polling in Webhook Mode**
   - When `--webhooks` flag set
   - Provider does ONE initial poll
   - Then waits indefinitely (no more polls)
   - All updates come from webhooks

4. **Fallback Mechanism**
   - If no webhook received in 5 minutes
   - Optionally trigger a refresh
   - Or alert user "No webhooks received, check setup"

**Tasks**:
- [ ] Add initial snapshot broadcast on WebSocket connect
- [ ] Verify webhook → store → broadcast flow
- [ ] Test with demo webhook: `curl -X POST localhost:9090/webhooks/github -d '{...}'`
- [ ] Test with real GitHub webhook (via ngrok)
- [ ] Add webhook health check (last received time)
- [ ] Add fallback mechanism (optional refresh)
- [ ] Update activity feed to show webhook events
- [ ] Add webhook setup instructions to README

**Success Criteria**:
- New clients see current data immediately
- Webhooks update UI in real-time
- No redundant polling
- Clear feedback if webhooks not working

---

### Phase 9: E2E Tests Update (3 hours) - **PRIORITY 4**

**Goal**: Comprehensive E2E test coverage for new UI

**Current State**: 10 tests, 3 failing

**Fix Existing Tests**:
- [ ] Test 2: Timer updates (should show timestamp)
- [ ] Test 3: Provider count (should be 1, not 2)
- [ ] Test 9: WebSocket receives messages

**Add New Tests**:
- [ ] Activity feed renders and updates
- [ ] Activity feed shows different event types
- [ ] Activity feed auto-scrolls
- [ ] Activity feed clear button works
- [ ] Provider cards render correctly
- [ ] Provider refresh button works
- [ ] Auth status indicators show correctly
- [ ] Webhook events appear in activity feed
- [ ] Search bar above runs works
- [ ] Debug panel toggle works
- [ ] Debug panel tabs work
- [ ] Server logs stream to debug panel

**Screenshot Regression Tests**:
- [ ] Full UI screenshot
- [ ] Activity feed screenshot
- [ ] Provider cards screenshot
- [ ] Debug panel screenshot
- [ ] Compare against baseline

**Tasks**:
- [ ] Fix 3 failing tests
- [ ] Add 12 new tests
- [ ] Add screenshot regression tests
- [ ] Run full suite: `npx playwright test`
- [ ] Generate HTML report: `npx playwright show-report`
- [ ] Update test documentation

**Success Criteria**:
- All 25+ tests pass ✅
- <1 minute test run time
- Screenshots for visual regression
- Tests catch regressions

---

### Phase 10: Documentation Update (1 hour) - **PRIORITY 4**

**Goal**: Update all docs to reflect new architecture

**Files to Update**:

1. **`docs/plan.md`** (this file)
   - Mark TUI as removed
   - Update architecture diagram
   - Update feature list

2. **`docs/agents.md`**
   - Remove all TUI sections
   - Remove TUI testing guide
   - Add UI testing guide
   - Add debugging guide
   - Update quick reference

3. **`docs/readme.md`**
   - Remove TUI screenshots
   - Add new dashboard screenshots
   - Update quick start (web-only)
   - Add webhook setup guide
   - Add debugging section
   - Update config examples

4. **`README.md`** (symlink to docs/readme.md)
   - Already updated via symlink

5. **Add New Docs**:
   - `docs/references/debugging-guide.md` - Using debug panel
   - `docs/references/webhook-setup.md` - Complete webhook setup
   - Update `docs/references/web-ui-architecture.md` - New design

**Tasks**:
- [ ] Update `docs/plan.md`
- [ ] Update `docs/agents.md`
- [ ] Update `docs/readme.md`
- [ ] Create `docs/references/debugging-guide.md`
- [ ] Update `docs/references/webhook-setup.md`
- [ ] Update architecture diagrams
- [ ] Take new screenshots
- [ ] Add troubleshooting section
- [ ] Review all docs for accuracy

**Success Criteria**:
- Accurate, up-to-date docs
- No mention of TUI (removed)
- Clear setup instructions
- Helpful troubleshooting
- Professional presentation

---

## Previous Completed Features

### Core Features (100%)
- TUI with Bubble Tea
- Dashboard with WebSocket
- Provider abstraction
- Real-time updates
- SafeProvider wrapper
- Contract testing
- Integration testing

### Providers (100%)
- GitHub Actions ✅
- Azure DevOps ✅
- GitLab CI ✅
- Demo provider ✅

### Monitoring (100%)
- Historical data (SQLite)
- Trends & analytics
- Success rate tracking
- Duration analysis
- Build frequency metrics

### Alerting (100%)
- Alert engine with 4 conditions
- 3 notification channels (Slack, Webhook, Log)
- Cooldown & rate limiting
- Alert history
- TUI alert panel
- Web alerts page
- Alert details modal
- Rule statistics

### User Experience (100%)
All 5 phases complete:
- C.1: Keyboard Shortcuts (6 shortcuts + help)
- C.2: Theme System (4 themes)
- C.3: Advanced Filtering (multi-select + pills)
- C.4: Workspaces (named presets)
- C.5: Settings Page (9 preferences)

**Total Time**: 5.5 hours (vs 40 hours planned!)

---

---

## Task Summary

| Phase | Priority | Hours | Status |
|-------|----------|-------|--------|
| **0. React Migration** | **P0 🔥🔥🔥** | **22h** | **🚧 IN PROGRESS** |
| 0.1. Foundation Setup | P0 🔥🔥🔥 | 4h | ⏳ **NEXT UP** |
| 0.2. Core Components | P0 🔥🔥🔥 | 8h | ⏳ Blocked by 0.1 |
| 0.3. Real-Time Integration | P0 🔥🔥🔥 | 4h | ⏳ Blocked by 0.2 |
| 0.4. Polish & Excellence | P0 🔥🔥🔥 | 6h | ⏳ Blocked by 0.3 |
| 1. Remove TUI | P1 | 2h | ✅ Done |
| 1.5. Remove "Web UI" Term | P1 | 0.25h | ✅ Done |
| 2. Fix Critical Bugs | P1 | 1h | ✅ Done |
| 3-10. Old Phases | P2-P4 | 22h | ❌ DEPRECATED (superseded by Phase 0) |
| **Total** | | **22h** | |

**Next Steps**:
1. ✅ Phase 1 (Remove TUI) - DONE
2. ✅ Phase 1.5 (Remove "Web UI" terminology) - DONE  
3. ✅ Phase 2 (Fix Bugs) - DONE
4. 🚧 Phase 0.1 (React Foundation) - **STARTING NOW** 🚀
5. Build the most beautiful CI/CD dashboard ever!

---


### Advanced Testing (Not needed yet)
- Load testing
- Chaos engineering
- E2E tests
- Performance benchmarks

### Additional Providers (As needed)
- Jenkins
- CircleCI
- Buildkite
- Travis CI

### Enterprise Features (When needed)
- Authentication
- RBAC
- Audit logging
- Multi-tenancy

---

## Deployment

### Quick Start

```bash
# Build
go build -o bin/ceye ./cmd/ceye

# Run TUI
ceye

# Run Web
ceye --web --port 8080

# With demo
ceye --demo --demo-duration 5m
```

### Configuration

Create `ceye.yaml`:

```yaml
providers:
  - type: github
    repos:
      - owner: "myorg"
        repo: "myrepo"
  
  - type: azure
    org: "myorg"
    projects:
      - name: "MyProject"
        pipelines: [123]

alerting:
  rules:
    - name: "Prod Failures"
      condition: "workflow_failed"
      severity: "critical"
      
  channels:
    - type: "slack"
      webhook: "${SLACK_WEBHOOK_URL}"
      
server:
  port: 8080
```

### Environment Variables

```bash
export GITHUB_TOKEN="ghp_..."
export AZURE_PAT="..."
export SLACK_WEBHOOK_URL="https://..."
```

---

## Project Statistics

- **Development Time**: ~6 weeks
- **Lines of Code**: ~8,000+
- **Tests**: 175+ (all passing)
- **Features**: 50+ major features
- **Bugs**: 0 known issues

---

## Next Steps

1. Deploy to production
2. Use with real CI/CD systems
3. Gather user feedback
4. Add features as needed

**The project is feature-complete!** 🚀

---

## Documentation

- **README**: `/docs/readme.md`
- **Agent Context**: `/docs/agents.md`
- **Testing Guide**: `/docs/references/testing-guide.md`
- **Webhook Guide**: `/docs/references/webhook-guide.md`

---

**Status**: ✅ COMPLETE - Ready for production!
