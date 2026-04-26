'use client'
import { useEffect, useState } from 'react'
import { Table, Button, Modal, Form, Input, InputNumber, Select, Switch, Typography, Tag, Space, message, Drawer, Descriptions } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import ImageUpload from '@/components/ImageUpload'

type Payment = { id: number; name: string; logo: string; config: string; sort: number; status: number; created_at: string }

const PAYMENT_FIELDS: Record<string, { label: string; key: string; type?: string }[]> = {
  '微信支付': [
    { label: 'AppID', key: 'app_id' }, { label: '商户号', key: 'mch_id' },
    { label: 'API密钥', key: 'api_key', type: 'password' }, { label: '证书序列号', key: 'serial_no' },
    { label: '私钥', key: 'private_key', type: 'textarea' }, { label: '回调地址', key: 'notify_url' },
  ],
  '支付宝': [
    { label: 'AppID', key: 'app_id' }, { label: '应用私钥', key: 'private_key', type: 'textarea' },
    { label: '支付宝公钥', key: 'alipay_public_key', type: 'textarea' }, { label: '回调地址', key: 'notify_url' },
  ],
  'PayPal': [
    { label: 'Client ID', key: 'client_id' }, { label: 'Secret', key: 'secret', type: 'password' },
    { label: '模式', key: 'mode' }, { label: '回调地址', key: 'notify_url' },
  ],
  '线下支付': [{ label: '说明', key: 'description', type: 'textarea' }],
  '钱包支付': [{ label: '说明', key: 'description', type: 'textarea' }],
}

export default function PaymentPage() {
  const [list, setList] = useState<Payment[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Payment | null>(null)
  const [configDrawer, setConfigDrawer] = useState<Payment | null>(null)
  const [form] = Form.useForm()
  const [configForm] = Form.useForm()

  const load = () => api.get<Payment[]>('/payments').then(r => setList(Array.isArray(r) ? r : []))
  useEffect(() => { load() }, [])

  const openConfig = (p: Payment) => {
    setConfigDrawer(p)
    try { configForm.setFieldsValue(JSON.parse(p.config || '{}')) } catch { configForm.resetFields() }
  }

  const saveConfig = async () => {
    const values = configForm.getFieldsValue()
    await api.put(`/admin/payments/${configDrawer!.id}`, { config: JSON.stringify(values) })
    message.success('配置已保存'); setConfigDrawer(null); load()
  }

  const fields = configDrawer ? (PAYMENT_FIELDS[configDrawer.name] || Object.keys(JSON.parse(configDrawer.config || '{}')).map(k => ({ label: k, key: k }))) : []

  return (
    <>
      <Typography.Title level={4}>支付方式</Typography.Title>
      <Button type="primary" icon={<PlusOutlined />} style={{ marginBottom: 16 }} onClick={() => { setEditing(null); form.resetFields(); setOpen(true) }}>新增支付方式</Button>
      <Table dataSource={list} rowKey="id" pagination={false}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: 'Logo', dataIndex: 'logo', width: 60, render: (v: string) => v ? <img src={v} alt="" style={{ height: 24 }} /> : '-' },
          { title: '名称', dataIndex: 'name' },
          { title: '排序', dataIndex: 'sort', width: 60 },
          { title: '状态', dataIndex: 'status', width: 80, render: (v: number, r: Payment) => <Switch size="small" checked={v === 1} onChange={async s => { await api.put(`/admin/payments/${r.id}/status`, { status: s ? 1 : 0 }); load() }} /> },
          { title: '操作', width: 200, render: (_: unknown, r: Payment) => (
            <Space>
              <a onClick={() => openConfig(r)}>配置</a>
              <a onClick={() => { setEditing(r); form.setFieldsValue(r); setOpen(true) }}>编辑</a>
              <a style={{ color: 'red' }} onClick={async () => { await api.del(`/admin/payments/${r.id}`); message.success('已删除'); load() }}>删除</a>
            </Space>
          )},
        ]}
      />

      <Modal title={editing?.id ? '编辑支付方式' : '新增支付方式'} open={open} onCancel={() => setOpen(false)} onOk={() => form.submit()} forceRender>
        <Form form={form} layout="vertical" onFinish={async v => {
          if (editing?.id) await api.put(`/admin/payments/${editing.id}`, v); else await api.post('/admin/payments', v)
          message.success('保存成功'); setOpen(false); load()
        }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Select options={['微信支付', '支付宝', 'PayPal', '线下支付', '钱包支付'].map(n => ({ value: n, label: n }))} />
          </Form.Item>
          <Form.Item name="logo" label="Logo"><ImageUpload /></Form.Item>
          <Form.Item name="sort" label="排序" initialValue={0}><InputNumber min={0} /></Form.Item>
        </Form>
      </Modal>

      <Drawer title={`${configDrawer?.name} - 支付配置`} open={!!configDrawer} onClose={() => setConfigDrawer(null)} width={560}
        extra={<Button type="primary" onClick={saveConfig}>保存配置</Button>}>
        <Form form={configForm} layout="vertical">
          {fields.map(f => (
            <Form.Item key={f.key} name={f.key} label={f.label}>
              {f.type === 'password' ? <Input.Password /> : f.type === 'textarea' ? <Input.TextArea rows={4} /> : <Input />}
            </Form.Item>
          ))}
          {!fields.length && <Typography.Text type="secondary">该支付方式暂无预设配置字段，配置将以JSON格式存储</Typography.Text>}
        </Form>
      </Drawer>
    </>
  )
}
