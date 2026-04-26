'use client'
import { useEffect, useState } from 'react'
import { Table, Button, Modal, Form, Input, Switch, Space, Typography, Popconfirm, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'

interface WH { id: number; name: string; address: string; contact: string; phone: string; is_default: number; status: number }

export default function WarehousePage() {
  const [list, setList] = useState<WH[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<WH | null>(null)
  const [form] = Form.useForm()

  const load = () => api.get<WH[]>('/admin/warehouses').then(r => setList(Array.isArray(r) ? r : []))
  useEffect(() => { load() }, [])

  const onSave = async (v: Record<string, unknown>) => {
    if (editing) { await api.put(`/admin/warehouses/${editing.id}`, v) }
    else { await api.post('/admin/warehouses', v) }
    message.success('保存成功'); setOpen(false); setEditing(null); load()
  }

  return (
    <>
      <Typography.Title level={4}>仓库管理</Typography.Title>
      <Button type="primary" icon={<PlusOutlined />} style={{ marginBottom: 16 }} onClick={() => { setEditing(null); form.resetFields(); setOpen(true) }}>新增仓库</Button>
      <Table dataSource={list} rowKey="id" pagination={false}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '名称', dataIndex: 'name' },
          { title: '地址', dataIndex: 'address' },
          { title: '联系人', dataIndex: 'contact', width: 100 },
          { title: '电话', dataIndex: 'phone', width: 120 },
          { title: '默认', dataIndex: 'is_default', width: 60, render: (v: number) => v ? '是' : '否' },
          { title: '操作', width: 140, render: (_: unknown, r: WH) => (
            <Space>
              <a onClick={() => { setEditing(r); form.setFieldsValue(r); setOpen(true) }}>编辑</a>
              <Popconfirm title="确认删除?" onConfirm={async () => { await api.del(`/admin/warehouses/${r.id}`); message.success('已删除'); load() }}>
                <a style={{ color: 'red' }}>删除</a>
              </Popconfirm>
            </Space>
          )},
        ]}
      />
      <Modal title={editing ? '编辑仓库' : '新增仓库'} open={open} onCancel={() => setOpen(false)} onOk={() => form.submit()} forceRender>
        <Form form={form} layout="vertical" onFinish={onSave}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="address" label="地址"><Input /></Form.Item>
          <Form.Item name="contact" label="联系人"><Input /></Form.Item>
          <Form.Item name="phone" label="电话"><Input /></Form.Item>
          <Form.Item name="is_default" label="默认仓库" valuePropName="checked"><Switch /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}
