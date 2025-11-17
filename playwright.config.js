const { defineConfig } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './e2e',
  timeout: 30000,
  use: {
    baseURL: 'http://localhost:8080',
    screenshot: 'only-on-failure',
    video: 'on',
    trace: 'retain-on-failure',
  },
  webServer: {
    command: './bin/ceye --web',
    port: 8080,
    timeout: 30000,
    reuseExistingServer: true,
  },
  reporter: [
    ['html', { outputFolder: 'tmp/playwright-report' }],
    ['list'],
  ],
});
