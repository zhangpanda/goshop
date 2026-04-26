'use client'
import { useEffect, useState, useCallback, Suspense } from 'react'
import { Card, Col, Row, Statistic, Typography, Radio, Badge, Space, Button, Table, Tag } from 'antd'
import { ShoppingOutlined, OrderedListOutlined, UserOutlined, DollarOutlined, AlertOutlined } from '@ant-design/icons'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'
import dynamic from 'next/dynamic'

const Line = dynamic(() => import('@ant-design/charts').then(m => m.Line), { ssr: false })
const Pie = dynamic(() => import('@ant-design/charts').then(m => m.Pie), { ssr: false })
const Column = dynamic(() => import('@ant-design/charts').then(m => m.Column), { ssr: false })


interface Stat { order_count: number; sales: number; user_count: number; goods_count: number }
interface Stats {
  today: Stat; yesterday: Stat; week: Stat; month: Stat
  trend: { date: string; sales: number; count: number }[]
  goods_top: { goods_id: number; title: string; sales: number }[]
  user_top: { user_id: number; nickname: string; amount: number }[]
  order_dist: { status: number; count: number }[]
  order_pending_count: number; aftersale_pending_count: number; goods_offline_count: number; review_pending_count: number
  pay_type_stats: { client_type: string; count: number; amount: number }[]
  region_stats: { province: string; count: number }[]
  new_user_trend: { date: string; count: number }[]
}

const SM: Record<number, string> = { 0: '待付款', 1: '已付款', 2: '已发货', 3: '已完成', 4: '已取消', 5: '已关闭', 6: '退款中' }

// 每个图表独立的时间范围组件
function ChartDateRange({ onChange }: { onChange: (days: number) => void }) {
  const [active, setActive] = useState(30)
  const click = (d: number) => { setActive(d); onChange(d) }
  return (
    <Space size={4}>
      {[{ d: 7, l: '7天' }, { d: 15, l: '15天' }, { d: 30, l: '30天' }].map(({ d, l }) => (
        <Button key={d} size="small" type={active === d ? 'primary' : 'default'} onClick={() => click(d)}>{l}</Button>
      ))}
    </Space>
  )
}

