import { test, expect } from '@playwright/test'
import { loginAdmin } from './helpers'

test.describe('营销模块', () => {
  test.beforeEach(async ({ page }) => {
    await loginAdmin(page)
  })

  test('促销页加载', async ({ page }) => {
    await page.goto('/promotions')
    await expect(page.getByRole('heading', { name: '促销活动' })).toBeVisible()
  })

  test('秒杀页与新建弹窗', async ({ page }) => {
    await page.goto('/seckills')
    await expect(page.getByRole('heading', { name: '秒杀活动' })).toBeVisible()
    await page.getByRole('button', { name: '新建秒杀' }).click()
    await expect(page.getByRole('dialog', { name: '新建秒杀' })).toBeVisible()
    await page.getByRole('dialog', { name: '新建秒杀' }).getByRole('button', { name: /取消|Cancel/ }).click()
    await expect(page.getByRole('dialog', { name: '新建秒杀' })).toBeHidden()
  })

  test('拼团页与新建弹窗', async ({ page }) => {
    await page.goto('/group-buys')
    await expect(page.getByRole('heading', { name: '拼团活动' })).toBeVisible()
    await page.getByRole('button', { name: '新建拼团' }).click()
    await expect(page.getByRole('dialog', { name: '新建拼团' })).toBeVisible()
    await page.getByRole('dialog', { name: '新建拼团' }).getByRole('button', { name: /取消|Cancel/ }).click()
    await expect(page.getByRole('dialog', { name: '新建拼团' })).toBeHidden()
  })

  test('优惠券页加载', async ({ page }) => {
    await page.goto('/coupons')
    await expect(page.getByRole('heading', { name: '优惠券管理' })).toBeVisible()
  })
})
