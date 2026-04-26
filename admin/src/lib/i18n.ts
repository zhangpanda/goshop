const zhCN: Record<string, string> = {
  // 通用
  'save': '保存', 'cancel': '取消', 'confirm': '确认', 'delete': '删除', 'edit': '编辑',
  'add': '新增', 'search': '查询', 'reset': '重置', 'export': '导出', 'import': '导入',
  'detail': '详情', 'status': '状态', 'operation': '操作', 'loading': '加载中...',
  'success': '操作成功', 'failed': '操作失败', 'confirm_delete': '确认删除？',
  'total_items': '共 {total} 条', 'selected_items': '已选 {count} 项',
  'batch_delete': '批量删除', 'batch_enable': '批量启用', 'batch_disable': '批量禁用',
  'no_data': '暂无数据', 'upload': '上传', 'download': '下载',

  // 登录
  'login.title': 'GoShop 管理后台', 'login.username': '用户名', 'login.password': '密码',
  'login.submit': '登录', 'login.captcha': '验证码',

  // 菜单
  'menu.dashboard': '仪表盘', 'menu.system': '系统', 'menu.site': '站点', 'menu.auth': '权限',
  'menu.user': '用户', 'menu.goods': '商品', 'menu.order': '订单', 'menu.website': '网站',
  'menu.brand': '品牌', 'menu.data': '数据', 'menu.article': '文章', 'menu.mobile': '手机',
  'menu.app': '应用', 'menu.warehouse': '仓库', 'menu.tool': '工具',

  // 仪表盘
  'dashboard.order_count': '订单数', 'dashboard.sales': '销售额', 'dashboard.new_users': '新增用户',
  'dashboard.goods_count': '在售商品', 'dashboard.today': '今日', 'dashboard.yesterday': '昨日',
  'dashboard.week': '近7天', 'dashboard.month': '近30天', 'dashboard.pending': '待处理事项',
  'dashboard.trend': '订单走势', 'dashboard.hot_goods': '热销商品', 'dashboard.pay_type': '支付方式',
  'dashboard.new_user_trend': '新增用户走势', 'dashboard.region': '地域分布',
  'dashboard.order_dist': '订单状态分布', 'dashboard.user_top': '用户消费TOP10',
  'dashboard.quick_links': '快捷入口',

  // 商品
  'goods.title': '商品管理', 'goods.add': '新增商品', 'goods.name': '商品名称',
  'goods.category': '分类', 'goods.price': '价格', 'goods.stock': '库存',
  'goods.sales': '销量', 'goods.status': '状态', 'goods.online': '上架', 'goods.offline': '下架',

  // 订单
  'order.title': '订单管理', 'order.no': '订单编号', 'order.amount': '金额',
  'order.status.pending': '待付款', 'order.status.paid': '已付款', 'order.status.shipped': '已发货',
  'order.status.completed': '已完成', 'order.status.cancelled': '已取消',
  'order.ship': '发货', 'order.cancel': '取消', 'order.confirm_pay': '确认收款',
  'order.confirm_receive': '确认收货', 'order.remark': '备注', 'order.print': '打印',

  // 用户
  'user.title': '用户管理', 'user.username': '用户名', 'user.nickname': '昵称',
  'user.phone': '手机', 'user.email': '邮箱', 'user.points': '积分', 'user.balance': '余额',

  // 配置
  'config.title': '系统配置', 'config.save': '保存配置',
  'config.base': '基础配置', 'config.site': '站点设置', 'config.seo': 'SEO配置',
  'config.order': '订单配置', 'config.email': '邮件配置', 'config.sms': '短信配置',
}

let currentLang = zhCN

export function t(key: string, params?: Record<string, string | number>): string {
  let text = currentLang[key] || key
  if (params) {
    Object.entries(params).forEach(([k, v]) => { text = text.replace(`{${k}}`, String(v)) })
  }
  return text
}

export function setLang(lang: Record<string, string>) { currentLang = lang }
export function getLang() { return currentLang }
export default { t, setLang, getLang }
