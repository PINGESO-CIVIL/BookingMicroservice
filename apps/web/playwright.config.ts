import { defineConfig, devices } from '@playwright/test'
export default defineConfig({
  testDir: './tests/e2e', outputDir: '../../reports/playwright', fullyParallel: true, workers: 2,
  use: { baseURL: 'http://127.0.0.1:5180', trace: 'retain-on-failure', screenshot: 'only-on-failure' },
  projects: [{ name: 'desktop', use: { ...devices['Desktop Chrome'], channel: process.env.PLAYWRIGHT_CHANNEL, viewport: { width: 1440, height: 1000 } } }],
  webServer: [
    { command: 'go run ./cmd/api', cwd: '../api', url: 'http://127.0.0.1:4010/healthz', reuseExistingServer: !process.env.CI },
    { command: 'pnpm dev --port 5180', env: { VITE_MOCK_API: 'false' }, url: 'http://127.0.0.1:5180', reuseExistingServer: !process.env.CI },
  ],
  reporter: [['list']],
})