export default function DashboardPage() {
  const [data, setData] = useState<Stats | null>(null)
  const [range, setRange] = useState<'today' | 'yesterday' | 'week' | 'month'>('today')
  const [salesDays, setSalesDays] = useState(30)
  const [orderDays, setOrderDays] = useState(30)
  const [userDays, setUserDays] = useState(30)
  const router = useRouter()

  const loadData = useCallback(async (days: number) => {
    const d = await api.get<Stats>(`/admin/statistical?days=${days}`)
    setData(d)
  }, [])

  // Fetch with the max of all chart day ranges
  const maxDays = Math.max(salesDays, orderDays, userDays)
  useEffect(() => { loadData(maxDays) }, [maxDays, loadData])

  if (!data) return null

  const cur = data[range]
  const prev = range === 'today' ? data.yesterday : range === 'yesterday' ? data.today : range === 'week' ? data.month : data.week

  const statCards = [
    { title: '订单数', value: cur.order_count, prev: prev.order_count, icon: <OrderedListOutlined />, color: '#faad14', link: '/orders' },
    { title: '销售额(元)', value: cur.sales / 100, prev: prev.sales / 100, prefix: '¥', icon: <DollarOutlined />, color: '#f5222d', link: '/orders' },
    { title: '新增用户', value: cur.user_count, prev: prev.user_count, icon: <UserOutlined />, color: '#1890ff', link: '/users' },
    { title: '在售商品', value: cur.goods_count, prev: prev.goods_count, icon: <ShoppingOutlined />, color: '#52c41a', link: '/goods' },
  ]

  const pendingItems = [
    { label: '待付款订单', count: data.order_pending_count, link: '/orders' },
    { label: '待处理售后', count: data.aftersale_pending_count, link: '/aftersale' },
    { label: '下架商品', count: data.goods_offline_count, link: '/goods' },
    { label: '待回复评价', count: data.review_pending_count, link: '/reviews' },
  ]

  const salesTrend = (data.trend || []).slice(-salesDays).map(t => ({ date: t.date, value: t.sales / 100, type: '销售额(元)' }))
  const orderTrend = (data.trend || []).slice(-orderDays).map(t => ({ date: t.date, value: t.count, type: '订单数' }))
  const payData = (data.pay_type_stats || []).map(p => ({ type: p.client_type || '未知', value: p.count }))
  const hotData = (data.goods_top || []).slice(0, 10).map(g => ({ name: g.title.length > 8 ? g.title.slice(0, 8) + '..' : g.title, sales: g.sales }))
  const newUserData = (data.new_user_trend || []).slice(-userDays).map(t => ({ date: t.date, value: t.count, type: '新增用户' }))

  return (
    <>
      <Typography.Title level={4}>仪表盘</Typography.Title>

      {/* 时间范围切换 - 影响顶部统计卡片 */}
      <Radio.Group value={range} onChange={e => setRange(e.target.value)} style={{ marginBottom: 16 }}>
        <Radio.Button value="today">今日</Radio.Button>
        <Radio.Button value="yesterday">昨日</Radio.Button>
        <Radio.Button value="week">近7天</Radio.Button>
        <Radio.Button value="month">近30天</Radio.Button>
      </Radio.Group>

      {/* 统计卡片 - 带环比 */}
      <Row gutter={[16, 16]}>
        {statCards.map(s => (
          <Col span={6} key={s.title}>
            <Card hoverable onClick={() => router.push(s.link)} size="small">
              <Statistic title={s.title} value={s.value} prefix={s.prefix || s.icon} valueStyle={{ color: s.color }} />
              <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
                对比: {s.prev} {s.value > s.prev ? '↑' : s.value < s.prev ? '↓' : '—'}
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {/* 待处理事项 */}
      <Card title={<><AlertOutlined /> 待处理事项</>} size="small" style={{ marginTop: 16 }}>
        <Row gutter={16}>
          {pendingItems.map(p => (
            <Col span={6} key={p.label}>
              <Card size="small" hoverable onClick={() => router.push(p.link)}>
                <Badge count={p.count} overflowCount={999}><span style={{ paddingRight: 16 }}>{p.label}</span></Badge>
              </Card>
            </Col>
          ))}
        </Row>
      </Card>

      {/* 订单成交金额走势 - 独立时间范围 */}
      <Card title="订单成交金额走势" size="small" style={{ marginTop: 16 }}
        extra={<ChartDateRange onChange={setSalesDays} />}>
        <Suspense fallback={<div style={{ height: 300 }} />}>
          <Line data={salesTrend} xField="date" yField="value" colorField="type" height={300} style={{ lineWidth: 2 }} />
        </Suspense>
      </Card>

      {/* 订单交易走势 - 独立时间范围 */}
      <Card title="订单交易走势" size="small" style={{ marginTop: 16 }}
        extra={<ChartDateRange onChange={setOrderDays} />}>
        <Suspense fallback={<div style={{ height: 300 }} />}>
          <Line data={orderTrend} xField="date" yField="value" colorField="type" height={300} style={{ lineWidth: 2 }} />
        </Suspense>
      </Card>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="热销商品TOP10" size="small">
            <Suspense fallback={<div style={{ height: 250 }} />}>
              {hotData.length ? <Column data={hotData} xField="name" yField="sales" height={250} /> : <Typography.Text type="secondary">暂无数据</Typography.Text>}
            </Suspense>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="支付方式统计" size="small">
            <Suspense fallback={<div style={{ height: 250 }} />}>
              {payData.length ? <Pie data={payData} angleField="value" colorField="type" height={250} innerRadius={0.5} legend={{ position: 'bottom' }} /> : <Typography.Text type="secondary">暂无数据</Typography.Text>}
            </Suspense>
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="新增用户走势" size="small" extra={<ChartDateRange onChange={setUserDays} />}>
            <Suspense fallback={<div style={{ height: 250 }} />}>
              <Line data={newUserData} xField="date" yField="value" colorField="type" height={250} style={{ lineWidth: 2 }} />
            </Suspense>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="订单地域分布TOP10" size="small">
            <Table dataSource={data.region_stats || []} rowKey="province" pagination={false} size="small"
              columns={[{ title: '省份', dataIndex: 'province' }, { title: '订单数', dataIndex: 'count' }]} />
            {!(data.region_stats?.length) && <Typography.Text type="secondary">暂无数据</Typography.Text>}
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="订单状态分布" size="small">
            <Table dataSource={data.order_dist || []} rowKey="status" pagination={false} size="small"
              columns={[
                { title: '状态', dataIndex: 'status', render: (v: number) => <Tag>{SM[v] || v}</Tag> },
                { title: '数量', dataIndex: 'count' },
              ]} />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="用户消费TOP10" size="small">
            <Table dataSource={data.user_top || []} rowKey="user_id" pagination={false} size="small"
              columns={[
                { title: '用户ID', dataIndex: 'user_id', width: 80 },
                { title: '昵称', dataIndex: 'nickname' },
                { title: '消费额', dataIndex: 'amount', render: (v: number) => `¥${(v / 100).toFixed(2)}` },
              ]} />
          </Card>
        </Col>
      </Row>

      <Card title="快捷入口" size="small" style={{ marginTop: 16 }}>
        <Space wrap>
          {[
            { label: '商品管理', link: '/goods' }, { label: '订单管理', link: '/orders' },
            { label: '用户管理', link: '/users' }, { label: '系统配置', link: '/config' },
            { label: '文章管理', link: '/articles' }, { label: '售后管理', link: '/aftersale' },
            { label: '缓存管理', link: '/cache' }, { label: '附件管理', link: '/attachment' },
            { label: 'DIY装修', link: '/diy' }, { label: '表单设计', link: '/form-input' },
          ].map(q => <Button key={q.link} onClick={() => router.push(q.link)}>{q.label}</Button>)}
        </Space>
      </Card>
    </>
  )
}
