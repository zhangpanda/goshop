import { test } from '@playwright/test'
import * as path from 'path'
import { loginAdmin } from './helpers'

/**
 * 生成 README 用截屏到仓库根目录 docs/screenshots/（需先启动 Go 后端且 GOSHOP_E2E=1）。
 * 运行：仓库根目录启动后端后，cd admin && npx playwright test e2e/readme-screenshots.spec.ts
 */
const outDir = path.join(__dirname, '..', '..', 'docs', 'screenshots')

test.describe.serial('README 截屏', () => {
  test('导出管理端登录、仪表盘、商品列表', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 })
    await page.goto('/login', { waitUntil: 'networkidle' })
    await page.getByPlaceholder('用户名').waitFor({ timeout: 30_000 })
    await page.screenshot({ path: path.join(outDir, 'admin-login.png'), fullPage: false })

    await loginAdmin(page)
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.screenshot({ path: path.join(outDir, 'admin-dashboard.png'), fullPage: false })

    await page.goto('/goods')
    await page.waitForLoadState('networkidle')
    await page.screenshot({ path: path.join(outDir, 'admin-goods.png'), fullPage: false })
  })
})
