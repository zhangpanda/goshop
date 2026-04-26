'use client'
import { useEffect, useState, useCallback, ReactNode } from 'react'
import { Table, Button, Modal, Form, Input, Typography, Space, Popconfirm, message, Switch, Card, Row, Col, Descriptions, Drawer } from 'antd'
import { PlusOutlined, SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import type { FormInstance } from 'antd'
import type { ColumnsType } from 'antd/es/table'

interface CrudPageProps<T> {
  title: string
  listUrl: string
  createUrl?: string
  updateUrl?: (r: T) => string
  deleteUrl?: (r: T) => string
  statusUrl?: (r: T) => string
  columns: ColumnsType<T>
  formItems: (form: FormInstance) => ReactNode
  detailItems?: (r: T) => { label: string; value: ReactNode }[]
  rowKey?: string
  paginated?: boolean
  modalWidth?: number
  searchable?: boolean
  searchPlaceholder?: string
  batchDelete?: boolean
  noSearch?: boolean
  noDetail?: boolean
}

export default function CrudPage<T extends { id: number }>({
  title, listUrl, createUrl, updateUrl, deleteUrl, statusUrl,
  columns, formItems, detailItems, rowKey = 'id', paginated = false, modalWidth = 520,
  searchable, searchPlaceholder = '搜索', batchDelete = false,
  noSearch = false, noDetail = false,
}: CrudPageProps<T>) {
  const [list, setList] = useState<T[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<T | null>(null)
  const [detail, setDetail] = useState<T | null>(null)
  const [keyword, setKeyword] = useState('')
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [form] = Form.useForm()

  // 默认开启搜索（除非 noSearch）
  const hasSearch = searchable !== undefined ? searchable : !noSearch

  const load = useCallback(async (p = 1) => {
    let url = paginated ? `${listUrl}${listUrl.includes('?') ? '&' : '?'}page=${p}&page_size=20` : listUrl
    if (hasSearch && keyword) url += `${url.includes('?') ? '&' : '?'}keyword=${encodeURIComponent(keyword)}`
    const res = await api.get<{ total?: number; list?: T[] } | T[]>(url)
    if (Array.isArray(res)) { setList(res); setTotal(res.length) }
    else { setList(res.list || []); setTotal(res.total || 0) }
    setPage(p); setSelectedIds([])
  }, [listUrl, paginated, hasSearch, keyword])

  useEffect(() => { load() }, [load])

  const onSave = async (v: Record<string, unknown>) => {
    if (editing && updateUrl) await api.put(updateUrl(editing), v)
    else if (createUrl) await api.post(createUrl, v)
    message.success('保存成功'); setOpen(false); form.resetFields(); setEditing(null); load(page)
  }

  const batchDel = async () => {
    if (!deleteUrl || !selectedIds.length) return
    for (const id of selectedIds) await api.del(deleteUrl({ id } as T))
    message.success(`已删除 ${selectedIds.length} 条`); load(page)
  }

  // 自动生成详情：从 columns 提取
  const autoDetailItems = (r: T): { label: string; value: ReactNode }[] => {
    const rec = r as Record<string, unknown>
    return columns.filter(c => 'dataIndex' in c && c.dataIndex).map(c => {
      const col = c as { title?: ReactNode; dataIndex?: string; render?: (v: unknown, r: T) => ReactNode }
      const val = rec[col.dataIndex!]
      const label = typeof col.title === 'string' ? col.title : String(col.dataIndex)
      // 时间字段格式化
      if (typeof val === 'string' && col.dataIndex?.includes('_at') && val) return { label, value: new Date(val).toLocaleString('zh-CN') }
      return { label, value: val != null ? String(val) : '-' }
    })
  }

  const showDetail = !noDetail
  const getDetailItems = detailItems || autoDetailItems

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const allCols: ColumnsType<any> = [
    ...columns,
    ...((statusUrl) ? [{
      title: '状态', dataIndex: 'status', width: 80,
      render: (v: number, r: T) => <Switch size="small" checked={v === 1} onChange={async s => { await api.put(statusUrl(r), { status: s ? 1 : 0 }); load(page) }} />,
    }] : []),
    ...((deleteUrl || updateUrl || showDetail) ? [{
      title: '操作', width: (showDetail ? 40 : 0) + (updateUrl ? 40 : 0) + (deleteUrl ? 40 : 0) + 40,
      render: (_: unknown, r: T) => (
        <Space>
          {showDetail && <a onClick={() => setDetail(r)}>详情</a>}
          {updateUrl && <a onClick={() => { setEditing(r); form.setFieldsValue(r); setOpen(true) }}>编辑</a>}
          {deleteUrl && <Popconfirm title="确认删除?" onConfirm={async () => { await api.del(deleteUrl(r)); message.success('已删除'); load(page) }}><a style={{ color: 'red' }}>删除</a></Popconfirm>}
        </Space>
      ),
    }] : []),
  ]

  return (
    <>
      <Typography.Title level={4}>{title}</Typography.Title>

      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={12} align="middle">
          {hasSearch && <Col><Input placeholder={searchPlaceholder} prefix={<SearchOutlined />} value={keyword} onChange={e => setKeyword(e.target.value)} onPressEnter={() => load()} allowClear style={{ width: 220 }} /></Col>}
          {hasSearch && <Col><Button type="primary" onClick={() => load()}>查询</Button></Col>}
          {hasSearch && keyword && <Col><Button onClick={() => { setKeyword(''); }}>重置</Button></Col>}
          {batchDelete && selectedIds.length > 0 && <Col><Popconfirm title={`确认删除 ${selectedIds.length} 条?`} onConfirm={batchDel}><Button danger size="small">批量删除({selectedIds.length})</Button></Popconfirm></Col>}
          {createUrl && <Col flex="auto" style={{ textAlign: 'right' }}><Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setOpen(true) }}>新增</Button></Col>}
        </Row>
      </Card>

      <Table dataSource={list} rowKey={rowKey} columns={allCols} size="small"
        rowSelection={batchDelete ? { selectedRowKeys: selectedIds, onChange: keys => setSelectedIds(keys as number[]) } : undefined}
        pagination={paginated ? { current: page, total, pageSize: 20, onChange: p => load(p), showTotal: t => `共 ${t} 条`, showSizeChanger: false } : { pageSize: 50, showTotal: t => `共 ${t} 条` }} />

      {(createUrl || updateUrl) && (
        <Modal title={editing ? '编辑' : '新增'} open={open} onCancel={() => { setOpen(false); form.resetFields() }} onOk={() => form.submit()} width={modalWidth} forceRender>
          <Form form={form} layout="vertical" onFinish={onSave}>{formItems(form)}</Form>
        </Modal>
      )}

      {showDetail && (
        <Drawer title={`${title}详情 #${detail?.id}`} open={!!detail} onClose={() => setDetail(null)} width={640}>
          {detail && (
            <Descriptions column={1} bordered size="small">
              {getDetailItems(detail).map((item, i) => (
                <Descriptions.Item key={i} label={item.label}>{item.value ?? '-'}</Descriptions.Item>
              ))}
            </Descriptions>
          )}
        </Drawer>
      )}
    </>
  )
}
