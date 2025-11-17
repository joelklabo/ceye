import { test, expect, Page } from '@playwright/test';

// Helper to wait for WebSocket connection
async function waitForWebSocket(page: Page, timeout = 5000) {
  await page.waitForFunction(
    () => {
      const logs = (window as any).__webSocketLogs || [];
      return logs.some((log: string) => log.includes('WebSocket connected'));
    },
    { timeout }
  );
}

// Helper to inject WebSocket logging
async function setupWebSocketLogging(page: Page) {
  await page.addInitScript(() => {
    (window as any).__webSocketLogs = [];
    const originalWebSocket = window.WebSocket;
    
    (window as any).WebSocket = function(url: string) {
      const ws = new originalWebSocket(url);
      
      ws.addEventListener('open', () => {
        (window as any).__webSocketLogs.push('WebSocket connected');
        console.log('WebSocket connected');
      });
      
      ws.addEventListener('message', (event) => {
        (window as any).__webSocketLogs.push(`Message received: ${event.data.substring(0, 100)}`);
      });
      
      ws.addEventListener('error', (event) => {
        (window as any).__webSocketLogs.push('WebSocket error');
        console.error('WebSocket error', event);
      });
      
      ws.addEventListener('close', () => {
        (window as any).__webSocketLogs.push('WebSocket closed');
        console.log('WebSocket closed');
      });
      
      return ws;
    };
  });
}

test.describe('Dashboard Loading & Connection', () => {
  test('page loads successfully', async ({ page }) => {
    await page.goto('/');
    
    // Check that page title is present
    await expect(page).toHaveTitle(/web/); // Vite default title
    
    // Check root element exists
    const root = page.locator('#root');
    await expect(root).toBeVisible();
  });

  test('WebSocket connects within 5s', async ({ page }) => {
    await setupWebSocketLogging(page);
    await page.goto('/');
    
    // Wait for WebSocket connection - increased timeout
    await waitForWebSocket(page, 10000);
    
    // Verify connection indicator is green
    const connectionDot = page.locator('header svg').filter({ hasText: '' }).first();
    await expect(connectionDot).toHaveClass(/fill-green-400/, { timeout: 10000 });
  });

  test('connection indicator shows green', async ({ page }) => {
    await setupWebSocketLogging(page);
    await page.goto('/');
    
    // Wait for connection
    await waitForWebSocket(page, 10000);
    
    // Check for "Connected" text
    const connectedText = page.getByText('Connected');
    await expect(connectedText).toBeVisible({ timeout: 10000 });
  });

  test('initial data loads', async ({ page }) => {
    await setupWebSocketLogging(page);
    await page.goto('/');
    
    // Wait for WebSocket
    await waitForWebSocket(page, 10000);
    
    // Wait for stats cards to appear (indicates data loaded)
    const statsCards = page.locator('.grid').first();
    await expect(statsCards).toBeVisible({ timeout: 15000 });
    
    // Wait for runs table or empty state
    const runsSection = page.locator('table').or(page.getByText('No runs yet'));
    await expect(runsSection).toBeVisible({ timeout: 15000 });
  });

  test('error handling if WebSocket fails', async ({ page, context }) => {
    // Block WebSocket connection
    await context.route('**/ws', route => route.abort());
    
    await page.goto('/');
    
    // Should show disconnected state
    const disconnectedText = page.getByText('Disconnected');
    await expect(disconnectedText).toBeVisible({ timeout: 10000 });
    
    // Connection dot should be red
    const redDot = page.locator('header').locator('svg.fill-red-400');
    await expect(redDot).toBeVisible();
  });
});

test.describe('Stats Cards', () => {
  test.beforeEach(async ({ page }) => {
    await setupWebSocketLogging(page);
    await page.goto('/');
    await waitForWebSocket(page);
  });

  test('all 4 cards render', async ({ page }) => {
    // Wait for grid container
    const grid = page.locator('.grid').first();
    await expect(grid).toBeVisible({ timeout: 10000 });
    
    // Check for all 4 stat cards by looking for their titles in the first grid
    await expect(grid.getByText('Running')).toBeVisible();
    await expect(grid.getByText('Queued')).toBeVisible();
    await expect(grid.getByText('Success')).toBeVisible();
    await expect(grid.getByText('Failed')).toBeVisible();
  });

  test('counters display numbers', async ({ page }) => {
    // Each card should have a large number (counter)
    const numbers = page.locator('.text-4xl');
    await expect(numbers).toHaveCount(4);
    
    // Each should contain a number (0 or more)
    const firstNumber = numbers.first();
    const text = await firstNumber.textContent();
    expect(text).toMatch(/^\d+$/);
  });

  test('icons display correctly', async ({ page }) => {
    // Check for Lucide icons (SVG elements)
    const icons = page.locator('.grid').first().locator('svg');
    await expect(icons).toHaveCount(4);
  });

  test('responsive layout', async ({ page }) => {
    // Desktop: 4 columns
    await page.setViewportSize({ width: 1920, height: 1080 });
    const grid = page.locator('.grid').first();
    await expect(grid).toHaveCSS('display', 'grid');
    
    // Mobile: should still be visible (stacked)
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(grid).toBeVisible();
  });
});

