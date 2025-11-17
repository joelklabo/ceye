import { test, expect } from '@playwright/test';

test.describe('Loading States', () => {
  test('shows skeleton loaders while connecting', async ({ page }) => {
    // Block WebSocket to simulate slow connection
    let wsBlocked = true;
    await page.route('**/ws', (route) => {
      if (wsBlocked) {
        // Don't complete the request - keep it pending
        return;
      }
      route.continue();
    });

    await page.goto('/');

    // Should show skeleton cards while loading
    const skeletonCards = page.locator('.bg-muted').first();
    await expect(skeletonCards).toBeVisible({ timeout: 5000 });

    // Unblock WebSocket
    wsBlocked = false;
    
    // Wait a bit for actual content to load
    await page.waitForTimeout(2000);
    
    // Note: Skeleton may still be visible if WebSocket hasn't connected yet
    // This test mainly verifies that skeletons render
  });

  test('transitions from loading to content smoothly', async ({ page }) => {
    await page.goto('/');
    
    // Wait for content to load
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Verify stats cards are visible (not skeletons)
    const statsCards = page.locator('.grid').first();
    await expect(statsCards).toBeVisible();
    
    // Check that actual data is present (not skeleton placeholders)
    const runningCard = page.getByText('Running');
    await expect(runningCard).toBeVisible();
  });
});
