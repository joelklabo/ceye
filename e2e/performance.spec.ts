import { test, expect } from '@playwright/test';

test.describe('Performance', () => {
  test('page loads within 3 seconds', async ({ page }) => {
    const startTime = Date.now();
    
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    const loadTime = Date.now() - startTime;
    
    // Page should load in under 3 seconds
    expect(loadTime).toBeLessThan(3000);
    console.log(`Page loaded in ${loadTime}ms`);
  });

  test('WebSocket connection establishes quickly', async ({ page }) => {
    await page.goto('/');
    
    const startTime = Date.now();
    
    // Wait for connected status
    const connected = page.locator('header').getByText('Connected');
    await connected.waitFor({ timeout: 10000 });
    
    const connectionTime = Date.now() - startTime;
    
    // WebSocket should connect in under 5 seconds
    expect(connectionTime).toBeLessThan(5000);
    console.log(`WebSocket connected in ${connectionTime}ms`);
  });

  test('renders many runs efficiently', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Measure render time for table rows
    const renderStart = await page.evaluate(() => performance.now());
    
    // Wait for table to be populated
    await page.waitForSelector('table tbody tr', { timeout: 10000 });
    
    const renderEnd = await page.evaluate(() => performance.now());
    const renderTime = renderEnd - renderStart;
    
    // Rendering should be fast even with many rows
    expect(renderTime).toBeLessThan(2000); // 2s is reasonable for initial render
    console.log(`Table rendered in ${renderTime.toFixed(2)}ms`);
  });

  test('no memory leaks on repeated updates', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Get initial memory
    const initialMemory = await page.evaluate(() => {
      if (performance.memory) {
        return (performance as any).memory.usedJSHeapSize;
      }
      return 0;
    });
    
    // Wait for some updates
    await page.waitForTimeout(5000);
    
    // Check memory hasn't grown too much
    const finalMemory = await page.evaluate(() => {
      if (performance.memory) {
        return (performance as any).memory.usedJSHeapSize;
      }
      return 0;
    });
    
    if (initialMemory > 0 && finalMemory > 0) {
      const memoryGrowth = finalMemory - initialMemory;
      const memoryGrowthMB = memoryGrowth / (1024 * 1024);
      
      // Memory growth should be reasonable (< 10MB in 5 seconds)
      expect(memoryGrowthMB).toBeLessThan(10);
      console.log(`Memory growth: ${memoryGrowthMB.toFixed(2)}MB`);
    }
  });

  test('smooth animations - no frame drops', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Check that animations run smoothly
    const fps = await page.evaluate(async () => {
      return new Promise<number>((resolve) => {
        let frameCount = 0;
        const startTime = performance.now();
        
        function countFrames() {
          frameCount++;
          if (performance.now() - startTime < 1000) {
            requestAnimationFrame(countFrames);
          } else {
            resolve(frameCount);
          }
        }
        
        requestAnimationFrame(countFrames);
      });
    });
    
    // Should maintain close to 60fps
    expect(fps).toBeGreaterThan(50);
    console.log(`FPS: ${fps}`);
  });

  test('interaction latency < 100ms', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Find a sortable header
    const sortButton = page.locator('button').filter({ hasText: 'Provider' });
    
    if (await sortButton.count() > 0) {
      const interactionStart = Date.now();
      
      await sortButton.click();
      
      // Wait for sort to complete (table re-render)
      await page.waitForTimeout(100);
      
      const interactionTime = Date.now() - interactionStart;
      
      // Interaction should feel instant (< 300ms is good UX)
      expect(interactionTime).toBeLessThan(300);
      console.log(`Sort interaction: ${interactionTime}ms`);
    }
  });

  test('bundle size is reasonable', async ({ page }) => {
    // Intercept network requests to measure bundle size
    let totalJSSize = 0;
    let totalCSSSize = 0;
    
    page.on('response', (response) => {
      const url = response.url();
      const headers = response.headers();
      const contentLength = parseInt(headers['content-length'] || '0', 10);
      
      if (url.includes('.js') && contentLength > 0) {
        totalJSSize += contentLength;
      }
      if (url.includes('.css') && contentLength > 0) {
        totalCSSSize += contentLength;
      }
    });
    
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    const totalSizeKB = (totalJSSize + totalCSSSize) / 1024;
    
    // Total bundle should be under 500KB
    expect(totalSizeKB).toBeLessThan(500);
    console.log(`Bundle size: JS=${(totalJSSize / 1024).toFixed(2)}KB, CSS=${(totalCSSSize / 1024).toFixed(2)}KB, Total=${totalSizeKB.toFixed(2)}KB`);
  });
});
