'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Input, Tag, Space, Modal, Form, Typography, Tabs, Button, Card, Row, Col, Descriptions, Select, message } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import ExportButton from '@/components/ExportButton'
import { api } from '@/lib/api'
import { useUserMap } from '@/lib/useIdMap'
import { useRouter } from 'next/navigation'
import BatchActions from '@/components/BatchActions'
import DetailDrawer from '@/components/DetailDrawer'

const SM: Record<number, { text: string; color: string }> = {
  0: { text: '待付款', color: 'default' }, 1: { text: '已付款', color: 'blue' },
  2: { text: '已发货', color: 'cyan' }, 3: { text: '已完成', color: 'green' },
  4: { text: '已取消', color: 'orange' }, 5: { text: '已关闭', color: 'default' }, 6: { text: '退款中', color: 'red' },
}

interface OI { id: number; title: string; image: string; sku_name: string; price: number; quantity: number }
interface Order { id: number; order_no: string; user_id: number; status: number; pay_amount: number; total_amount: number; remark: string; address: string; order_model: number; created_at: string; paid_at: string; shipped_at: string; completed_at: string; items?: OI[] }

export default function OrdersPage() {
  const [list, setList] = useState<Order[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')
  const [keyword, setKeyword] = useState('')
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [detail, setDetail] = useState<Order | null>(null)
  const [shipModal, setShipModal] = useState<Order | null>(null)
  const [remarkModal, setRemarkModal] = useState<Order | null>(null)
  const [expressList, setExpressList] = useState<{id:number;name:string;code:string}[]>([])
  const [form] = Form.useForm()
  const [shipForm] = Form.useForm()
  const router = useRouter()
  const userMap = useUserMap(list.map(o => o.user_id))

  const load = useCallback(async (p = 1) => {
    const params = new URLSearchParams({ page: String(p), page_size: '20' })
    if (status) params.set('status', status)
    if (keyword) params.set('keyword', keyword)
    const res = await api.get<{ total: number; list: Order[] }>(`/admin/orders?${params}`)
    setList(res.list || []); setTotal(res.total); setPage(p); setSelectedIds([])
  }, [status, keyword])

  useEffect(() => { load() }, [load])
  useEffect(() => { api.get<{id:number;name:string;code:string}[]>('/express').then(r => setExpressList(r || [])) }, [])

  const parseAddr = (s: string) => { try { return JSON.parse(s) } catch { return null } }

  const orderOps = (r: Order) => (
    <Space>
      <a onClick={() => setDetail(r)}>详情</a>
      <a onClick={() => { setRemarkModal(r); form.setFieldsValue({ remark: r.remark }) }}>备注</a>
      {r.status === 1 && <a onClick={() => { setShipModal(r); shipForm.resetFields() }}>发货</a>}
      {r.status === 0 && <a style={{ color: 'red' }} onClick={async () => { await api.put(`/admin/orders/${r.id}/cancel`, {}); message.success('已取消'); load(page) }}>取消</a>}
      {r.status === 0 && <a onClick={async () => { await api.put(`/admin/orders/pay-underline`, { order_id: r.id }); message.success('已确认收款'); load(page) }}>确认收款</a>}
    </Space>
  )

  const detailItems = detail ? (() => {
    const addr = parseAddr(detail.address)
    return [
      { label: '订单编号', value: detail.order_no },
      { label: '用户', value: userMap[detail.user_id] || `#${detail.user_id}` },
      { label: '订单状态', value: (() => { const s = SM[detail.status]; return s ? <Tag color={s.color}>{s.text}</Tag> : detail.status })() },
      { label: '总金额', value: `¥${(detail.total_amount / 100).toFixed(2)}` },
      { label: '实付金额', value: `¥${(detail.pay_amount / 100).toFixed(2)}` },
      { label: '订单模式', value: ['快递', '同城', '自提', '虚拟'][detail.order_model] || detail.order_model },
      { label: '备注', value: detail.remark || '-' },
      { label: '下单时间', value: detail.created_at ? new Date(detail.created_at).toLocaleString('zh-CN') : '-' },
      { label: '支付时间', value: detail.paid_at ? new Date(detail.paid_at).toLocaleString('zh-CN') : '-' },
      { label: '发货时间', value: detail.shipped_at ? new Date(detail.shipped_at).toLocaleString('zh-CN') : '-' },
      { label: '完成时间', value: detail.completed_at ? new Date(detail.completed_at).toLocaleString('zh-CN') : '-' },
      { label: '收货人', value: addr ? `${addr.name} ${addr.phone}` : '-' },
      { label: '收货地址', value: addr ? `${addr.province}${addr.city}${addr.district} ${addr.detail}` : '-' },
      { label: '商品明细', value: (
        <Table dataSource={detail.items || []} rowKey="id" pagination={false} size="small"
          columns={[
            { title: '商品', dataIndex: 'title' },
            { title: '规格', dataIndex: 'sku_name' },
            { title: '单价', dataIndex: 'price', render: (v: number) => `¥${(v/100).toFixed(2)}` },
            { title: '数量', dataIndex: 'quantity' },
          ]} />
      )},
    ]
  })() : []

  return (
    <>
      <Typography.Title level={4}>订单管理</Typography.Title>
      <Tabs activeKey={status} onChange={v => setStatus(v)}
        items={[{ key: '', label: '全部' }, ...Object.entries(SM).map(([k, v]) => ({ key: k, label: v.text }))]} />

      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={12}>
          <Col><Input placeholder="订单号" prefix={<SearchOutlined />} value={keyword} onChange={e => setKeyword(e.target.value)} onPressEnter={() => load()} allowClear style={{ width: 240 }} /></Col>
          <Col><Button type="primary" onClick={() => load()}>查询</Button></Col>
          <Col><Button onClick={() => setKeyword('')}>重置</Button></Col>
          <Col flex="auto" style={{ textAlign: 'right' }}>
            <ExportButton type="orders" />
          </Col>
        </Row>
      </Card>

      <BatchActions selectedIds={selectedIds} deleteUrl="/admin/orders" exportUrl="/admin/export" onDone={() => load(page)} />

      <Table dataSource={list} rowKey="id"
        rowSelection={{ selectedRowKeys: selectedIds, onChange: keys => setSelectedIds(keys as number[]) }}
        pagination={{ current: page, total, pageSize: 20, onChange: p => load(p), showTotal: t => `共 ${t} 条` }}
        expandable={{ expandedRowRender: (r: Order) => (
          <Table dataSource={r.items || []} rowKey="id" pagination={false} size="small"
            columns={[
              { title: '商品', dataIndex: 'title' }, { title: '规格', dataIndex: 'sku_name' },
              { title: '单价', dataIndex: 'price', render: (v: number) => `¥${(v/100).toFixed(2)}` },
              { title: '数量', dataIndex: 'quantity' },
            ]} />
        )}}
        columns={[
          { title: '订单号', dataIndex: 'order_no', width: 190, render: (v: string, r: Order) => <a onClick={() => router.push(`/orders/detail?id=${r.id}`)}>{v}</a> },
          { title: '用户', dataIndex: 'user_id', width: 120, render: (v: number) => userMap[v] || `#${v}` },
          { title: '金额', dataIndex: 'pay_amount', width: 100, render: (v: number) => `¥${(v/100).toFixed(2)}` },
          { title: '状态', dataIndex: 'status', width: 80, render: (v: number) => { const s = SM[v]; return s ? <Tag color={s.color}>{s.text}</Tag> : v } },
          { title: '下单时间', dataIndex: 'created_at', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
          { title: '备注', dataIndex: 'remark', ellipsis: true, width: 120 },
          { title: '操作', width: 220, render: (_: unknown, r: Order) => orderOps(r) },
        ]}
      />

      <DetailDrawer open={!!detail} onClose={() => setDetail(null)} title={`订单详情 #${detail?.order_no}`} items={detailItems} width={800} />

      <Modal title="编辑备注" open={!!remarkModal} onCancel={() => setRemarkModal(null)} onOk={async () => {
        await api.put(`/admin/orders/${remarkModal!.id}/remark`, form.getFieldsValue())
        message.success('已更新'); setRemarkModal(null); load(page)
      }} forceRender>
        <Form form={form}><Form.Item name="remark"><Input.TextArea rows={3} /></Form.Item></Form>
      </Modal>

      <Modal title="发货" open={!!shipModal} onCancel={() => setShipModal(null)} onOk={async () => {
        await api.post('/admin/orders/ship', { order_id: shipModal!.id, ...shipForm.getFieldsValue() })
        message.success('已发货'); setShipModal(null); load(page)
      }} forceRender>
        <Form form={shipForm} layout="vertical">
          <Form.Item name="express_no" label="快递单号" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="express_id" label="快递公司"><Select options={expressList.map(e => ({value: e.id, label: e.name}))} placeholder="选择快递公司" showSearch optionFilterProp="label" /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}
