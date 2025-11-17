const { test, expect } = require('@playwright/test');

const BASE_URL = 'http://localhost:8080';

test.describe('Tailwind CSS Integration', () => {
  test('Tailwind utility classes are available and working', async ({ page }) => {
    await page.goto(BASE_URL);
    
    // Inject a test element with Tailwind classes
    await page.evaluate(() => {
      const testDiv = document.createElement('div');
      testDiv.id = 'tailwind-test';
      testDiv.className = 'bg-blue-500 text-white p-4 rounded-lg shadow-md';
      testDiv.textContent = 'Test';
      testDiv.style.position = 'fixed';
      testDiv.style.top = '-9999px'; // Hide off-screen
      document.body.appendChild(testDiv);
    });
    
    // Check if Tailwind classes apply styles
    const styles = await page.evaluate(() => {
      const el = document.getElementById('tailwind-test');
      const computed = window.getComputedStyle(el);
      return {
        backgroundColor: computed.backgroundColor,
        color: computed.color,
        padding: computed.padding,
        borderRadius: computed.borderRadius,
        boxShadow: computed.boxShadow,
      };
    });
    
    // Tailwind bg-blue-500 should result in rgb(59, 130, 246)
    expect(styles.backgroundColor).toMatch(/rgb\(59,\s*130,\s*246\)/);
    
    // text-white should result in white or rgb(255, 255, 255)
    expect(styles.color).toMatch(/rgb\(255,\s*255,\s*255\)/);
    
    // p-4 should result in 16px padding (1rem = 16px)
    expect(styles.padding).toBe('16px');
    
    // rounded-lg should apply border-radius
    expect(styles.borderRadius).toBe('8px');
    
    // shadow-md should apply box-shadow
    expect(styles.boxShadow).not.toBe('none');
    
    console.log('✅ Tailwind CSS is working correctly');
  });
  
  test('Custom design tokens are available', async ({ page }) => {
    await page.goto(BASE_URL);
    
    // Check if CSS custom properties are defined
    const customProps = await page.evaluate(() => {
      const root = document.documentElement;
      const styles = window.getComputedStyle(root);
      return {
        bgPrimary: styles.getPropertyValue('--bg-primary').trim(),
        bgSurface: styles.getPropertyValue('--bg-surface').trim(),
        textPrimary: styles.getPropertyValue('--text-primary').trim(),
        accent: styles.getPropertyValue('--accent').trim(),
        success: styles.getPropertyValue('--success').trim(),
        error: styles.getPropertyValue('--error').trim(),
        warning: styles.getPropertyValue('--warning').trim(),
      };
    });
    
    // Verify our custom colors are defined
    expect(customProps.bgPrimary).toBe('#0f0f23');
    expect(customProps.bgSurface).toBe('#1a1a3e');
    expect(customProps.textPrimary).toBe('#e5e7eb');
    expect(customProps.accent).toBe('#60a5fa');
    expect(customProps.success).toBe('#34d399');
    expect(customProps.error).toBe('#f87171');
    expect(customProps.warning).toBe('#fbbf24');
    
    console.log('✅ Custom design tokens are available');
  });
  
  test('DaisyUI components are available', async ({ page }) => {
    await page.goto(BASE_URL);
    
    // Inject a DaisyUI button and check if it has DaisyUI styles
    const hasDaisyUI = await page.evaluate(() => {
      const btn = document.createElement('button');
      btn.id = 'daisyui-test';
      btn.className = 'btn btn-primary';
      btn.style.position = 'fixed';
      btn.style.top = '-9999px';
      document.body.appendChild(btn);
      
      const computed = window.getComputedStyle(btn);
      
      // DaisyUI buttons have specific styles
      // Check if some DaisyUI-specific properties are present
      const hasMinHeight = computed.minHeight !== 'auto' && computed.minHeight !== '0px';
      const hasDisplay = computed.display !== 'inline';
      
      return hasMinHeight && hasDisplay;
    });
    
    expect(hasDaisyUI).toBe(true);
    console.log('✅ DaisyUI components are available');
  });
});
