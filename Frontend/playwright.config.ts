import { defineConfig, devices } from '@playwright/test'
export default defineConfig({
  testDir: './tests/e2e', fullyParallel: true, workers: 2,
  use: { baseURL: 'http://127.0.0.1:5180', trace: 'retain-on-failure', screenshot: 'only-on-failure' },
  projects: [{ name: 'desktop', use: { ...devices['Desktop Chrome'], channel: 'msedge', viewport: { width: 1440, height: 1000 } } }],
  webServer: [
    { command: 'go run ./cmd/mock-api', cwd: '../Backend', url: 'http://127.0.0.1:4010/healthz', reuseExistingServer: !process.env.CI },
    { command: 'npm run dev -- --port 5180', url: 'http://127.0.0.1:5180', reuseExistingServer: !process.env.CI },
  ],
  reporter: [['list']],
})
