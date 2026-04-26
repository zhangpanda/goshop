'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, InputNumber, Select, DatePicker, Typography, Tag, message, Card, Row, Col, Space } from 'antd'
import { PlusOutlined, SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import DetailDrawer from '@/components/DetailDrawer'
import BatchActions from '@/components/BatchActions'

type Coupon = { id: number; name: string; type: number; min_amount: number; value: number; total: number; received: number; per_limit: number; start_time: string; end_time: string; status: number; created_at: string }

export default function CouponsPage() {
  const [list, setList] = useState<Coupon[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [kw, setKw] = useState(''); const [open, setOpen] = useState(false); const [editing, setEditing] = useState<Coupon | null>(null)
  const [detail, setDetail] = useState<Coupon | null>(null); const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [form] = Form.useForm()

  const load = useCallback(async (p = 1) => {
    const params = new URLSearchParams({ page: String(p), page_size: '20' }); if (kw) params.set('keyword', kw)
    const r = await api.get<Coupon[] | { total: number; list: Coupon[] }>(`/coupons?${params}`)
    if (Array.isArray(r)) { setList(r); setTotal(r.length) } else { setList(r.list || []); setTotal(r.total) }
    setPage(p); setSelectedIds([])
  }, [kw])

  useEffect(() => { load() }, [load])

  return (
    <>
      <Typography.Title level={4}>优惠券管理</Typography.Title>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={12}>
          <Col><Input placeholder="优惠券名称" prefix={<SearchOutlined />} value={kw} onChange={e => setKw(e.target.value)} onPressEnter={() => load()} allowClear style={{ width: 220 }} /></Col>
          <Col><Button type="primary" onClick={() => load()}>查询</Button></Col>
          <Col flex="auto" style={{ textAlign: 'right' }}><Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setOpen(true) }}>新增优惠券</Button></Col>
        </Row>
      </Card>
      <BatchActions selectedIds={selectedIds} deleteUrl="/admin/coupons" onDone={() => load(page)} />
      <Table dataSource={list} rowKey="id"
        rowSelection={{ selectedRowKeys: selectedIds, onChange: keys => setSelectedIds(keys as number[]) }}
        pagination={{ current: page, total, pageSize: 20, onChange: p => load(p), showTotal: t => `共 ${t} 条` }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '名称', dataIndex: 'name', render: (v: string, r: Coupon) => <a onClick={() => setDetail(r)}>{v}</a> },
          { title: '类型', dataIndex: 'type', width: 70, render: (v: number) => v === 1 ? <Tag>满减</Tag> : v === 2 ? <Tag color="blue">折扣</Tag> : <Tag color="green">无门槛</Tag> },
          { title: '面值', dataIndex: 'value', width: 70 },
          { title: '门槛', dataIndex: 'min_amount', width: 70 },
          { title: '总量', dataIndex: 'total', width: 60 },
          { title: '已领', dataIndex: 'received', width: 60 },
          { title: '开始', dataIndex: 'start_time', width: 110, render: (v: string) => v ? new Date(v).toLocaleDateString('zh-CN') : '-' },
          { title: '结束', dataIndex: 'end_time', width: 110, render: (v: string) => v ? new Date(v).toLocaleDateString('zh-CN') : '-' },
          { title: '操作', width: 100, render: (_: unknown, r: Coupon) => <Space><a onClick={() => setDetail(r)}>详情</a><a onClick={() => { setEditing(r); form.setFieldsValue(r); setOpen(true) }}>编辑</a></Space> },
        ]}
      />
      <DetailDrawer open={!!detail} onClose={() => setDetail(null)} title={`优惠券详情 #${detail?.id}`} items={detail ? [
        { label: '名称', value: detail.name }, { label: '类型', value: detail.type === 1 ? '满减' : detail.type === 2 ? '折扣' : '无门槛' },
        { label: '面值', value: detail.value }, { label: '最低消费', value: detail.min_amount },
        { label: '总量', value: detail.total }, { label: '已领取', value: detail.received }, { label: '每人限领', value: detail.per_limit },
        { label: '开始时间', value: detail.start_time ? new Date(detail.start_time).toLocaleString('zh-CN') : '-' },
        { label: '结束时间', value: detail.end_time ? new Date(detail.end_time).toLocaleString('zh-CN') : '-' },
        { label: '状态', value: detail.status === 1 ? <Tag color="green">启用</Tag> : <Tag>禁用</Tag> },
      ] : []} />
      <Modal title={editing?.id ? '编辑优惠券' : '新增优惠券'} open={open} onCancel={() => setOpen(false)} onOk={() => form.submit()} forceRender width={500}>
        <Form form={form} layout="vertical" onFinish={async v => {
          const data = { ...v, start_time: v.start_time?.format?.('YYYY-MM-DD HH:mm:ss') || v.start_time, end_time: v.end_time?.format?.('YYYY-MM-DD HH:mm:ss') || v.end_time }
          if (editing?.id) await api.put(`/admin/coupons/${editing.id}`, data); else await api.post('/admin/coupons', data)
          message.success('保存成功'); setOpen(false); load(page)
        }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="type" label="类型" initialValue={1}><Select options={[{ value: 1, label: '满减' }, { value: 2, label: '折扣' }, { value: 3, label: '无门槛' }]} /></Form.Item>
          <Form.Item name="value" label="面值" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="min_amount" label="最低消费" initialValue={0}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="total" label="发放总量" initialValue={100}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="per_limit" label="每人限领" initialValue={1}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="start_time" label="开始时间"><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="end_time" label="结束时间"><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}
