import { test, expect } from '@playwright/test'

test.describe('Debug Panel', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:8080')
    await page.waitForTimeout(3000) // Wait for WebSocket connection
  })

  test.skip('should have debug toggle button', async ({ page }) => {
    // Should have a debug toggle button (maybe in header or corner)
    const debugToggle = page.locator('[aria-label*="Debug"]').or(page.locator('text=/Debug|🐛/i'))
    await expect(debugToggle).toBeVisible()
  })

  test.skip('should open debug panel when clicked', async ({ page }) => {
    // Click debug toggle
    const debugToggle = page.locator('[aria-label*="Debug"]').or(page.locator('text=/Debug|🐛/i'))
    await debugToggle.click()

    // Panel should appear
    const debugPanel = page.locator('text=/WebSocket|Logs|Events/i').first()
    await expect(debugPanel).toBeVisible()
  })

  test.skip('should show WebSocket messages in inspector', async ({ page }) => {
    // Open debug panel
    const debugToggle = page.locator('[aria-label*="Debug"]')
    await debugToggle.click()

    // Click WebSocket tab
    await page.locator('text=WebSocket').click()

    // Should show messages received
    const messageList = page.locator('[class*="message-list"]').or(page.locator('text=/runs_update|snapshot/i'))
    await expect(messageList.first()).toBeVisible()
  })

  test.skip('should show WebSocket connection status', async ({ page }) => {
    // Open debug panel
    const debugToggle = page.locator('[aria-label*="Debug"]')
    await debugToggle.click()

    // Should show connection status
    const status = page.locator('text=/Connected|Disconnected|Connecting/i')
    await expect(status).toBeVisible()
  })

  test.skip('should show message timestamps', async ({ page }) => {
    // Open debug panel
    const debugToggle = page.locator('[aria-label*="Debug"]')
    await debugToggle.click()

    await page.locator('text=WebSocket').click()

    // Should show timestamps for each message
    const timestamps = page.locator('text=/\\d{1,2}:\\d{2}:\\d{2}/').first()
    await expect(timestamps).toBeVisible()
  })

  test.skip('should show message payload preview', async ({ page }) => {
    // Open debug panel
    const debugToggle = page.locator('[aria-label*="Debug"]')
    await debugToggle.click()

    await page.locator('text=WebSocket').click()

    // Should show message type and preview
    await expect(page.locator('text=runs_update')).toBeVisible()
  })

  test.skip('should expand message to show full JSON', async ({ page }) => {
    // Open debug panel
    const debugToggle = page.locator('[aria-label*="Debug"]')
    await debugToggle.click()

    await page.locator('text=WebSocket').click()

    // Click on a message to expand
    const firstMessage = page.locator('[class*="debug-message"]').first()
    await firstMessage.click()

    // Should show full JSON
    await expect(page.locator('pre')).toBeVisible()
  })

  test.skip('should clear messages when clear button clicked', async ({ page }) => {
    // Open debug panel
    const debugToggle = page.locator('[aria-label*="Debug"]')
    await debugToggle.click()

    await page.locator('text=WebSocket').click()

    // Wait for messages to arrive
    await page.waitForTimeout(2000)

    // Click clear button
    await page.locator('text=Clear').or(page.locator('button[aria-label="Clear messages"]')).click()

    // Messages should be cleared
    const messages = page.locator('[class*="debug-message"]')
    await expect(messages).toHaveCount(0)
  })

  test.skip('should persist debug panel state across page reloads', async ({ page }) => {
    // Open debug panel
    const debugToggle = page.locator('[aria-label*="Debug"]')
    await debugToggle.click()

    // Reload page
    await page.reload()
    await page.waitForTimeout(3000)

    // Panel should still be open
    const debugPanel = page.locator('text=/WebSocket|Logs|Events/i').first()
    await expect(debugPanel).toBeVisible()
  })

  test.skip('should show event timeline with visual indicators', async ({ page }) => {
    // Open debug panel
    const debugToggle = page.locator('[aria-label*="Debug"]')
    await debugToggle.click()

    // Click Events tab
    await page.locator('text=Events').click()

    // Should show timeline
    const timeline = page.locator('[class*="timeline"]')
    await expect(timeline).toBeVisible()
  })
})