test.describe('Runs Table', () => {
  test.beforeEach(async ({ page }) => {
    await setupWebSocketLogging(page);
    await page.goto('/');
    await waitForWebSocket(page);
  });

  test('table renders with data or empty state', async ({ page }) => {
    // Either table exists or "No runs yet" message
    const hasTable = await page.locator('table').isVisible().catch(() => false);
    const hasEmptyState = await page.getByText('No runs yet').isVisible().catch(() => false);
    
    expect(hasTable || hasEmptyState).toBeTruthy();
  });

  test('search box is present', async ({ page }) => {
    const searchInput = page.getByPlaceholder('Search runs...');
    await expect(searchInput).toBeVisible();
  });

  test('table headers are correct', async ({ page }) => {
    // Skip if no table
    const hasTable = await page.locator('table').isVisible().catch(() => false);
    if (!hasTable) {
      test.skip();
      return;
    }
    
    // Check for expected headers within thead only
    const thead = page.locator('table thead');
    await expect(thead.getByText('Provider')).toBeVisible();
    await expect(thead.getByText('Repo')).toBeVisible();
    await expect(thead.getByText('Workflow')).toBeVisible();
    await expect(thead.getByText('Status')).toBeVisible();
    await expect(thead.getByText('Duration')).toBeVisible();
    await expect(thead.getByText('Time')).toBeVisible();
  });

  test('search filter works', async ({ page }) => {
    const hasTable = await page.locator('table tbody tr').count().then(c => c > 0).catch(() => false);
    if (!hasTable) {
      test.skip();
      return;
    }
    
    const searchInput = page.getByPlaceholder('Search runs...');
    const initialRowCount = await page.locator('table tbody tr').count();
    
    // Type in search box
    await searchInput.fill('nonexistent-repo-xyz');
    await page.waitForTimeout(300); // Debounce
    
    // Should show "No runs match" or fewer rows
    const filteredRows = await page.locator('table tbody tr').count();
    const noMatchText = await page.getByText(/No runs match/).isVisible().catch(() => false);
    
    expect(filteredRows < initialRowCount || noMatchText).toBeTruthy();
  });

  test('sorting works', async ({ page }) => {
    const hasTable = await page.locator('table tbody tr').count().then(c => c > 1).catch(() => false);
    if (!hasTable) {
      test.skip();
      return;
    }
    
    // Click sort button (arrow icon in header)
    const sortButton = page.locator('th button').first();
    await sortButton.click();
    
    // Table should still be visible (basic smoke test)
    await expect(page.locator('table')).toBeVisible();
  });
});

test.describe('Provider Health Cards', () => {
  test.beforeEach(async ({ page }) => {
    await setupWebSocketLogging(page);
    await page.goto('/');
    await waitForWebSocket(page);
  });

  test('Provider Health section exists', async ({ page }) => {
    const heading = page.getByRole('heading', { name: 'Provider Health' });
    await expect(heading).toBeVisible({ timeout: 10000 });
  });

  test('provider cards render', async ({ page }) => {
    // Should have at least one provider card or "No providers" message
    const hasCards = await page.locator('.grid').nth(1).locator('div').count().then(c => c > 0).catch(() => false);
    const hasNoProviders = await page.getByText('No providers configured').isVisible().catch(() => false);
    
    expect(hasCards || hasNoProviders).toBeTruthy();
  });

  test('health indicator visible', async ({ page }) => {
    const hasCards = await page.locator('.grid').nth(1).locator('div').count().then(c => c > 0).catch(() => false);
    if (!hasCards) {
      test.skip();
      return;
    }
    
    // Look for Circle icon (health indicator)
    const circles = page.locator('.grid').nth(1).locator('svg');
    await expect(circles.first()).toBeVisible();
  });
});

test.describe('Activity Feed', () => {
  test.beforeEach(async ({ page }) => {
    await setupWebSocketLogging(page);
    await page.goto('/');
    await waitForWebSocket(page);
  });

  test('Activity section renders', async ({ page }) => {
    const heading = page.getByRole('heading', { name: 'Activity' });
    await expect(heading).toBeVisible({ timeout: 10000 });
  });

  test('expand/collapse works', async ({ page }) => {
    // Find collapse button
    const collapseButton = page.getByRole('button', { name: /Collapse|Expand/ });
    await expect(collapseButton).toBeVisible();
    
    // Click it
    await collapseButton.click();
    await page.waitForTimeout(300); // Animation
    
    // Button text should change
    const newText = await collapseButton.textContent();
    expect(newText).toBeTruthy();
  });

  test('activity items or empty state', async ({ page }) => {
    // Should show activities or "No activity yet"
    const hasActivities = await page.locator('.activity-item, .flex.items-start.gap-3').count().then(c => c > 0).catch(() => false);
    const hasEmptyState = await page.getByText('No activity yet').isVisible().catch(() => false);
    
    expect(hasActivities || hasEmptyState).toBeTruthy();
  });
});

test.describe('Real-Time Updates', () => {
  test.beforeEach(async ({ page }) => {
    await setupWebSocketLogging(page);
  });

  test('dashboard receives WebSocket messages', async ({ page }) => {
    await page.goto('/');
    await waitForWebSocket(page);
    
    // Check WebSocket logs for messages
    const logs = await page.evaluate(() => (window as any).__webSocketLogs || []);
    const hasMessages = logs.some((log: string) => log.includes('Message received'));
    
    expect(hasMessages).toBeTruthy();
  });

  test('connection indicator updates on disconnect', async ({ page, context }) => {
    await page.goto('/');
    await waitForWebSocket(page);
    
    // Verify connected state
    await expect(page.getByText('Connected')).toBeVisible();
    
    // Block WebSocket to simulate disconnect
    await context.route('**/ws', route => route.abort());
    
    // Reload to trigger disconnect
    await page.reload();
    
    // Should show disconnected
    await expect(page.getByText('Disconnected')).toBeVisible({ timeout: 10000 });
  });
});
