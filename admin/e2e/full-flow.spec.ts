import { test, expect } from '@playwright/test'
import { loginAdmin } from './helpers'

test.describe.serial('管理后台全流程 E2E', () => {
  test.beforeEach(async ({ page }) => {
    await loginAdmin(page)
  })

  test('仪表盘加载统计卡片', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await expect(page.getByText('订单数').first()).toBeVisible()
    await expect(page.getByText('新增用户').first()).toBeVisible()
    await expect(page.getByText('在售商品').first()).toBeVisible()
  })

  test('商品分类页加载', async ({ page }) => {
    await page.goto('/categories')
    await expect(page.getByRole('heading', { name: '分类管理' })).toBeVisible()
  })

  test('商品列表加载', async ({ page }) => {
    await page.goto('/goods')
    await expect(page.getByRole('heading', { name: '商品管理' })).toBeVisible()
    await expect(page.locator('.ant-table')).toBeVisible()
  })

  test('订单列表加载', async ({ page }) => {
    await page.goto('/orders')
    await expect(page.getByRole('heading', { name: '订单管理' })).toBeVisible()
  })

  test('用户列表加载', async ({ page }) => {
    await page.goto('/users')
    await expect(page.getByRole('heading', { name: '用户管理' })).toBeVisible()
  })

  test('文章列表加载', async ({ page }) => {
    await page.goto('/articles')
    await expect(page.getByRole('heading', { name: '文章管理' })).toBeVisible()
  })

  test('优惠券页加载', async ({ page }) => {
    await page.goto('/coupons')
    await expect(page.getByRole('heading', { name: '优惠券管理' })).toBeVisible()
  })

  test('促销页加载', async ({ page }) => {
    await page.goto('/promotions')
    await expect(page.getByRole('heading', { name: '促销活动' })).toBeVisible()
  })

  test('售后列表加载', async ({ page }) => {
    await page.goto('/aftersale')
    await expect(page.getByRole('heading', { name: '订单售后' })).toBeVisible()
  })

  test('系统配置页加载', async ({ page }) => {
    await page.goto('/config')
    await page.waitForLoadState('networkidle')
    await expect(page.getByRole('heading', { name: '系统配置' })).toBeVisible()
  })

  test('站点设置页加载', async ({ page }) => {
    await page.goto('/site')
    await page.waitForLoadState('networkidle')
    await expect(page.getByRole('heading', { name: '站点设置' })).toBeVisible()
  })

  test('权限管理加载', async ({ page }) => {
    await page.goto('/rbac')
    await expect(page.getByRole('heading', { name: '权限管理' })).toBeVisible()
  })

  test('导航管理加载', async ({ page }) => {
    await page.goto('/navigation')
    await page.waitForLoadState('networkidle')
    await expect(page.getByText('导航').first()).toBeVisible()
  })

  test('支付方式加载', async ({ page }) => {
    await page.goto('/payment')
    await expect(page.getByRole('heading', { name: '支付方式' })).toBeVisible()
  })

  test('操作日志加载', async ({ page }) => {
    await page.goto('/operation-log')
    await expect(page.getByRole('heading', { name: '操作审计日志' })).toBeVisible()
  })

  test('分销管理加载', async ({ page }) => {
    await page.goto('/distribution')
    await expect(page.getByRole('heading', { name: '分销管理' })).toBeVisible()
  })

  test('缓存管理加载', async ({ page }) => {
    await page.goto('/cache')
    await expect(page.getByRole('heading', { name: '缓存管理' })).toBeVisible()
  })

  test('侧边栏导航到商品页', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.locator('.ant-menu').getByText('商品', { exact: true }).first().click()
    await page.locator('.ant-menu').getByText('商品管理').click()
    await page.waitForURL(/\/goods/)
    await expect(page.getByRole('heading', { name: '商品管理' })).toBeVisible()
  })

  test('退出登录回到登录页', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await page.locator('header').getByText('admin').click()
    await page.getByText('退出登录').click()
    await page.waitForURL(/\/login/)
    await expect(page.getByRole('button', { name: /登.*录/ })).toBeVisible()
  })
})
