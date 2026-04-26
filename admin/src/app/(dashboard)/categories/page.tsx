'use client'
import { useEffect, useState } from 'react'
import { Table, Button, Modal, Form, Input, InputNumber, TreeSelect, Space, Typography, Popconfirm, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import ImageUpload from '@/components/ImageUpload'

interface Cat { id: number; name: string; icon: string; sort: number; parent_id: number; children?: Cat[] }
interface TreeNode { value: number; title: string; children?: TreeNode[] }

const toTreeData = (arr: Cat[]): TreeNode[] =>
  arr.map(c => ({ value: c.id, title: c.name, ...(c.children?.length ? { children: toTreeData(c.children) } : {}) }))

export default function CategoriesPage() {
  const [list, setList] = useState<Cat[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Cat | null>(null)
  const [form] = Form.useForm()

  const load = () => api.get<Cat[]>('/categories').then(setList)
  useEffect(() => { load() }, [])

  const onSave = async (v: Record<string, unknown>) => {
    if (editing) { await api.put(`/admin/categories/${editing.id}`, v) }
    else { await api.post('/admin/categories', v) }
    message.success('保存成功'); setOpen(false); form.resetFields(); setEditing(null); load()
  }

  return (
    <>
      <Typography.Title level={4}>分类管理</Typography.Title>
      <Button type="primary" icon={<PlusOutlined />} style={{ marginBottom: 16 }} onClick={() => { setEditing(null); form.resetFields(); setOpen(true) }}>新增分类</Button>
      <Table dataSource={list} rowKey="id" pagination={false} expandable={{ childrenColumnName: 'children' }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '名称', dataIndex: 'name' },
          { title: '图标', dataIndex: 'icon', width: 100, render: (v: string) => v ? <img src={v} style={{width:24,height:24}} /> : '-' },
          { title: '排序', dataIndex: 'sort', width: 80 },
          { title: '操作', width: 140, render: (_: unknown, r: Cat) => (
            <Space>
              <a onClick={() => { setEditing(r); form.setFieldsValue(r); setOpen(true) }}>编辑</a>
              <Popconfirm title="确认删除?" onConfirm={async () => {
                try { await api.del(`/admin/categories/${r.id}`); message.success('已删除'); load() }
                catch (e: unknown) { message.error((e as Error).message) }
              }}><a style={{ color: 'red' }}>删除</a></Popconfirm>
            </Space>
          )},
        ]}
      />
      <Modal title={editing ? '编辑分类' : '新增分类'} open={open} onCancel={() => setOpen(false)} onOk={() => form.submit()} forceRender>
        <Form form={form} layout="vertical" onFinish={onSave} initialValues={{ sort: 0, parent_id: 0 }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="parent_id" label="上级分类">
            <TreeSelect treeData={[{ value: 0, title: '顶级分类' }, ...toTreeData(list)]} placeholder="选择上级" />
          </Form.Item>
          <Form.Item name="icon" label="图标"><ImageUpload /></Form.Item>
          <Form.Item name="sort" label="排序"><InputNumber min={0} /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}
