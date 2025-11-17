import { test, expect } from '@playwright/test'

test.describe('Enhanced Activity Feed', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:8080')
    await page.waitForTimeout(3000) // Wait for initial load and WebSocket
  })

  test('should show activity items with rich details', async ({ page }) => {
    // Wait for activity feed
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    // Should have at least one activity item (use correct selector)
    const activityItems = page.locator('[data-testid="activity-item"]')
    await expect(activityItems.first()).toBeVisible({ timeout: 10000 })

    // Should show workflow name
    await expect(activityItems.first()).toContainText(/•/) // separator exists
  })

  test('should show commit SHA in activity items', async ({ page }) => {
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    const activityItems = page.locator('[data-testid="activity-item"]')
    await expect(activityItems.first()).toBeVisible({ timeout: 10000 })
    
    // Should show commit SHA (7 chars) in monospace font
    const commitSHA = activityItems.first().locator('.font-mono')
    await expect(commitSHA).toBeVisible({ timeout: 5000 })
    const shaText = await commitSHA.textContent()
    expect(shaText).toMatch(/[0-9a-f]{7}/)
  })

  test('should show duration in activity items', async ({ page }) => {
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    const activityItems = page.locator('[data-testid="activity-item"]')
    await expect(activityItems.first()).toBeVisible({ timeout: 10000 })
    
    // Should show duration (e.g. "2m 34s" or "45s")
    await expect(activityItems.first()).toContainText(/Duration:/)
  })

  test('should expand activity item to show full details', async ({ page }) => {
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    const activityItems = page.locator('[data-testid="activity-item"]')
    await expect(activityItems.first()).toBeVisible({ timeout: 10000 })
    const firstItem = activityItems.first()

    // Should show chevron down initially
    await expect(firstItem.locator('svg').last()).toBeVisible()

    // Click to expand
    await firstItem.click()
    await page.waitForTimeout(500) // Wait for animation

    // Should show expanded details (Status, Conclusion, etc.)
    await expect(firstItem).toContainText(/Status:|Conclusion:|Repository:|Branch:/)
  })

  test('should show external link to run URL', async ({ page }) => {
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    const activityItems = page.locator('[data-testid="activity-item"]')
    await expect(activityItems.first()).toBeVisible({ timeout: 10000 })
    
    // Should have external link icon or button
    const externalLink = activityItems.first().locator('[data-testid="activity-external-link"]')
    await expect(externalLink).toBeVisible({ timeout: 5000 })
  })

  test('should show status icons for different run states', async ({ page }) => {
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    const activityItems = page.locator('[data-testid="activity-item"]')
    await expect(activityItems.first()).toBeVisible({ timeout: 10000 })

    // Should show status icons (success checkmark, failure X, etc)
    const icon = activityItems.first().locator('svg').first()
    await expect(icon).toBeVisible()
  })

  test('should show timestamp for each activity', async ({ page }) => {
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    const activityItems = page.locator('[data-testid="activity-item"]')
    await expect(activityItems.first()).toBeVisible({ timeout: 10000 })
    
    // Should show time (e.g. "2:45:32 PM" or "10:45:32 AM")
    const timeLocator = activityItems.first().locator('.text-xs.text-muted-foreground').last()
    await expect(timeLocator).toBeVisible()
    const timeText = await timeLocator.textContent()
    expect(timeText).toMatch(/\d{1,2}:\d{2}:\d{2}\s*(AM|PM)?/i)
  })
})
