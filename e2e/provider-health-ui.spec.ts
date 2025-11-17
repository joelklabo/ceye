import { test, expect } from '@playwright/test'

test.describe('Provider Health UI', () => {
  test.beforeEach(async ({ page }) => {
    // Start with demo mode
    await page.goto('http://localhost:8080')
    await page.waitForTimeout(2000) // Wait for initial load
  })

  test('should display provider cards in full-width layout', async ({ page }) => {
    // Wait for provider health section
    const providerSection = page.locator('text=Provider Health').first()
    await expect(providerSection).toBeVisible()

    // Provider cards should be in a vertical stack, not grid
    const providerCards = page.locator('[data-testid="provider-cards"]')
    await expect(providerCards).toBeVisible()

    // Should NOT have grid layout classes
    const hasGrid = await page.locator('[class*="grid-cols"]').count()
    expect(hasGrid).toBe(0)
  })

  test('should show refresh button in header', async ({ page }) => {
    const providerSection = page.locator('text=Provider Health').first()
    await expect(providerSection).toBeVisible()

    // Look for refresh button (RefreshCw icon)
    const refreshButton = page.locator('button').filter({ has: page.locator('svg') }).first()
    await expect(refreshButton).toBeVisible()
  })

  test('should display provider with health indicator', async ({ page }) => {
    // Wait for at least one provider card (demo mode uses "demo" provider)
    await page.waitForSelector('text=demo', { timeout: 10000 })

    // Should show provider name
    const providerName = page.locator('text=demo').first()
    await expect(providerName).toBeVisible()

    // Should show health status (green circle)
    const healthIndicator = page.locator('circle').first()
    await expect(healthIndicator).toBeVisible()

    // Should show "Healthy" or error count
    const healthText = page.locator('text=/Healthy|error/i')
    await expect(healthText.first()).toBeVisible()
  })

  test.skip('should show webhook message count', async ({ page }) => {
    // This will fail until backend implements webhook metadata
    await page.waitForSelector('text=github')

    // Look for message count indicator
    const messageCount = page.locator('text=/\\d+ messages/')
    await expect(messageCount).toBeVisible()
  })

  test.skip('should show last webhook event type', async ({ page }) => {
    // This will fail until backend implements webhook metadata
    await page.waitForSelector('text=github')

    // Look for last webhook info
    const webhookInfo = page.locator('text=/Last webhook:/')
    await expect(webhookInfo).toBeVisible()
  })

  test.skip('should allow viewing webhook payload', async ({ page }) => {
    // This will fail until we implement payload viewing
    await page.waitForSelector('text=github')

    // Look for "View Payload" button
    const viewPayloadBtn = page.locator('text=View Payload')
    await expect(viewPayloadBtn).toBeVisible()

    // Click it
    await viewPayloadBtn.click()

    // Should show JSON payload
    const payload = page.locator('pre').filter({ hasText: /"action"/ })
    await expect(payload).toBeVisible()
  })

  test.skip('should flash when webhook received', async ({ page }) => {
    // This will fail until we implement webhook animations
    await page.waitForSelector('text=github')

    const providerCard = page.locator('text=github').locator('..').locator('..')

    // Watch for border color change (flash effect)
    const initialBorder = await providerCard.evaluate(el => 
      window.getComputedStyle(el).borderColor
    )

    // Wait for webhook (in demo mode, should happen)
    await page.waitForTimeout(5000)

    // Check if border changed (flash animation)
    const afterBorder = await providerCard.evaluate(el => 
      window.getComputedStyle(el).borderColor
    )

    // Border should have changed at some point
    // This is a simplified test - real test would need to watch for transitions
  })

  test('should match Activity feed width', async ({ page }) => {
    // Both Provider Health and Activity should have similar widths
    const providerSection = page.locator('[data-testid="provider-health-section"]')
    const activitySection = page.locator('[data-testid="activity-feed"]')

    await expect(providerSection).toBeVisible()
    await expect(activitySection).toBeVisible()

    const providerWidth = await providerSection.boundingBox()
    const activityWidth = await activitySection.boundingBox()

    // Widths should be similar (within 10%)
    if (providerWidth && activityWidth) {
      const diff = Math.abs(providerWidth.width - activityWidth.width)
      const percentDiff = (diff / activityWidth.width) * 100
      expect(percentDiff).toBeLessThan(10)
    }
  })
})
