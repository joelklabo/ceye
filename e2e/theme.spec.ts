import { test, expect } from '@playwright/test';

test.describe('Theme Toggle', () => {
  test('theme toggle button exists', async ({ page }) => {
    await page.goto('/');
    
    // Should have a theme toggle button
    const themeToggle = page.getByRole('button', { name: /theme|dark|light/i });
    await expect(themeToggle).toBeVisible({ timeout: 15000 });
  });

  test('can switch from dark to light mode', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('header', { timeout: 15000 });
    
    // Get initial theme
    const initialIsDark = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    );
    
    // Find theme toggle
    const themeToggle = page.getByRole('button', { name: /theme|dark|light/i });
    await expect(themeToggle).toBeVisible();
    
    // Click to toggle
    await themeToggle.click();
    await page.waitForTimeout(300);
    
    // Get new theme state
    const afterToggleIsDark = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    );
    
    // Should have changed
    expect(afterToggleIsDark).not.toBe(initialIsDark);
  });

  test('theme preference persists across page reloads', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('header', { timeout: 15000 });
    
    // Get initial theme
    const initialTheme = await page.evaluate(() => 
      document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    );
    
    // Toggle theme
    const themeToggle = page.getByRole('button', { name: /theme|dark|light/i });
    await themeToggle.click();
    await page.waitForTimeout(500);
    
    // Get new theme
    const newTheme = await page.evaluate(() =>
      document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    );
    
    // Should be different
    expect(newTheme).not.toBe(initialTheme);
    
    // Reload page
    await page.reload();
    await page.waitForSelector('header', { timeout: 15000 });
    
    // Theme should persist
    const persistedTheme = await page.evaluate(() =>
      document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    );
    
    expect(persistedTheme).toBe(newTheme);
  });

  test('respects system preference on first load', async ({ browser }) => {
    // Create a new context with dark mode preference
    const context = await browser.newContext({
      colorScheme: 'dark'
    });
    
    const page = await context.newPage();
    
    // Go to page (no localStorage yet, should use system preference)
    await page.goto('/');
    await page.waitForSelector('header', { timeout: 15000 });
    
    // Should respect system preference (dark)
    const isDark = await page.evaluate(() => 
      document.documentElement.classList.contains('dark')
    );
    
    expect(isDark).toBe(true);
    
    await context.close();
  });

  test('theme toggle shows correct icon', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('header', { timeout: 15000 });
    
    const themeToggle = page.getByRole('button', { name: /theme|dark|light/i });
    
    // Should have an icon (sun or moon)
    const hasIcon = await themeToggle.locator('svg').count();
    expect(hasIcon).toBeGreaterThan(0);
  });

  test('light mode has correct colors', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('header', { timeout: 15000 });
    
    // Force light mode
    await page.evaluate(() => {
      document.documentElement.classList.remove('dark');
      document.documentElement.classList.add('light');
    });
    
    // Check background is light
    const bgColor = await page.evaluate(() => {
      const body = document.body;
      return window.getComputedStyle(body).backgroundColor;
    });
    
    // Light mode should have a light background (rgb values high)
    const isLightBackground = bgColor.includes('255') || bgColor.includes('white');
    expect(isLightBackground).toBe(true);
  });

  test('dark mode has correct colors', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('header', { timeout: 15000 });
    
    // Force dark mode
    await page.evaluate(() => {
      document.documentElement.classList.remove('light');
      document.documentElement.classList.add('dark');
    });
    
    // Check background is dark
    const bgColor = await page.evaluate(() => {
      const body = document.body;
      return window.getComputedStyle(body).backgroundColor;
    });
    
    // Dark mode should have a dark background (rgb values low)
    const isDarkBackground = 
      bgColor.includes('rgb(0, 0, 0)') || 
      bgColor.includes('rgb(10,') ||
      bgColor.includes('rgb(15,') ||
      bgColor.includes('rgb(20,');
    
    expect(isDarkBackground).toBe(true);
  });
});
