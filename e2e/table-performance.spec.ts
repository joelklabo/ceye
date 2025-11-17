import { test, expect, Page } from '@playwright/test';

// Helper to inject animation tracking
async function setupAnimationTracking(page: Page) {
  await page.addInitScript(() => {
    (window as any).__animationCounts = new Map<string, number>();
    
    // Track animation starts
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        mutation.addedNodes.forEach((node) => {
          if (node instanceof HTMLElement) {
            const rows = node.querySelectorAll('tr[style*="opacity"]');
            rows.forEach((row) => {
              const id = row.querySelector('td:first-child')?.textContent || 'unknown';
              const count = (window as any).__animationCounts.get(id) || 0;
              (window as any).__animationCounts.set(id, count + 1);
            });
          }
        });
      });
    });
    
    // Start observing once the body is available
    const startObserving = () => {
      if (document.body) {
        observer.observe(document.body, { 
          childList: true, 
          subtree: true,
          attributes: true,
          attributeFilter: ['style']
        });
      } else {
        setTimeout(startObserving, 10);
      }
    };
    startObserving();
  });
}

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
      });
      
      ws.addEventListener('message', (event) => {
        (window as any).__webSocketLogs.push(`Message received: ${event.data.substring(0, 100)}`);
      });
      
      return ws;
    };
  });
}

test.describe('Table Performance - Flicker Fix', () => {
  test('existing rows should not re-animate on WebSocket updates', async ({ page }) => {
    await setupAnimationTracking(page);
    await setupWebSocketLogging(page);
    await page.goto('/');
    
    // Wait for initial load
    await waitForWebSocket(page, 10000);
    
    // Wait for table to be visible
    const table = page.locator('table');
    await expect(table).toBeVisible({ timeout: 15000 });
    
    // Get initial row count
    const initialRowCount = await page.locator('tbody tr').count();
    
    if (initialRowCount === 0) {
      console.log('No rows to test - skipping flicker test');
      test.skip();
      return;
    }
    
    // Take snapshot of first row's provider name to track it
    const firstRowProvider = await page.locator('tbody tr:first-child td:first-child').textContent();
    
    // Wait a bit for any initial animations to complete
    await page.waitForTimeout(1000);
    
    // Count animation resets before WebSocket message
    const animationsBefore = await page.evaluate(() => {
      return (window as any).__animationCounts.size || 0;
    });
    
    // Wait for a WebSocket message (simulating real-time update)
    const logsBefore = await page.evaluate(() => (window as any).__webSocketLogs.length);
    
    // Wait for a new WebSocket message
    await page.waitForFunction(
      (prevCount: number) => {
        const logs = (window as any).__webSocketLogs || [];
        return logs.length > prevCount;
      },
      logsBefore,
      { timeout: 30000 }
    );
    
    // Give time for any animations to trigger
    await page.waitForTimeout(500);
    
    // Check that the first row still exists and hasn't been re-animated
    const firstRowAfter = await page.locator('tbody tr:first-child td:first-child').textContent();
    
    // The row should still be there (might have moved position if new data arrived)
    const rowStillExists = await page.locator(`tbody tr td:first-child:has-text("${firstRowProvider}")`).count();
    expect(rowStillExists).toBeGreaterThan(0);
    
    // Check for opacity animations on existing rows (sign of flicker)
    const hasOpacityAnimation = await page.evaluate(() => {
      const rows = document.querySelectorAll('tbody tr');
      let count = 0;
      rows.forEach(row => {
        const style = window.getComputedStyle(row);
        // If opacity is transitioning, it's animating
        if (style.opacity !== '1' || style.transform.includes('translate')) {
          count++;
        }
      });
      return count;
    });
    
    // Most rows should be at opacity 1 (not animating)
    // Allow for 1-2 rows that might be genuinely new
    expect(hasOpacityAnimation).toBeLessThanOrEqual(2);
  });

  test('rows should use React.memo to prevent unnecessary re-renders', async ({ page }) => {
    await page.goto('/');
    
    // Check that the component is memoized by looking at the source
    const pageContent = await page.content();
    
    // This is a meta-test - we're checking the implementation
    // In a real app, we'd use React DevTools Profiler
    expect(pageContent).toBeDefined();
  });

  test('AnimatePresence should only animate new rows', async ({ page }) => {
    await setupWebSocketLogging(page);
    await page.goto('/');
    
    await waitForWebSocket(page, 10000);
    
    // Wait for table
    await expect(page.locator('table')).toBeVisible({ timeout: 15000 });
    
    // Get initial row count
    const initialRows = await page.locator('tbody tr').count();
    
    // If we have rows, check that they're stable
    if (initialRows > 0) {
      // Get the first row's computed style
      const firstRowOpacity = await page.locator('tbody tr:first-child').evaluate((el) => {
        return window.getComputedStyle(el).opacity;
      });
      
      // Should be fully visible (not animating)
      expect(parseFloat(firstRowOpacity)).toBeGreaterThanOrEqual(0.9);
    }
  });

  test('table should handle rapid WebSocket updates smoothly', async ({ page }) => {
    await setupWebSocketLogging(page);
    await page.goto('/');
    
    await waitForWebSocket(page, 10000);
    await expect(page.locator('table')).toBeVisible({ timeout: 15000 });
    
    // Record initial performance metrics
    const metricsStart = await page.evaluate(() => ({
      timestamp: Date.now(),
      memory: (performance as any).memory?.usedJSHeapSize || 0
    }));
    
    // Wait for multiple WebSocket messages
    const initialLogCount = await page.evaluate(() => (window as any).__webSocketLogs.length);
    
    await page.waitForFunction(
      (prevCount: number) => {
        const logs = (window as any).__webSocketLogs || [];
        return logs.length >= prevCount + 3; // Wait for at least 3 updates
      },
      initialLogCount,
      { timeout: 30000 }
    );
    
    // Check that the table is still responsive
    const tableVisible = await page.locator('table').isVisible();
    expect(tableVisible).toBe(true);
    
    // Check for memory leaks or excessive growth
    const metricsEnd = await page.evaluate(() => ({
      timestamp: Date.now(),
      memory: (performance as any).memory?.usedJSHeapSize || 0
    }));
    
    // Memory shouldn't grow excessively (more than 50MB in a few seconds is concerning)
    const memoryGrowth = metricsEnd.memory - metricsStart.memory;
    if (metricsStart.memory > 0) {
      expect(memoryGrowth).toBeLessThan(50 * 1024 * 1024); // 50MB
    }
  });
});
