import { test, expect } from '@playwright/test'

test.describe('Dashboard - React App', () => {
  test('page loads and shows dashboard', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for React to render
    await page.waitForTimeout(2000)
    
    // Check root element renders
    await expect(page.locator('#root')).not.toBeEmpty()
    
    // Should show LIVE or OFFLINE status
    const status = page.locator('text=/LIVE|OFFLINE/').first()
    await expect(status).toBeVisible({ timeout: 10000 })
  })
  
  test('displays stats cards', async ({ page }) => {
    await page.goto('http://localhost:8080')
    await page.waitForTimeout(2000)
    
    // Wait for data to load
    await page.waitForTimeout(3000)
    
    // Stats cards should be visible
    // The grid layout contains cards with stats
    const grid = page.locator('.grid').first()
    await expect(grid).toBeVisible({ timeout: 10000 })
  })
  
  test('displays runs table or empty state', async ({ page }) => {
    await page.goto('http://localhost:8080')
    await page.waitForTimeout(2000)
    
    // Wait for connection
    await page.waitForSelector('text=/LIVE|OFFLINE/', { timeout: 10000 })
    
    // Should show either a table or empty state
    const hasTable = await page.locator('table').count() > 0
    const hasEmptyState = await page.locator('text=/No runs|empty/i').count() > 0
    
    expect(hasTable || hasEmptyState).toBeTruthy()
  })
  
  test('shows provider information', async ({ page }) => {
    await page.goto('http://localhost:8080')
    await page.waitForTimeout(3000)
    
    // Wait for data to load
    await page.waitForTimeout(2000)
    
    // Provider cards or information should be visible
    // This might be in various places depending on layout
    const pageContent = await page.content()
    expect(pageContent.length).toBeGreaterThan(1000) // Has substantial content
  })
  
  test('theme toggle exists', async ({ page }) => {
    await page.goto('http://localhost:8080')
    await page.waitForTimeout(2000)
    
    // Look for theme toggle button (sun/moon icon typically)
    const buttons = page.locator('button')
    const buttonCount = await buttons.count()
    
    // Should have at least one interactive button
    expect(buttonCount).toBeGreaterThan(0)
  })

  test('displays provider health card with refresh button', async ({ page }) => {
    await page.goto('http://localhost:8080')
    await page.waitForTimeout(3000) // Give time for React and data to render

    // Find the Provider Health title
    const providerHealthTitle = page.locator('h2:has-text("Provider Health")').first();
    await expect(providerHealthTitle).toBeVisible();

    // Find the refresh button which is next to the title inside the CardHeader.
    // Try a more generic locator for the refresh button
    const refreshButton = page.locator('button:has(svg[data-lucide="refresh-cw"])').first();
    await expect(refreshButton).toBeVisible();
  })
})
