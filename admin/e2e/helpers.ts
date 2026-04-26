import type { Page } from '@playwright/test'

const DEFAULT_USER = process.env.E2E_ADMIN_USER ?? 'admin'
const DEFAULT_PASS = process.env.E2E_ADMIN_PASS ?? 'admin123'

/**
 * 管理端登录（依赖后端 GOSHOP_E2E=1 时跳过验证码，任意验证码即可）。
 */
export async function loginAdmin(page: Page, user = DEFAULT_USER, pass = DEFAULT_PASS) {
  await page.goto('/login')
  await page.getByPlaceholder('用户名').fill(user)
  await page.getByPlaceholder('密码').fill(pass)
  await page.getByPlaceholder('验证码').fill('000000')
  await page.getByRole('button', { name: '登录' }).click()
  await page.waitForURL(/\/($|\?)/, { timeout: 20_000 })
}
