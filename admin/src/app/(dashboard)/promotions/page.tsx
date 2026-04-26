'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, Select, DatePicker, Typography, Tag, message, Card, Row, Col, Space } from 'antd'
import { PlusOutlined, SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import DetailDrawer from '@/components/DetailDrawer'

type Promo = { id: number; name: string; start_time: string; end_time: string; status: number; created_at: string }

export default function PromotionsPage() {
  const [list, setList] = useState<Promo[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [kw, setKw] = useState('')
  const [open, setOpen] = useState(false); const [editing, setEditing] = useState<Promo | null>(null)
  const [detail, setDetail] = useState<Promo | null>(null)
  const [form] = Form.useForm()

  const load = useCallback(async (p = 1) => {
    const params = new URLSearchParams({ page: String(p), page_size: '20' })
    if (kw) params.set('keyword', kw)
    const r = await api.get<{ total: number; list: Promo[] }>(`/admin/promotions?${params}`)
    setList(r.list || []); setTotal(r.total || 0); setPage(p)
  }, [kw])

  useEffect(() => { load(1) }, [load])

  const filtered = list

  return (
    <>
      <Typography.Title level={4}>促销活动</Typography.Title>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={12}>
          <Col><Input placeholder="活动名称" prefix={<SearchOutlined />} value={kw} onChange={e => setKw(e.target.value)} onPressEnter={() => load(1)} allowClear style={{ width: 220 }} /></Col>
          <Col><Button type="primary" onClick={() => load(1)}>查询</Button></Col>
          <Col flex="auto" style={{ textAlign: 'right' }}><Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setOpen(true) }}>新增活动</Button></Col>
        </Row>
      </Card>
      <Table dataSource={filtered} rowKey="id" pagination={{ current: page, total, pageSize: 20, onChange: p => load(p), showTotal: t => `共 ${t} 条` }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '名称', dataIndex: 'name', render: (v: string, r: Promo) => <a onClick={() => setDetail(r)}>{v}</a> },
          { title: '状态', dataIndex: 'status', width: 80, render: (v: number) => v === 1 ? <Tag color="green">启用</Tag> : <Tag>禁用</Tag> },
          { title: '开始', dataIndex: 'start_time', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
          { title: '结束', dataIndex: 'end_time', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
          { title: '操作', width: 140, render: (_: unknown, r: Promo) => (
            <Space>
              <a onClick={() => setDetail(r)}>详情</a>
              <a onClick={() => { setEditing(r); form.setFieldsValue(r); setOpen(true) }}>编辑</a>
              <a style={{ color: 'red' }} onClick={async () => { await api.del(`/admin/promotions/${r.id}`); message.success('已删除'); load(page) }}>删除</a>
            </Space>
          )},
        ]}
      />
      <DetailDrawer open={!!detail} onClose={() => setDetail(null)} title={`活动详情 #${detail?.id}`} items={detail ? [
        { label: '名称', value: detail.name }, { label: '状态', value: detail.status === 1 ? '启用' : '禁用' },
        { label: '开始时间', value: detail.start_time ? new Date(detail.start_time).toLocaleString('zh-CN') : '-' },
        { label: '结束时间', value: detail.end_time ? new Date(detail.end_time).toLocaleString('zh-CN') : '-' },
      ] : []} />
      <Modal title={editing?.id ? '编辑活动' : '新增活动'} open={open} onCancel={() => setOpen(false)} onOk={() => form.submit()} forceRender>
        <Form form={form} layout="vertical" onFinish={async v => {
          const data = { ...v, start_time: v.start_time?.format?.('YYYY-MM-DD HH:mm:ss') || v.start_time, end_time: v.end_time?.format?.('YYYY-MM-DD HH:mm:ss') || v.end_time }
          if (editing?.id) await api.put(`/admin/promotions/${editing.id}`, data); else await api.post('/admin/promotions', data)
          message.success('保存成功'); setOpen(false); load(page)
        }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="status" label="状态" initialValue={1}><Select options={[{ value: 1, label: '启用' }, { value: 0, label: '禁用' }]} /></Form.Item>
          <Form.Item name="start_time" label="开始时间"><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="end_time" label="结束时间"><DatePicker showTime style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}
