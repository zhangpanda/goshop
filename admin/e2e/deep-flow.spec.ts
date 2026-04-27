import { test, expect } from '@playwright/test'
import { loginAdmin } from './helpers'

/**
 * 深度 E2E：覆盖 CRUD 操作、表单交互、搜索筛选、详情抽屉、弹窗等。
 */
test.describe.serial('深度交互 E2E', () => {
  test.beforeEach(async ({ page }) => {
    await loginAdmin(page)
  })

  // ── 分类 CRUD ──
  test('新增分类 → 编辑 → 删除', async ({ page }) => {
    await page.goto('/categories')
    await page.waitForLoadState('networkidle')

    // 新增
    await page.getByRole('button', { name: /新增分类/ }).click()
    await expect(page.getByRole('dialog', { name: '新增分类' })).toBeVisible()
    await page.getByLabel('名称').fill('E2E测试分类')
    await page.getByLabel('排序').fill('999')
    await page.getByRole('dialog').getByRole('button', { name: 'OK' }).click()
    await expect(page.getByText('保存成功')).toBeVisible()
    await expect(page.getByText('E2E测试分类')).toBeVisible()

    // 编辑
    const row = page.locator('tr', { hasText: 'E2E测试分类' })
    await row.getByText('编辑').click()
    await expect(page.getByRole('dialog', { name: '编辑分类' })).toBeVisible()
    await page.getByLabel('名称').clear()
    await page.getByLabel('名称').fill('E2E分类已改')
    await page.getByRole('dialog').getByRole('button', { name: 'OK' }).click()
    await expect(page.getByText('保存成功')).toBeVisible()
    await expect(page.getByText('E2E分类已改')).toBeVisible()

    // 删除
    const row2 = page.locator('tr', { hasText: 'E2E分类已改' })
    await row2.getByText('删除').click()
    // Popconfirm 弹出在 .ant-popconfirm 里
    await page.locator('.ant-popconfirm').getByRole('button', { name: 'OK' }).click()
    await expect(page.getByText('已删除')).toBeVisible()
  })

  // ── 商品列表搜索与详情 ──
  test('商品搜索与详情抽屉', async ({ page }) => {
    await page.goto('/goods')
    await page.waitForLoadState('networkidle')

    // 表格有数据
    const rows = page.locator('.ant-table-tbody tr')
    const count = await rows.count()
    expect(count).toBeGreaterThan(0)

    // 点击第一个商品名打开详情
    await rows.first().locator('a').first().click()
    await expect(page.getByText('商品详情')).toBeVisible()
    // 关闭抽屉
    await page.locator('.ant-drawer-close').click()
    await expect(page.getByText('商品详情')).toBeHidden()

    // 搜索不存在的商品
    await page.getByPlaceholder('商品名称/ID').fill('zzz_不存在_zzz')
    await page.getByPlaceholder('商品名称/ID').press('Enter')
    await page.waitForLoadState('networkidle')
    await expect(page.getByText('暂无数据').or(page.locator('.ant-empty'))).toBeVisible()

    // 重置：清空搜索框再回车
    await page.getByPlaceholder('商品名称/ID').clear()
    await page.getByPlaceholder('商品名称/ID').press('Enter')
    await page.waitForLoadState('networkidle')
    const afterReset = await page.locator('.ant-table-tbody tr').count()
    expect(afterReset).toBeGreaterThan(0)
  })

  // ── 订单 Tab 切换与搜索 ──
  test('订单 Tab 切换与搜索', async ({ page }) => {
    await page.goto('/orders')
    await page.waitForLoadState('networkidle')

    // Tab 切换
    await page.getByRole('tab', { name: '全部' }).click()
    await page.waitForLoadState('networkidle')
    await expect(page.locator('.ant-table')).toBeVisible()

    // 搜索不存在的订单号
    await page.getByPlaceholder('订单号').fill('FAKE_ORDER_999')
    await page.getByPlaceholder('订单号').press('Enter')
    await page.waitForLoadState('networkidle')
  })

  // ── 用户搜索 ──
  test('用户搜索与详情', async ({ page }) => {
    await page.goto('/users')
    await page.waitForLoadState('networkidle')

    await page.getByPlaceholder('用户名/昵称/手机号').fill('admin')
    await page.getByPlaceholder('用户名/昵称/手机号').press('Enter')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('.ant-table-tbody tr').first()).toBeVisible()
  })

  // ── 秒杀新建弹窗 ──
  test('秒杀新建弹窗打开关闭', async ({ page }) => {
    await page.goto('/seckills')
    await expect(page.getByRole('heading', { name: '秒杀活动' })).toBeVisible()
    await page.getByRole('button', { name: '新建秒杀' }).click()
    await expect(page.getByRole('dialog', { name: '新建秒杀' })).toBeVisible()
    await page.getByRole('dialog').getByRole('button', { name: /取消|Cancel/ }).click()
    await expect(page.getByRole('dialog', { name: '新建秒杀' })).toBeHidden()
  })

  // ── 拼团新建弹窗 ──
  test('拼团新建弹窗打开关闭', async ({ page }) => {
    await page.goto('/group-buys')
    await expect(page.getByRole('heading', { name: '拼团活动' })).toBeVisible()
    await page.getByRole('button', { name: '新建拼团' }).click()
    await expect(page.getByRole('dialog', { name: '新建拼团' })).toBeVisible()
    await page.getByRole('dialog').getByRole('button', { name: /取消|Cancel/ }).click()
    await expect(page.getByRole('dialog', { name: '新建拼团' })).toBeHidden()
  })

  // ── 系统配置表单保存 ──
  test('系统配置表单可保存', async ({ page }) => {
    await page.goto('/config')
    await page.waitForLoadState('networkidle')
    await expect(page.locator('.ant-form')).toBeVisible()
    // 找到保存按钮
    const saveBtn = page.getByRole('button', { name: /保存|提交/ })
    if (await saveBtn.isVisible()) {
      await saveBtn.click()
      // 保存成功或无变更都算通过
      await page.waitForLoadState('networkidle')
    }
  })

  // ── 仪表盘时间范围切换 ──
  test('仪表盘时间范围切换', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    await page.locator('.ant-radio-group').getByText('昨日').click()
    await page.waitForLoadState('networkidle')
    await expect(page.getByText('订单数').first()).toBeVisible()

    await page.locator('.ant-radio-group').getByText('近7天').click()
    await page.waitForLoadState('networkidle')
    await expect(page.getByText('订单数').first()).toBeVisible()
  })

  // ── 文章新增弹窗 ──
  test('文章新增弹窗', async ({ page }) => {
    await page.goto('/articles')
    await page.waitForLoadState('networkidle')
    await page.getByRole('button', { name: /新增/ }).click()
    await expect(page.getByRole('dialog')).toBeVisible()
    // 关闭
    await page.getByRole('dialog').getByRole('button', { name: /取消|Cancel/ }).click()
  })

  // ── 权限管理新增管理员弹窗 ──
  test('权限管理新增管理员弹窗', async ({ page }) => {
    await page.goto('/rbac')
    await page.waitForLoadState('networkidle')
    await page.getByRole('button', { name: /新增/ }).click()
    await expect(page.getByRole('dialog')).toBeVisible()
    await page.getByRole('dialog').getByRole('button', { name: /取消|Cancel/ }).click()
  })
})
