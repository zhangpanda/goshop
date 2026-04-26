'use client'
import { useState, ReactNode } from 'react'
import { Layout, Menu, Button, theme, Dropdown } from 'antd'
import {
  DashboardOutlined, SettingOutlined, GlobalOutlined, SafetyOutlined,
  UserOutlined, ShoppingOutlined, OrderedListOutlined, AppstoreOutlined,
  DatabaseOutlined, FileTextOutlined, MobileOutlined, ApiOutlined,
  HomeOutlined, ToolOutlined, MenuFoldOutlined, MenuUnfoldOutlined, LogoutOutlined,
  TagsOutlined, TeamOutlined,
} from '@ant-design/icons'
import { useRouter, usePathname } from 'next/navigation'
import { useAdmin, AdminAuthProvider } from '@/lib/admin-auth'

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '仪表盘' },
  { key: 'system', icon: <SettingOutlined />, label: '系统', children: [
    { key: '/config', label: '系统配置' },
    { key: '/store-info', label: '商店信息' },
    { key: '/locale', label: '语言与货币' },
  ]},
  { key: 'site', icon: <GlobalOutlined />, label: '站点', children: [
    { key: '/site', label: '站点设置' },
    { key: '/sms-setting', label: '短信设置' },
    { key: '/seo', label: 'SEO设置' },
    { key: '/email-setting', label: '邮箱设置' },
    { key: '/agreement', label: '协议管理' },
  ]},
  { key: 'auth', icon: <SafetyOutlined />, label: '权限', children: [
    { key: '/rbac', label: '管理员列表' },
  ]},
  { key: 'user', icon: <UserOutlined />, label: '用户', children: [
    { key: '/users', label: '用户列表' },
    { key: '/user-address', label: '用户地址' },
  ]},
  { key: 'goods', icon: <ShoppingOutlined />, label: '商品', children: [
    { key: '/goods', label: '商品管理' },
    { key: '/categories', label: '商品分类' },
    { key: '/reviews', label: '商品评论' },
    { key: '/goods-browse', label: '商品浏览' },
    { key: '/goods-favor', label: '商品收藏' },
    { key: '/goods-cart', label: '商品购物车' },
    { key: '/goods-params-tpl', label: '商品参数模板' },
    { key: '/goods-spec-tpl', label: '商品规格模板' },
  ]},
  { key: 'order', icon: <OrderedListOutlined />, label: '订单', children: [
    { key: '/orders', label: '订单管理' },
    { key: '/aftersale', label: '订单售后' },
  ]},
  { key: 'website', icon: <AppstoreOutlined />, label: '网站', children: [
    { key: '/navigation', label: '导航管理' },
    { key: '/custom-view', label: '自定义页面' },
    { key: '/slides', label: '首页轮播' },
    { key: '/region', label: '地区管理' },
    { key: '/express', label: '快递管理' },
    { key: '/screening-price', label: '筛选价格' },
    { key: '/links', label: '友情链接' },
    { key: '/theme', label: '主题管理' },
    { key: '/payment', label: '支付方式' },
    { key: '/quick-nav', label: '快捷导航' },
    { key: '/design', label: '页面设计' },
    { key: '/theme-data', label: '主题数据' },
    { key: '/attachment', label: '附件管理' },
    { key: '/attachment-category', label: '附件分类' },
    { key: '/form-input', label: 'Form表单' },
    { key: '/form-data', label: '表单数据' },
    { key: '/layout', label: '首页布局' },
    { key: '/shortcut-menu', label: '快捷菜单' },
  ]},
  { key: 'brand', icon: <TagsOutlined />, label: '品牌', children: [
    { key: '/brands', label: '品牌管理' },
    { key: '/brand-category', label: '品牌分类' },
  ]},
  { key: 'data', icon: <DatabaseOutlined />, label: '数据', children: [
    { key: '/messages', label: '消息管理' },
    { key: '/pay-log', label: '支付日志' },
    { key: '/pay-request-log', label: '支付请求日志' },
    { key: '/integral-log', label: '积分日志' },
    { key: '/refund-log', label: '退款日志' },
    { key: '/sms-log', label: '短信日志' },
    { key: '/email-log', label: '邮件日志' },
    { key: '/error-log', label: '错误日志' },
    { key: '/search-history', label: '搜索记录' },
    { key: '/operation-log', label: '操作日志' },
  ]},
  { key: 'article', icon: <FileTextOutlined />, label: '文章', children: [
    { key: '/articles', label: '文章管理' },
    { key: '/article-category', label: '文章分类' },
  ]},
  { key: 'mobile', icon: <MobileOutlined />, label: '手机', children: [
    { key: '/app-home-nav', label: '首页导航' },
    { key: '/app-config', label: '基础配置' },
    { key: '/app-mini', label: '小程序配置' },
    { key: '/app-center-nav', label: '用户中心导航' },
    { key: '/diy', label: 'DIY装修' },
  ]},
  { key: 'app', icon: <ApiOutlined />, label: '应用', children: [
    { key: '/plugins', label: '应用管理' },
  ]},
  { key: 'warehouse', icon: <HomeOutlined />, label: '仓库', children: [
    { key: '/warehouse', label: '仓库管理' },
    { key: '/warehouse-goods', label: '仓库商品管理' },
  ]},
  { key: 'distribution', icon: <TeamOutlined />, label: '分销', children: [
    { key: '/distribution', label: '分销管理' },
  ]},
  { key: 'tool', icon: <ToolOutlined />, label: '工具', children: [
    { key: '/cache', label: '缓存管理' },
    { key: '/sql-console', label: 'SQL控制台' },
  ]},
]

function Shell({ children }: { children: ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  const { token: { colorBgContainer } } = theme.useToken()
  const router = useRouter()
  const pathname = usePathname()
  const { admin, logout } = useAdmin()

  if (!admin) return null

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Layout.Sider trigger={null} collapsible collapsed={collapsed} theme="dark" width={220}
        breakpoint="lg" collapsedWidth={window?.innerWidth < 768 ? 0 : 80}
        onBreakpoint={broken => setCollapsed(broken)}>
        <div style={{ height: 48, margin: 16, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#fff', fontWeight: 700, fontSize: collapsed ? 16 : 20 }}>
          {collapsed ? 'GS' : 'GoShop'}
        </div>
        <Menu theme="dark" mode="inline" selectedKeys={[pathname]} items={menuItems}
          onClick={({ key }) => { if (key.startsWith('/')) { router.push(key); if (window?.innerWidth < 768) setCollapsed(true) } }} style={{ fontSize: 13 }} />
      </Layout.Sider>
      <Layout>
        <Layout.Header style={{ padding: '0 16px', background: colorBgContainer, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Button type="text" icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />} onClick={() => setCollapsed(!collapsed)} />
          <Dropdown menu={{ items: [{ key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: logout }] }}>
            <Button type="text">{admin.username}</Button>
          </Dropdown>
        </Layout.Header>
        <Layout.Content style={{ margin: '8px 8px 8px 8px', padding: '16px', background: colorBgContainer, borderRadius: 8, minHeight: 280, overflow: 'auto' }}>
          {children}
        </Layout.Content>
      </Layout>
    </Layout>
  )
}

export default function AdminShell({ children }: { children: ReactNode }) {
  return <AdminAuthProvider><Shell>{children}</Shell></AdminAuthProvider>
}
