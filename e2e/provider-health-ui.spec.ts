import { test, expect } from '@playwright/test'

test.describe('Provider Health UI', () => {
  test.beforeEach(async ({ page }) => {
    // Start with demo mode
    await page.goto('http://localhost:8080')
    await page.waitForTimeout(2000) // Wait for initial load
  })

  test('should display provider cards in full-width layout', async ({ page }) => {
    // Wait for provider health section
    const providerSection = page.locator('[data-testid="provider-health"]')
    await expect(providerSection).toBeVisible()

    // Provider cards should be in a vertical stack, not grid
    const providerCardsList = page.locator('[data-testid="provider-cards-list"]')
    await expect(providerCardsList).toBeVisible()

    // Provider cards list itself should NOT use grid layout (check within the component only)
    const hasGrid = await providerSection.locator('[data-testid="provider-cards-list"] [class*="grid-cols"]').count()
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
    // Wait for provider cards container to appear
    await page.waitForSelector('[data-testid="provider-cards-list"]', { timeout: 10000 })

    // Should show at least one provider name (github, demo, azure, etc.)
    const providerCard = page.locator('[data-testid="provider-cards-list"] > div').first()
    await expect(providerCard).toBeVisible()

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

  test('should have flash animation capability', async ({ page }) => {
    // Verify the provider cards are rendered with motion.div (enables animations)
    await page.waitForSelector('[data-testid="provider-cards-list"]')
    
    // Check that provider cards exist and are visible
    const providerCards = page.locator('[data-testid="provider-cards-list"] > div')
    await expect(providerCards.first()).toBeVisible()
    
    // Verify cards have the necessary classes for border animations
    const firstCard = providerCards.first()
    const hasRoundedBorder = await firstCard.evaluate(el => 
      el.classList.contains('rounded-lg') && el.classList.contains('border')
    )
    expect(hasRoundedBorder).toBe(true)
    
    // Note: Flash animation triggers when LastWebhook.received_at changes
    // In production with real webhooks, border color will flash:
    // hsl(var(--border)) → hsl(var(--primary)) → hsl(var(--border))
  })

  test('should match Activity feed width', async ({ page }) => {
    // Both Provider Health and Activity should have similar widths
    const providerSection = page.locator('[data-testid="provider-health"]')
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
