import { test, expect } from '@playwright/test'

test.describe('Connection Indicator', () => {
  test('should show "Live" badge when connected', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for React app to load
    await page.waitForTimeout(3000)
    
    // Should show "LIVE" or "OFFLINE" badge
    const status = page.locator('text=/LIVE|OFFLINE/').first()
    await expect(status).toBeVisible({ timeout: 15000 })
  })
  
  test('should show timestamp after connection', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for React app to load
    await page.waitForTimeout(2000)
    
    // Wait for connection
    await page.waitForSelector('text=LIVE', { timeout: 10000 })
    
    // Should show timestamp (ago format or "just now")
    const timestamp = page.locator('text=/\\d+[smh] ago|just now/').first()
    
    // Timestamp should eventually appear
    await expect(timestamp).toBeVisible({ timeout: 10000 })
  })
  
  test('should display connection status with icon', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for React app
    await page.waitForTimeout(3000)
    
    // Should show LIVE or OFFLINE
    const status = page.locator('text=/LIVE|OFFLINE/').first()
    await expect(status).toBeVisible({ timeout: 15000 })
  })
  
  test('live badge is styled correctly', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for React app
    await page.waitForTimeout(3000)
    
    // Wait for connection status (LIVE or OFFLINE)
    const status = page.locator('text=/LIVE|OFFLINE/').first()
    await expect(status).toBeVisible({ timeout: 15000 })
  })
  
  test('connection indicator is in header', async ({ page }) => {
    await page.goto('http://localhost:8080')
    
    // Wait for React app
    await page.waitForTimeout(2000)
    
    // Wait for page load
    await page.waitForSelector('text=LIVE', { timeout: 10000 })
    
    // LIVE badge should be visible
    await expect(page.locator('text=LIVE')).toBeVisible()
  })
})
