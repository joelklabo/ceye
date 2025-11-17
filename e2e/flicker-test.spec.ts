import { test, expect } from '@playwright/test'

test.describe('Table Flicker Prevention', () => {
  test('should not re-animate rows on WebSocket updates', async ({ page }) => {
    // Start on dashboard
    await page.goto('http://localhost:8080')
    
    // Wait for initial load - look for header
    await page.waitForSelector('h1:has-text("ceye")', { timeout: 10000 })
    
    // Wait for table to appear
    await page.waitForSelector('table tbody tr', { timeout: 10000 })
    
    // Get initial row count
    const initialRows = await page.locator('table tbody tr').count()
    expect(initialRows).toBeGreaterThan(0)
    
    // Get a specific row by its ID to track it
    const firstRow = page.locator('table tbody tr').first()
    
    // Check initial opacity
    await expect(firstRow).toHaveCSS('opacity', '1')
    
    // Wait for potential WebSocket updates (2 seconds)
    await page.waitForTimeout(2000)
    
    // Row should still be fully visible (not re-animating from 0)
    await expect(firstRow).toBeVisible()
    await expect(firstRow).toHaveCSS('opacity', '1')
    
    // Check all visible rows have full opacity (not flickering)
    const allRows = page.locator('table tbody tr')
    const rowCount = await allRows.count()
    for (let i = 0; i < Math.min(rowCount, 5); i++) {
      const row = allRows.nth(i)
      await expect(row).toHaveCSS('opacity', '1')
    }
  })
  
  test('should smoothly update when new runs arrive', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for table
    await page.waitForSelector('table tbody tr', { timeout: 10000 })
    
    // Get initial row count
    const initialCount = await page.locator('table tbody tr').count()
    
    // Wait for potential updates
    await page.waitForTimeout(3000)
    
    // New runs might arrive, but table should not flicker
    const finalCount = await page.locator('table tbody tr').count()
    
    // Table should still be visible
    await expect(page.locator('table')).toBeVisible()
    
    // All rows should be fully visible (opacity 1)
    const rows = page.locator('table tbody tr')
    const count = await rows.count()
    for (let i = 0; i < Math.min(count, 5); i++) {
      await expect(rows.nth(i)).toHaveCSS('opacity', '1')
    }
  })
  
  test('should not flicker during rapid sort changes', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for table
    await page.waitForSelector('table tbody tr', { timeout: 10000 })
    
    // Click sort buttons rapidly
    const providerSort = page.locator('button:has-text("Provider")')
    const statusSort = page.locator('button:has-text("Status")')
    const timeSort = page.locator('button:has-text("Time")')
    
    await providerSort.click()
    await page.waitForTimeout(100)
    await statusSort.click()
    await page.waitForTimeout(100)
    await timeSort.click()
    await page.waitForTimeout(100)
    
    // Table should be stable and visible
    await expect(page.locator('table')).toBeVisible()
    
    // All rows should be fully visible
    const rows = page.locator('table tbody tr')
    const count = await rows.count()
    for (let i = 0; i < Math.min(count, 3); i++) {
      await expect(rows.nth(i)).toHaveCSS('opacity', '1')
    }
  })
})
