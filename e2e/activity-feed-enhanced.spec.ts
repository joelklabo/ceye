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

    // Should have at least one activity item
    const activityItems = page.locator('[class*="hover:bg-muted"]')
    await expect(activityItems.first()).toBeVisible()

    // Should show workflow name
    await expect(activityItems.first()).toContainText(/CI|Build|Test/i)
  })

  test.skip('should show commit SHA in activity items', async ({ page }) => {
    // This will fail until we implement rich data
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    const activityItems = page.locator('[class*="hover:bg-muted"]')
    
    // Should show commit SHA (7 chars)
    await expect(activityItems.first()).toContainText(/[0-9a-f]{7}/)
  })

  test.skip('should show duration in activity items', async ({ page }) => {
    // This will fail until we implement rich data
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    const activityItems = page.locator('[class*="hover:bg-muted"]')
    
    // Should show duration (e.g. "2m 34s" or "45s")
    await expect(activityItems.first()).toContainText(/\d+[ms]/)
  })

  test.skip('should expand activity item to show full details', async ({ page }) => {
    // This will fail until we implement expand/collapse
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    const activityItems = page.locator('[class*="hover:bg-muted"]')
    const firstItem = activityItems.first()

    // Click to expand
    await firstItem.click()

    // Should show expanded details
    const expandedDetails = page.locator('text=/Event type|Conclusion|Actor/i')
    await expect(expandedDetails.first()).toBeVisible()
  })

  test.skip('should show external link to run URL', async ({ page }) => {
    // This will fail until we add external links
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    const activityItems = page.locator('[class*="hover:bg-muted"]')
    
    // Should have external link icon or button
    const externalLink = activityItems.first().locator('a[target="_blank"]')
    await expect(externalLink).toBeVisible()
  })

  test('should show status icons for different run states', async ({ page }) => {
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    // Should show status icons (success checkmark, failure X, etc)
    const statusIcons = page.locator('svg[class*="text-green"]')
    await expect(statusIcons.first()).toBeVisible()
  })

  test('should show timestamp for each activity', async ({ page }) => {
    const activityFeed = page.locator('text=Activity').first()
    await expect(activityFeed).toBeVisible()

    const activityItems = page.locator('[class*="hover:bg-muted"]')
    
    // Should show time (e.g. "2:45:32 PM")
    await expect(activityItems.first()).toContainText(/\d{1,2}:\d{2}:\d{2}\s*(AM|PM)/i)
  })
})
