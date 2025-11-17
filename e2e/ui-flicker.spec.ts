import { test, expect } from '@playwright/test'

const URL = 'http://localhost:8080'

test.describe('UI Flicker Detection', () => {
  test('should not flicker on WebSocket updates', async ({ page }) => {
    await page.goto(URL)
    await page.waitForTimeout(3000) // Wait for initial load
    
    // Wait for runs to appear
    const runsTable = page.locator('table tbody tr').first()
    await expect(runsTable).toBeVisible({ timeout: 15000 })
    
    // Capture initial opacity of first 5 rows
    const rows = page.locator('table tbody tr')
    const rowCount = Math.min(await rows.count(), 5)
    
    const initialOpacities = []
    for (let i = 0; i < rowCount; i++) {
      const opacity = await rows.nth(i).evaluate(el => window.getComputedStyle(el).opacity)
      initialOpacities.push(opacity)
    }
    
    // All rows should be fully visible (opacity = 1)
    for (let i = 0; i < rowCount; i++) {
      expect(parseFloat(initialOpacities[i])).toBe(1)
    }
    
    // Wait for WebSocket update (demo mode updates every ~15s)
    await page.waitForTimeout(16000)
    
    // Check opacities again - should still be 1 (no re-animation)
    for (let i = 0; i < rowCount; i++) {
      const opacity = await rows.nth(i).evaluate(el => window.getComputedStyle(el).opacity)
      expect(parseFloat(opacity)).toBe(1)
    }
  })

  test('StatsCards should not re-animate on updates', async ({ page }) => {
    await page.goto(URL)
    await page.waitForTimeout(3000)
    
    // Find stat cards
    const statCards = page.locator('div').filter({ hasText: /^Running|^Queued|^Success|^Failed/ })
    await expect(statCards.first()).toBeVisible({ timeout: 10000 })
    
    // Capture initial state
    const cardCount = Math.min(await statCards.count(), 4)
    const initialOpacities = []
    for (let i = 0; i < cardCount; i++) {
      const opacity = await statCards.nth(i).evaluate(el => window.getComputedStyle(el).opacity)
      initialOpacities.push(opacity)
    }
    
    // All cards should be visible
    for (let i = 0; i < cardCount; i++) {
      expect(parseFloat(initialOpacities[i])).toBeGreaterThanOrEqual(0.9)
    }
    
    // Wait for update
    await page.waitForTimeout(16000)
    
    // Cards should still be visible (not re-animating from 0)
    for (let i = 0; i < cardCount; i++) {
      const opacity = await statCards.nth(i).evaluate(el => window.getComputedStyle(el).opacity)
      expect(parseFloat(opacity)).toBeGreaterThanOrEqual(0.9)
    }
  })

  test('ActivityFeed should not flicker on new items', async ({ page }) => {
    await page.goto(URL)
    await page.waitForTimeout(3000)
    
    // Wait for activity items
    const activityItems = page.locator('[data-testid="activity-item"]')
    await expect(activityItems.first()).toBeVisible({ timeout: 15000 })
    
    // Get initial count
    const initialCount = await activityItems.count()
    
    // Capture opacity of existing items
    const itemCount = Math.min(initialCount, 3)
    const initialOpacities = []
    for (let i = 0; i < itemCount; i++) {
      const opacity = await activityItems.nth(i).evaluate(el => window.getComputedStyle(el).opacity)
      initialOpacities.push(opacity)
    }
    
    // All should be visible
    for (let i = 0; i < itemCount; i++) {
      expect(parseFloat(initialOpacities[i])).toBe(1)
    }
    
    // Wait for new activity
    await page.waitForTimeout(16000)
    
    // Existing items should not have re-animated
    // (they might have moved position but opacity should stay 1)
    for (let i = 0; i < itemCount; i++) {
      const opacity = await activityItems.nth(i).evaluate(el => window.getComputedStyle(el).opacity)
      expect(parseFloat(opacity)).toBe(1)
    }
  })

  test('should have smooth animations (no jank)', async ({ page }) => {
    await page.goto(URL)
    await page.waitForTimeout(3000)
    
    // Start performance monitoring
    await page.evaluate(() => {
      // @ts-ignore
      window.frameDrops = 0
      // @ts-ignore
      window.lastFrameTime = performance.now()
      
      function checkFrame() {
        const now = performance.now()
        // @ts-ignore
        const delta = now - window.lastFrameTime
        
        // If frame took longer than 33ms (below 30fps), count as dropped
        if (delta > 33) {
          // @ts-ignore
          window.frameDrops++
        }
        
        // @ts-ignore
        window.lastFrameTime = now
        requestAnimationFrame(checkFrame)
      }
      
      requestAnimationFrame(checkFrame)
    })
    
    // Wait for several WebSocket updates
    await page.waitForTimeout(20000)
    
    // Check frame drops
    const frameDrops = await page.evaluate(() => {
      // @ts-ignore
      return window.frameDrops
    })
    
    // Allow some frame drops but not excessive
    // In 20 seconds at 60fps = 1200 frames
    // Allow up to 5% drops = 60 frames
    expect(frameDrops).toBeLessThan(60)
  })
})
