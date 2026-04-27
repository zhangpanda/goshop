import type { Page } from '@playwright/test'

const DEFAULT_USER = process.env.E2E_ADMIN_USER ?? 'admin'
const DEFAULT_PASS = process.env.E2E_ADMIN_PASS ?? 'admin123'

/**
 * 管理端登录（依赖后端 GOSHOP_E2E=1 时跳过验证码，任意验证码即可）。
 * 带重试：偶发登录失败时自动重试一次。
 */
export async function loginAdmin(page: Page, user = DEFAULT_USER, pass = DEFAULT_PASS) {
  for (let attempt = 0; attempt < 2; attempt++) {
    await page.goto('/login', { waitUntil: 'networkidle' })
    await page.getByPlaceholder('用户名').waitFor({ timeout: 30_000 })
    await page.getByPlaceholder('用户名').fill(user)
    await page.getByPlaceholder('密码').fill(pass)
    await page.getByPlaceholder('验证码').fill('000000')
    await page.getByRole('button', { name: /登.*录/ }).click()
    try {
      await page.waitForURL(/\/($|\?)/, { timeout: 15_000 })
      return
    } catch {
      if (attempt === 1) throw new Error('Login failed after 2 attempts')
    }
  }
}
