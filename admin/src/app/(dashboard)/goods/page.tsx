'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Button, Input, Select, Space, Tag, Typography, Switch, message, Row, Col, Card } from 'antd'
import { PlusOutlined, SearchOutlined } from '@ant-design/icons'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'
import BatchActions from '@/components/BatchActions'
import ExportButton from '@/components/ExportButton'
import DetailDrawer from '@/components/DetailDrawer'

interface SKU { id: number; price: number; stock: number; name: string }
interface Goods { id: number; title: string; subtitle: string; main_image: string; category_id: number; status: number; sort: number; sales_count: number; access_count: number; skus?: SKU[]; category?: { id: number; name: string }; created_at: string }
interface Cat { id: number; name: string; children?: Cat[] }

export default function GoodsPage() {
  const [list, setList] = useState<Goods[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<number | undefined>()
  const [catId, setCatId] = useState<number | undefined>()
  const [cats, setCats] = useState<Cat[]>([])
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [detail, setDetail] = useState<Goods | null>(null)
  const router = useRouter()

  const load = useCallback(async (p = 1) => {
    const params = new URLSearchParams({ page: String(p), page_size: '20' })
    if (keyword) params.set('keyword', keyword)
    if (status !== undefined) params.set('status', String(status))
    if (catId) params.set('category_id', String(catId))
    const res = await api.get<{ total: number; list: Goods[] }>(`/goods?${params}`)
    setList(res.list || []); setTotal(res.total); setPage(p); setSelectedIds([])
  }, [keyword, status, catId])

  useEffect(() => { load(); api.get<Cat[]>('/categories').then(setCats) }, [load])

  const flatCats = (arr: Cat[], prefix = ''): { value: number; label: string }[] =>
    arr.flatMap(c => [{ value: c.id, label: prefix + c.name }, ...(c.children ? flatCats(c.children, prefix + c.name + '/') : [])])

  return (
    <>
      <Typography.Title level={4}>商品管理</Typography.Title>

      {/* 搜索筛选 */}
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={[12, 12]}>
          <Col><Input placeholder="商品名称/ID" prefix={<SearchOutlined />} value={keyword} onChange={e => setKeyword(e.target.value)} onPressEnter={() => load()} allowClear style={{ width: 200 }} /></Col>
          <Col><Select placeholder="分类" allowClear style={{ width: 160 }} value={catId} onChange={v => { setCatId(v); }} options={flatCats(cats)} /></Col>
          <Col><Select placeholder="状态" allowClear style={{ width: 120 }} value={status} onChange={v => setStatus(v)} options={[{ value: 1, label: '上架' }, { value: 0, label: '下架' }]} /></Col>
          <Col><Button type="primary" onClick={() => load()}>查询</Button></Col>
          <Col><Button onClick={() => { setKeyword(''); setStatus(undefined); setCatId(undefined) }}>重置</Button></Col>
          <Col flex="auto" style={{ textAlign: 'right' }}>
            <ExportButton type="goods" />
            <Button type="primary" icon={<PlusOutlined />} onClick={() => router.push('/goods/edit')}>新增商品</Button>
          </Col>
        </Row>
      </Card>

      {/* 批量操作 */}
      <BatchActions selectedIds={selectedIds} deleteUrl="/admin/goods" statusUrl="/admin/goods" onDone={() => load(page)} />

      {/* 列表 */}
      <Table dataSource={list} rowKey="id"
        rowSelection={{ selectedRowKeys: selectedIds, onChange: keys => setSelectedIds(keys as number[]) }}
        pagination={{ current: page, total, pageSize: 20, onChange: p => load(p), showTotal: t => `共 ${t} 条` }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '主图', dataIndex: 'main_image', width: 70, render: (v: string) => v ? <img src={v} alt="" style={{ width: 40, height: 40, objectFit: 'cover', borderRadius: 4 }} /> : <div style={{ width: 40, height: 40, background: '#f5f5f5', borderRadius: 4 }} /> },
          { title: '商品名称', dataIndex: 'title', ellipsis: true, render: (v: string, r: Goods) => <a onClick={() => setDetail(r)}>{v}</a> },
          { title: '分类', key: 'cat', width: 100, render: (_: unknown, r: Goods) => r.category?.name || '-' },
          { title: '价格', key: 'price', width: 110, render: (_: unknown, r: Goods) => r.skus?.[0] ? `¥${(r.skus[0].price / 100).toFixed(2)}` : '-' },
          { title: '库存', key: 'stock', width: 70, render: (_: unknown, r: Goods) => r.skus?.[0]?.stock ?? '-' },
          { title: '销量', dataIndex: 'sales_count', width: 70 },
          { title: '浏览', dataIndex: 'access_count', width: 70 },
          { title: '状态', dataIndex: 'status', width: 80, render: (v: number, r: Goods) => <Switch size="small" checked={v === 1} onChange={async s => { await api.put(`/admin/goods/${r.id}/status`, { status: s ? 1 : 0 }); load(page) }} /> },
          { title: '创建时间', dataIndex: 'created_at', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
          { title: '操作', width: 120, render: (_: unknown, r: Goods) => (
            <Space>
              <a onClick={() => router.push(`/goods/edit?id=${r.id}`)}>编辑</a>
              <a onClick={() => setDetail(r)}>详情</a>
            </Space>
          )},
        ]}
      />

      {/* 详情抽屉 */}
      <DetailDrawer open={!!detail} onClose={() => setDetail(null)} title={`商品详情 #${detail?.id}`}
        items={detail ? [
          { label: 'ID', value: detail.id },
          { label: '标题', value: detail.title },
          { label: '副标题', value: detail.subtitle },
          { label: '分类', value: detail.category?.name },
          { label: '主图', value: detail.main_image ? <img src={detail.main_image} alt="" style={{ maxWidth: 200 }} /> : '-' },
          { label: '价格', value: detail.skus?.[0] ? `¥${(detail.skus[0].price / 100).toFixed(2)}` : '-' },
          { label: '库存', value: detail.skus?.[0]?.stock },
          { label: '销量', value: detail.sales_count },
          { label: '浏览量', value: detail.access_count },
          { label: '状态', value: detail.status === 1 ? <Tag color="green">上架</Tag> : <Tag>下架</Tag> },
          { label: '创建时间', value: new Date(detail.created_at).toLocaleString('zh-CN') },
          { label: 'SKU列表', value: detail.skus?.map(s => `${s.name}: ¥${(s.price/100).toFixed(2)} (库存${s.stock})`).join('\n') },
        ] : []}
      />
    </>
  )
}
