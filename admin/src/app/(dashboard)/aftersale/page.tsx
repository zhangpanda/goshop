'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Tag, Space, Modal, Input, Typography, message, Card, Row, Col, Button, Tabs } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import { useUserMap } from '@/lib/useIdMap'
import DetailDrawer from '@/components/DetailDrawer'
import BatchActions from '@/components/BatchActions'

const TM: Record<number, string> = { 0: '仅退款', 1: '退货退款' }
const SM: Record<number, { t: string; c: string }> = { 0: { t: '待确认', c: 'orange' }, 1: { t: '待退货', c: 'blue' }, 2: { t: '待审核', c: 'cyan' }, 3: { t: '已完成', c: 'green' }, 4: { t: '已拒绝', c: 'red' }, 5: { t: '已取消', c: 'default' } }
type AS = { id: number; order_id: number; order_detail_id: number; user_id: number; goods_id: number; type: number; status: number; reason: string; price: number; number: number; msg: string; images: string; refuse_reason: string; express_name: string; express_no: string; created_at: string }

export default function AftersalePage() {
  const [list, setList] = useState<AS[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [status, setStatus] = useState(''); const [kw, setKw] = useState('')
  const userMap = useUserMap(list.map(r => r.user_id))
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [detail, setDetail] = useState<AS | null>(null)
  const [refuseId, setRefuseId] = useState<number | null>(null); const [reason, setReason] = useState('')

  const load = useCallback(async (p = 1) => {
    const params = new URLSearchParams({ page: String(p), page_size: '20' })
    if (status) params.set('status', status)
    if (kw) params.set('keyword', kw)
    const res = await api.get<{ total: number; list: AS[] }>(`/admin/aftersale?${params}`)
    setList(res.list || []); setTotal(res.total); setPage(p); setSelectedIds([])
  }, [status, kw])

  useEffect(() => { load() }, [load])

  return (
    <>
      <Typography.Title level={4}>订单售后</Typography.Title>
      <Tabs activeKey={status} onChange={v => setStatus(v)}
        items={[{ key: '', label: '全部' }, ...Object.entries(SM).map(([k, v]) => ({ key: k, label: v.t }))]} />
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={12}>
          <Col><Input placeholder="订单ID" prefix={<SearchOutlined />} value={kw} onChange={e => setKw(e.target.value)} onPressEnter={() => load()} allowClear style={{ width: 200 }} /></Col>
          <Col><Button type="primary" onClick={() => load()}>查询</Button></Col>
        </Row>
      </Card>
      <BatchActions selectedIds={selectedIds} onDone={() => load(page)} />
      <Table dataSource={list} rowKey="id"
        rowSelection={{ selectedRowKeys: selectedIds, onChange: keys => setSelectedIds(keys as number[]) }}
        pagination={{ current: page, total, pageSize: 20, onChange: p => load(p), showTotal: t => `共 ${t} 条` }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '订单ID', dataIndex: 'order_id', width: 80, render: (v: number, r: AS) => <a onClick={() => setDetail(r)}>{v}</a> },
          { title: '用户', dataIndex: 'user_id', width: 120, render: (v: number) => userMap[v] || `#${v}` },
          { title: '类型', dataIndex: 'type', width: 90, render: (v: number) => TM[v] || v },
          { title: '金额', dataIndex: 'price', width: 90, render: (v: number) => `¥${(v / 100).toFixed(2)}` },
          { title: '状态', dataIndex: 'status', width: 80, render: (v: number) => { const s = SM[v]; return s ? <Tag color={s.c}>{s.t}</Tag> : v } },
          { title: '原因', dataIndex: 'reason', ellipsis: true },
          { title: '时间', dataIndex: 'created_at', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
          { title: '操作', width: 180, render: (_: unknown, r: AS) => (
            <Space>
              <a onClick={() => setDetail(r)}>详情</a>
              {r.status === 0 && <a onClick={async () => { await api.put(`/admin/aftersale/${r.id}/audit`, { status: 1 }); message.success('已同意'); load(page) }}>同意</a>}
              {r.status === 0 && <a style={{ color: 'red' }} onClick={() => { setRefuseId(r.id); setReason('') }}>拒绝</a>}
              <a style={{ color: 'red' }} onClick={async () => { await api.del(`/admin/aftersale/${r.id}`); message.success('已删除'); load(page) }}>删除</a>
              {r.status === 1 && <a onClick={async () => { await api.put(`/admin/aftersale/${r.id}/confirm`, {}); message.success('已确认'); load(page) }}>确认完成</a>}
            </Space>
          )},
        ]}
      />
      <DetailDrawer open={!!detail} onClose={() => setDetail(null)} title={`售后详情 #${detail?.id}`} width={700} items={detail ? [
        { label: '售后ID', value: detail.id }, { label: '订单ID', value: detail.order_id },
        { label: '用户', value: userMap[detail.user_id] || `#${detail.user_id}` }, { label: '商品', value: `#${detail.goods_id}` },
        { label: '类型', value: TM[detail.type] || detail.type },
        { label: '状态', value: (() => { const s = SM[detail.status]; return s ? <Tag color={s.c}>{s.t}</Tag> : detail.status })() },
        { label: '退款金额', value: `¥${(detail.price / 100).toFixed(2)}` },
        { label: '退货数量', value: detail.number },
        { label: '申请原因', value: detail.reason || '-' },
        { label: '补充说明', value: detail.msg || '-' },
        { label: '凭证图片', value: (() => { try { const imgs = JSON.parse(detail.images || '[]'); return imgs.length ? <Space>{imgs.map((u:string,i:number) => <img key={i} src={u} style={{width:80,height:80,objectFit:'cover',borderRadius:4}} />)}</Space> : '-' } catch { return detail.images || '-' } })() },
        { label: '拒绝原因', value: detail.refuse_reason || '-' },
        { label: '退货快递', value: detail.express_name ? `${detail.express_name} ${detail.express_no}` : '-' },
        { label: '申请时间', value: new Date(detail.created_at).toLocaleString('zh-CN') },
      ] : []} />
      <Modal title="拒绝原因" open={refuseId !== null} onCancel={() => setRefuseId(null)} onOk={async () => {
        await api.put(`/admin/aftersale/${refuseId}/refuse`, { reason }); message.success('已拒绝'); setRefuseId(null); load(page)
      }}>
        <Input.TextArea rows={3} value={reason} onChange={e => setReason(e.target.value)} placeholder="请输入拒绝原因" />
      </Modal>
    </>
  )
}
