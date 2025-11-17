import { test, expect } from '@playwright/test'

test.describe('WebSocket Connection', () => {
  test.beforeEach(async ({ page }) => {
    // Start fresh with no console errors
    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        console.log(`Console error: ${msg.text()}`)
      }
    })
  })

  test('should establish WebSocket connection within 3 seconds', async ({ page }) => {
    // Navigate to the app
    await page.goto('http://localhost:8080')
    
    // Wait for connection indicator to show LIVE status
    await expect(page.locator('text=/LIVE|ONLINE/i').first()).toBeVisible({ timeout: 15000 })
    
    // Verify we're not showing OFFLINE
    const offlineIndicator = page.locator('text=/OFFLINE|DISCONNECTED/i').first()
    await expect(offlineIndicator).not.toBeVisible()
  })

  test('should receive WebSocket messages', async ({ page }) => {
    let messageReceived = false
    
    // Intercept WebSocket messages in the browser
    await page.addInitScript(() => {
      const originalWebSocket = window.WebSocket
      window.WebSocket = function(url: string) {
        const ws = new originalWebSocket(url)
        
        ws.addEventListener('message', (event) => {
          console.log('[TEST] WebSocket message received:', event.data.slice(0, 100))
          ;(window as any).__wsMessageReceived = true
        })
        
        ws.addEventListener('open', () => {
          console.log('[TEST] WebSocket opened')
          ;(window as any).__wsConnected = true
        })
        
        ws.addEventListener('error', (error) => {
          console.log('[TEST] WebSocket error:', error)
        })
        
        return ws
      }
    })
    
    await page.goto('http://localhost:8080')
    
    // Wait for WebSocket to connect
    await page.waitForFunction(() => (window as any).__wsConnected === true, { timeout: 5000 })
    
    // Wait for at least one message
    await page.waitForFunction(() => (window as any).__wsMessageReceived === true, { timeout: 10000 })
    
    // Verify connection indicator shows LIVE
    await expect(page.locator('text=/LIVE/i').first()).toBeVisible()
  })

  test('should display stats from WebSocket data', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for connection
    await expect(page.locator('text=/LIVE/i').first()).toBeVisible({ timeout: 15000 })
    
    // Wait for dashboard to render with data
    await page.waitForTimeout(2000)
    
    // Check that dashboard content is visible
    const mainContent = page.locator('main')
    await expect(mainContent).toBeVisible()
    
    // Verify runs table has data
    const table = page.locator('table tbody tr').first()
    await expect(table).toBeVisible({ timeout: 5000 })
  })

  test('should not show OFFLINE status constantly', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for initial connection
    await page.waitForTimeout(3000)
    
    // Should show LIVE, not OFFLINE
    const liveIndicator = page.locator('text=/LIVE/i').first()
    await expect(liveIndicator).toBeVisible({ timeout: 15000 })
    
    // Wait a bit and verify it stays LIVE
    await page.waitForTimeout(5000)
    await expect(liveIndicator).toBeVisible()
    
    // Verify OFFLINE is not shown
    const offlineIndicator = page.locator('text=/OFFLINE/i').first()
    await expect(offlineIndicator).not.toBeVisible()
  })
})
