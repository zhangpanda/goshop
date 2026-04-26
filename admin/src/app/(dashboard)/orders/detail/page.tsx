'use client'
import { useEffect, useState } from 'react'
import { Card, Descriptions, Table, Tag, Button, Space, Typography, Modal, Form, Input, message, Timeline, Divider } from 'antd'
import { useRouter, useSearchParams } from 'next/navigation'
import { PrinterOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'

const SM: Record<number, { t: string; c: string }> = { 0: { t: '待付款', c: 'default' }, 1: { t: '已付款', c: 'blue' }, 2: { t: '已发货', c: 'cyan' }, 3: { t: '已完成', c: 'green' }, 4: { t: '已取消', c: 'orange' }, 5: { t: '已关闭', c: '' }, 6: { t: '退款中', c: 'red' } }
type OI = { id: number; title: string; image: string; sku_name: string; price: number; quantity: number }
type Order = { id: number; order_no: string; user_id: number; status: number; pay_amount: number; total_amount: number; remark: string; address: string; order_model: number; created_at: string; paid_at: string; shipped_at: string; completed_at: string; items?: OI[] }
type History = { id: number; original_status: number; new_status: number; msg: string; creator: string; created_at: string }

function LogisticsTrack({ orderId }: { orderId: number }) {
  const [data, setData] = useState<{ express_name: string; express_no: string; status: string; traces: { time: string; context: string }[] } | null>(null)
  useEffect(() => { api.get<typeof data>(`/admin/orders/${orderId}/logistics`).then(setData).catch(() => {}) }, [orderId])
  if (!data) return <Typography.Text type="secondary">加载中...</Typography.Text>
  return (
    <div>
      <Descriptions size="small" column={3}>
        <Descriptions.Item label="快递公司">{data.express_name}</Descriptions.Item>
        <Descriptions.Item label="快递单号">{data.express_no}</Descriptions.Item>
        <Descriptions.Item label="状态">{data.status}</Descriptions.Item>
      </Descriptions>
      {data.traces?.length > 0 && (
        <Timeline style={{ marginTop: 16 }} items={data.traces.map(t => ({ children: <><span style={{ color: '#999', fontSize: 12 }}>{t.time}</span> {t.context}</> }))} />
      )}
    </div>
  )
}

export default function OrderDetailPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const id = searchParams.get('id')
  const [order, setOrder] = useState<Order | null>(null)
  const [history, setHistory] = useState<History[]>([])
  const [shipOpen, setShipOpen] = useState(false)
  const [shipForm] = Form.useForm()

  useEffect(() => {
    if (!id) return
    api.get<{ list: Order[] }>(`/admin/orders?keyword=${id}&page=1&page_size=1`).then(r => { if (r.list?.[0]) setOrder(r.list[0]) })
    api.get<History[]>(`/orders/${id}/history`).then(r => setHistory(Array.isArray(r) ? r : [])).catch(() => {})
  }, [id])

  if (!order) return <Typography.Text>加载中...</Typography.Text>

  const addr = (() => { try { return JSON.parse(order.address) } catch { return null } })()
  const s = SM[order.status]

  return (
    <>
      <Space style={{ marginBottom: 16 }}>
        <Button onClick={() => router.push('/orders')}>← 返回列表</Button>
        <Typography.Title level={4} style={{ margin: 0 }}>订单详情</Typography.Title>
        <Button icon={<PrinterOutlined />} onClick={() => window.print()}>打印订单</Button>
      </Space>

      <style jsx global>{`
        @media print {
          .ant-layout-sider, .ant-layout-header, .ant-btn, .ant-card-head-title { display: none !important; }
          .ant-layout-content { margin: 0 !important; padding: 0 !important; }
          .ant-card { border: none !important; box-shadow: none !important; }
          body { -webkit-print-color-adjust: exact; }
        }
      `}</style>

      {/* 基础信息 */}
      <Card title="基础信息" size="small" style={{ marginBottom: 16 }}>
        <Descriptions column={3} size="small">
          <Descriptions.Item label="订单编号">{order.order_no}</Descriptions.Item>
          <Descriptions.Item label="用户ID">{order.user_id}</Descriptions.Item>
          <Descriptions.Item label="订单状态"><Tag color={s?.c}>{s?.t}</Tag></Descriptions.Item>
          <Descriptions.Item label="总金额">¥{(order.total_amount / 100).toFixed(2)}</Descriptions.Item>
          <Descriptions.Item label="实付金额">¥{(order.pay_amount / 100).toFixed(2)}</Descriptions.Item>
          <Descriptions.Item label="订单模式">{['快递', '同城', '自提', '虚拟'][order.order_model]}</Descriptions.Item>
          <Descriptions.Item label="下单时间">{new Date(order.created_at).toLocaleString('zh-CN')}</Descriptions.Item>
          <Descriptions.Item label="支付时间">{order.paid_at ? new Date(order.paid_at).toLocaleString('zh-CN') : '-'}</Descriptions.Item>
          <Descriptions.Item label="发货时间">{order.shipped_at ? new Date(order.shipped_at).toLocaleString('zh-CN') : '-'}</Descriptions.Item>
          <Descriptions.Item label="备注" span={3}>{order.remark || '-'}</Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 收货地址 */}
      {addr && (
        <Card title="收货地址" size="small" style={{ marginBottom: 16 }}>
          <Descriptions column={2} size="small">
            <Descriptions.Item label="收货人">{addr.name}</Descriptions.Item>
            <Descriptions.Item label="手机号">{addr.phone}</Descriptions.Item>
            <Descriptions.Item label="地址" span={2}>{addr.province}{addr.city}{addr.district} {addr.detail}</Descriptions.Item>
          </Descriptions>
        </Card>
      )}

      {/* 商品明细 */}
      <Card title="商品明细" size="small" style={{ marginBottom: 16 }}>
        <Table dataSource={order.items || []} rowKey="id" pagination={false} size="small"
          columns={[
            { title: '商品', dataIndex: 'title' },
            { title: '图片', dataIndex: 'image', width: 60, render: (v: string) => v ? <img src={v} alt="" style={{ width: 40, height: 40, objectFit: 'cover' }} /> : '-' },
            { title: '规格', dataIndex: 'sku_name', width: 100 },
            { title: '单价', dataIndex: 'price', width: 100, render: (v: number) => `¥${(v / 100).toFixed(2)}` },
            { title: '数量', dataIndex: 'quantity', width: 60 },
            { title: '小计', key: 'subtotal', width: 100, render: (_: unknown, r: OI) => `¥${(r.price * r.quantity / 100).toFixed(2)}` },
          ]}
          summary={() => (
            <Table.Summary.Row>
              <Table.Summary.Cell index={0} colSpan={5} align="right"><strong>合计</strong></Table.Summary.Cell>
              <Table.Summary.Cell index={5}><strong>¥{(order.pay_amount / 100).toFixed(2)}</strong></Table.Summary.Cell>
            </Table.Summary.Row>
          )}
        />
      </Card>

      {/* 操作按钮 */}
      <Card title="操作" size="small" style={{ marginBottom: 16 }}>
        <Space>
          {order.status === 0 && <Button type="primary" onClick={async () => { await api.put(`/admin/orders/pay-underline`, { order_id: order.id }); message.success('已确认收款'); router.push('/orders') }}>确认收款</Button>}
          {order.status === 1 && <Button type="primary" onClick={() => { shipForm.resetFields(); setShipOpen(true) }}>发货</Button>}
          {order.status === 0 && <Button danger onClick={async () => { await api.put(`/admin/orders/${order.id}/cancel`, {}); message.success('已取消'); router.push('/orders') }}>取消订单</Button>}
          {order.status === 2 && <Button onClick={async () => { await api.put(`/admin/orders/${order.id}/confirm`, {}); message.success('已确认收货'); router.push('/orders') }}>确认收货</Button>}
        </Space>
      </Card>

      {/* 物流轨迹 */}
      {(order.status === 2 || order.status === 3) && (
        <Card title="物流轨迹" size="small" style={{ marginBottom: 16 }}>
          <LogisticsTrack orderId={order.id} />
        </Card>
      )}

      {/* 状态历史 */}
      {history.length > 0 && (
        <Card title="状态变更记录" size="small">
          <Timeline items={history.map(h => ({
            children: <><Tag>{SM[h.new_status]?.t || h.new_status}</Tag> {h.msg} <span style={{ color: '#999', fontSize: 12 }}>({h.creator} {new Date(h.created_at).toLocaleString('zh-CN')})</span></>,
          }))} />
        </Card>
      )}

      <Modal title="发货" open={shipOpen} onCancel={() => setShipOpen(false)} onOk={async () => {
        const v = shipForm.getFieldsValue()
        await api.post('/admin/orders/ship', { order_id: order.id, ...v })
        message.success('已发货'); setShipOpen(false); router.push('/orders')
      }} forceRender>
        <Form form={shipForm} layout="vertical">
          <Form.Item name="express_no" label="快递单号" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="express_id" label="快递公司ID"><Input /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}
