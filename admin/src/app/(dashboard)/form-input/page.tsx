'use client'
import { useEffect, useState, useRef } from 'react'
import { Typography, Button, Table, Space, Switch, message, Modal } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'

type FormItem = { id: number; name: string; config: string; status: number; created_at: string }

export default function FormInputPage() {
  const [list, setList] = useState<FormItem[]>([])
  const [iframeUrl, setIframeUrl] = useState('')
  const newId = useRef<number | null>(null)

  const load = () => api.get<FormItem[]>('/admin/forms').then(r => setList(Array.isArray(r) ? r : []))
  useEffect(() => { load() }, [])

  useEffect(() => {
    const handler = (e: MessageEvent) => { if (e.data === 'diy-close') onClose() }
    window.addEventListener('message', handler)
    return () => window.removeEventListener('message', handler)
  })

  const openEditor = async (id?: number) => {
    const base = window.location.origin.replace(/:\d+$/, ':8080')
    newId.current = null
    if (!id) {
      try {
        const res = await api.post<{ id: number }>('/admin/forms', { name: '未命名表单', status: 1 })
        id = res.id
        newId.current = id
      } catch (e) { message.error((e as Error).message); return }
    }
    const token = localStorage.getItem('admin_token') || ''
    document.cookie = `admin_info=${encodeURIComponent(JSON.stringify({ token }))};path=/`
    setIframeUrl(`${base}/form.html?id=${id}&token=${token}`)
  }

  const onClose = async () => {
    // If we created a new record, check if it was actually saved (has config)
    if (newId.current) {
      try {
        const forms = await api.get<FormItem[]>('/admin/forms')
        const rec = (Array.isArray(forms) ? forms : []).find(f => f.id === newId.current)
        if (rec && !rec.config) {
          await api.del(`/admin/forms/${newId.current}`).catch(() => {})
        }
      } catch {}
    }
    newId.current = null
    setIframeUrl('')
    load()
  }

  return (
    <>
      <Typography.Title level={4}>Form表单设计</Typography.Title>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()}>新建表单</Button>
      </Space>
      <Table dataSource={list} rowKey="id" pagination={false}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '名称', dataIndex: 'name' },
          { title: '状态', dataIndex: 'status', width: 80, render: (v: number, r: FormItem) => <Switch size="small" checked={v===1} onChange={async s => { await api.put(`/admin/forms/${r.id}/fields`, {status: s?1:0}); load() }} /> },
          { title: '操作', width: 200, render: (_: unknown, r: FormItem) => (
            <Space>
              <a onClick={() => openEditor(r.id)}>编辑表单</a>
              <a style={{ color: 'red' }} onClick={async () => { await api.del(`/admin/forms/${r.id}`); message.success('已删除'); load() }}>删除</a>
            </Space>
          )},
        ]}
      />
      <Modal title="表单设计器" open={!!iframeUrl} onCancel={onClose}
        width="95vw" style={{ top: 20 }} footer={null} destroyOnClose>
        <iframe src={iframeUrl} style={{ width: '100%', height: '80vh', border: 'none' }} />
      </Modal>
    </>
  )
}
