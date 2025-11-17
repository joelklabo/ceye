import { test, expect } from '@playwright/test'

test.describe('React App Loading', () => {
  test('app loads and renders root element', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for React root to render
    await page.waitForSelector('#root', { timeout: 10000 })
    
    // Check that root has content (not empty)
    const root = page.locator('#root')
    await expect(root).not.toBeEmpty()
  })
  
  test('displays ceye dashboard title', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for app to load
    await page.waitForTimeout(2000)
    
    // Check for ceye branding
    const title = page.locator('text=/ceye|CI Eye/i').first()
    await expect(title).toBeVisible({ timeout: 10000 })
  })
  
  test('shows connection indicator', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for page load
    await page.waitForTimeout(2000)
    
    // Should show either LIVE or OFFLINE badge
    const liveOrOffline = page.locator('text=/LIVE|OFFLINE/').first()
    await expect(liveOrOffline).toBeVisible({ timeout: 10000 })
  })
})
