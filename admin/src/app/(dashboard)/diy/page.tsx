'use client'
import { useEffect, useState, useRef } from 'react'
import { Typography, Button, Table, Space, Switch, message, Modal } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'

type Diy = { id: number; name: string; data: string; access_count: number; status: number; created_at: string }

export default function DiyPage() {
  const [list, setList] = useState<Diy[]>([])
  const [iframeUrl, setIframeUrl] = useState('')
  const newId = useRef<number | null>(null)

  const load = () => api.get<Diy[]>('/admin/diy').then(r => setList(Array.isArray(r) ? r : []))
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
        const res = await api.post<{ id: number }>('/admin/diy', { name: '未命名页面', status: 0 })
        id = res.id
        newId.current = id
      } catch (e) { message.error((e as Error).message); return }
    }
    const token = localStorage.getItem('admin_token') || ''
    document.cookie = `admin_info=${encodeURIComponent(JSON.stringify({ token }))};path=/`
    setIframeUrl(`${base}/diy.html?id=${id}&token=${token}`)
  }

  const onClose = async () => {
    if (newId.current) {
      try {
        const pages = await api.get<Diy[]>('/admin/diy')
        const rec = (Array.isArray(pages) ? pages : []).find(d => d.id === newId.current)
        if (rec && !rec.data) {
          await api.del(`/admin/diy/${newId.current}`).catch(() => {})
        }
      } catch {}
    }
    newId.current = null
    setIframeUrl('')
    load()
  }

  return (
    <>
      <Typography.Title level={4}>DIY装修</Typography.Title>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => openEditor()}>新建页面</Button>
      </Space>
      <Table dataSource={list} rowKey="id" pagination={false}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '名称', dataIndex: 'name' },
          { title: '浏览量', dataIndex: 'access_count', width: 80 },
          { title: '状态', dataIndex: 'status', width: 80, render: (v: number, r: Diy) => <Switch size="small" checked={v===1} onChange={async s => { await api.put(`/admin/diy/${r.id}`, {status: s?1:0}); load() }} /> },
          { title: '操作', width: 200, render: (_: unknown, r: Diy) => (
            <Space>
              <a onClick={() => openEditor(r.id)}>编辑装修</a>
              <a style={{ color: 'red' }} onClick={async () => { await api.del(`/admin/diy/${r.id}`); message.success('已删除'); load() }}>删除</a>
            </Space>
          )},
        ]}
      />
      <Modal title="DIY页面装修" open={!!iframeUrl} onCancel={onClose}
        width="95vw" style={{ top: 20 }} footer={null} destroyOnClose>
        <iframe src={iframeUrl} style={{ width: '100%', height: '80vh', border: 'none' }} />
      </Modal>
    </>
  )
}
