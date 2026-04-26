'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Button, Modal, Form, Input, Select, Typography, Tag, message, Card, Row, Col, Space, Switch } from 'antd'
import { PlusOutlined, SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import DetailDrawer from '@/components/DetailDrawer'
import RichEditor from '@/components/RichEditor'
import ImageUpload from '@/components/ImageUpload'

type Article = { id: number; title: string; category_id: number; author: string; content: string; cover: string; status: number; access_count: number; created_at: string }
type ArtCat = { id: number; name: string }

export default function ArticlesPage() {
  const [list, setList] = useState<Article[]>([]); const [cats, setCats] = useState<ArtCat[]>([])
  const [total, setTotal] = useState(0); const [page, setPage] = useState(1); const [kw, setKw] = useState('')
  const [open, setOpen] = useState(false); const [editing, setEditing] = useState<Article | null>(null)
  const [detail, setDetail] = useState<Article | null>(null); const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [form] = Form.useForm()

  const load = useCallback(async (p = 1) => {
    const params = new URLSearchParams({ page: String(p), page_size: '20' }); if (kw) params.set('keyword', kw)
    const r = await api.get<{ total: number; list: Article[] } | Article[]>(`/articles?${params}`)
    if (Array.isArray(r)) { setList(r); setTotal(r.length) } else { setList(r.list || []); setTotal(r.total) }
    setPage(p); setSelectedIds([])
  }, [kw])

  useEffect(() => { load(); api.get<ArtCat[]>('/article-categories').then(r => setCats(Array.isArray(r) ? r : [])) }, [load])

  return (
    <>
      <Typography.Title level={4}>文章管理</Typography.Title>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={12}>
          <Col><Input placeholder="文章标题" prefix={<SearchOutlined />} value={kw} onChange={e => setKw(e.target.value)} onPressEnter={() => load()} allowClear style={{ width: 220 }} /></Col>
          <Col><Button type="primary" onClick={() => load()}>查询</Button></Col>
          <Col flex="auto" style={{ textAlign: 'right' }}><Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setOpen(true) }}>新增文章</Button></Col>
        </Row>
      </Card>
      <Table dataSource={list} rowKey="id"
        rowSelection={{ selectedRowKeys: selectedIds, onChange: keys => setSelectedIds(keys as number[]) }}
        pagination={{ current: page, total, pageSize: 20, onChange: p => load(p), showTotal: t => `共 ${t} 条` }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '标题', dataIndex: 'title', render: (v: string, r: Article) => <a onClick={() => setDetail(r)}>{v}</a> },
          { title: '分类', dataIndex: 'category_id', width: 100, render: (v: number) => cats.find(c => c.id === v)?.name || v },
          { title: '作者', dataIndex: 'author', width: 80 },
          { title: '浏览', dataIndex: 'access_count', width: 60 },
          { title: '状态', dataIndex: 'status', width: 80, render: (v: number, r: Article) => <Switch size="small" checked={v === 1} onChange={async s => { await api.put(`/admin/articles/${r.id}/status`, { status: s ? 1 : 0 }); load(page) }} /> },
          { title: '时间', dataIndex: 'created_at', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
          { title: '操作', width: 140, render: (_: unknown, r: Article) => (
            <Space>
              <a onClick={() => setDetail(r)}>详情</a>
              <a onClick={() => { setEditing(r); form.setFieldsValue(r); setOpen(true) }}>编辑</a>
              <a style={{ color: 'red' }} onClick={async () => { await api.del(`/admin/articles/${r.id}`); message.success('已删除'); load(page) }}>删除</a>
            </Space>
          )},
        ]}
      />
      <DetailDrawer open={!!detail} onClose={() => setDetail(null)} title={`文章详情 #${detail?.id}`} width={700} items={detail ? [
        { label: '标题', value: detail.title }, { label: '分类', value: cats.find(c => c.id === detail.category_id)?.name || '-' },
        { label: '作者', value: detail.author || '-' }, { label: '浏览量', value: detail.access_count },
        { label: '状态', value: detail.status === 1 ? <Tag color="green">发布</Tag> : <Tag>草稿</Tag> },
        { label: '内容', value: <div dangerouslySetInnerHTML={{ __html: detail.content || '-' }} style={{ maxHeight: 400, overflow: 'auto' }} /> },
        { label: '时间', value: new Date(detail.created_at).toLocaleString('zh-CN') },
      ] : []} />
      <Modal title={editing?.id ? '编辑文章' : '新增文章'} open={open} onCancel={() => setOpen(false)} onOk={() => form.submit()} forceRender width={700}>
        <Form form={form} layout="vertical" onFinish={async v => {
          if (editing?.id) await api.put(`/admin/articles/${editing.id}`, v); else await api.post('/admin/articles', v)
          message.success('保存成功'); setOpen(false); load(page)
        }}>
          <Form.Item name="title" label="标题" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="category_id" label="分类"><Select options={cats.map(c => ({ value: c.id, label: c.name }))} /></Form.Item>
          <Form.Item name="author" label="作者"><Input /></Form.Item>
          <Form.Item name="cover" label="封面图"><ImageUpload /></Form.Item>
          <Form.Item name="content" label="内容"><RichEditor /></Form.Item>
          <Form.Item name="status" label="状态" initialValue={1}><Select options={[{ value: 1, label: '发布' }, { value: 0, label: '草稿' }]} /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}
