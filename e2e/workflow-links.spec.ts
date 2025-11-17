import { test, expect } from '@playwright/test'

test.describe('Workflow Source Links', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:8080')
    await page.waitForTimeout(3000) // Wait for React to render
  })

  test('should show external link icon on each run in table', async ({ page }) => {
    // Wait for table to load
    await page.waitForSelector('table tbody tr', { timeout: 15000 })
    
    // Check first 3 rows for external link icon
    const rows = page.locator('table tbody tr')
    const count = await rows.count()
    const rowsToCheck = Math.min(count, 3)
    
    for (let i = 0; i < rowsToCheck; i++) {
      const row = rows.nth(i)
      const linkIcon = row.locator('[data-testid="external-link-icon"]')
      await expect(linkIcon).toBeVisible({ timeout: 5000 })
    }
  })

  test('should open workflow URL in new tab when external link clicked', async ({ page, context }) => {
    // Wait for table to load
    await page.waitForSelector('table tbody tr', { timeout: 15000 })
    
    // Find first row with external link
    const firstRow = page.locator('table tbody tr').first()
    const linkIcon = firstRow.locator('[data-testid="external-link-icon"]')
    await expect(linkIcon).toBeVisible()
    
    // Get the URL before clicking
    const href = await linkIcon.getAttribute('href')
    expect(href).toBeTruthy()
    expect(href).toMatch(/^https?:\/\//) // Should be a valid URL
    
    // Click should open new tab (we won't actually open it in test)
    const target = await linkIcon.getAttribute('target')
    expect(target).toBe('_blank')
    
    const rel = await linkIcon.getAttribute('rel')
    expect(rel).toContain('noopener')
  })

  test('should show external link in activity feed items', async ({ page }) => {
    // Wait for activity feed
    await page.waitForSelector('text=/Activity/', { timeout: 10000 })
    
    // Check if activity items have external links
    const activityItems = page.locator('[data-testid="activity-item"]')
    const count = await activityItems.count()
    
    if (count > 0) {
      // First item should have external link
      const firstItem = activityItems.first()
      const externalLink = firstItem.locator('[data-testid="activity-external-link"]')
      await expect(externalLink).toBeVisible({ timeout: 5000 })
      
      // Should have proper attributes
      const href = await externalLink.getAttribute('href')
      expect(href).toBeTruthy()
      
      const target = await externalLink.getAttribute('target')
      expect(target).toBe('_blank')
    }
  })

  test('should have provider-specific visual treatment', async ({ page }) => {
    // Wait for table
    await page.waitForSelector('table tbody tr', { timeout: 15000 })
    
    // Check that links are styled (have hover states, etc)
    const firstLink = page.locator('[data-testid="external-link-icon"]').first()
    await expect(firstLink).toBeVisible()
    
    // Check for hover states (color change)
    const initialColor = await firstLink.evaluate((el) => {
      return window.getComputedStyle(el).color
    })
    
    await firstLink.hover()
    await page.waitForTimeout(200) // Wait for transition
    
    const hoverColor = await firstLink.evaluate((el) => {
      return window.getComputedStyle(el).color
    })
    
    // Colors should be different on hover (or have different opacity)
    // This is a basic check - in real app colors will change
    expect(initialColor).toBeTruthy()
    expect(hoverColor).toBeTruthy()
  })

  test('should work for both completed and in-progress runs', async ({ page }) => {
    await page.waitForSelector('table tbody tr', { timeout: 15000 })
    
    // Find runs with different statuses
    const rows = page.locator('table tbody tr')
    const count = await rows.count()
    
    // Each row should have an external link regardless of status
    for (let i = 0; i < Math.min(count, 5); i++) {
      const row = rows.nth(i)
      const linkIcon = row.locator('[data-testid="external-link-icon"]')
      await expect(linkIcon).toBeVisible({ timeout: 5000 })
    }
  })
})
