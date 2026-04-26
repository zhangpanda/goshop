import { defineConfig, devices } from '@playwright/test'

/**
 * 管理后台 E2E：需先启动 Go 后端（8080）且设置 GOSHOP_E2E=1 以跳过验证码。
 * 本地：仓库根目录 `GOSHOP_E2E=1 ./bin/goshop` 或 `go run ./cmd/server`，再 `cd admin && npm run test:e2e`。
 */
export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? [['github'], ['html', { open: 'never' }]] : 'html',
  timeout: 60_000,
  expect: { timeout: 10_000 },
  use: {
    baseURL: 'http://127.0.0.1:3010',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    command: 'npm run dev',
    url: 'http://127.0.0.1:3010/login',
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
})
