'use client'
import { useEffect, useState, useCallback } from 'react'
import { Tabs, Table, Button, Modal, Form, InputNumber, Select, Space, Tag, Typography, message, Popconfirm } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'

interface User { id: number; username: string; nickname: string; phone: string }
interface Dist { id: number; user_id: number; parent_id: number; level: number; total_commission: number; balance: number; order_count: number; status: number }
interface Withdraw { id: number; distributor_id: number; user_id: number; amount: number; status: number; account_type: string; account_no: string; account_name: string; created_at: string }

// 用户名缓存
const userCache: Record<number, User> = {}
async function resolveUsers(ids: number[]) {
  const missing = ids.filter(id => id > 0 && !userCache[id])
  if (missing.length === 0) return
  const list = await api.get<{ list: User[] }>(`/admin/users?ids=${missing.join(',')}`)
  const users = (list as unknown as { list?: User[] })?.list || list as unknown as User[]
  if (Array.isArray(users)) users.forEach(u => { userCache[u.id] = u })
}
function userName(id: number) {
  if (!id) return '-'
  const u = userCache[id]
  return u ? `${u.nickname || u.username}(${u.id})` : `#${id}`
}

function DistributorList() {
  const [list, setList] = useState<Dist[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [addOpen, setAddOpen] = useState(false)
  const [users, setUsers] = useState<User[]>([])
  const [form] = Form.useForm()
  const [, setTick] = useState(0) // force re-render after user resolve

  const load = useCallback(async () => {
    const r = await api.get<{ total: number; list: Dist[] }>(`/admin/distributors?page=${page}&page_size=20`)
    const items = r?.list || []
    setList(items); setTotal(r?.total || 0)
    const ids = [...new Set(items.flatMap(d => [d.user_id, d.parent_id]))]
    await resolveUsers(ids)
    setTick(t => t + 1)
  }, [page])
  useEffect(() => { load() }, [load])

  const searchUsers = async (kw: string) => {
    if (kw.length < 1) return
    const r = await api.get<{ list: User[] }>(`/admin/users?keyword=${kw}&page_size=10`)
    setUsers(r?.list || [])
  }

  const onAdd = async (v: { user_id: number; parent_id?: number }) => {
    await api.post('/admin/distributors', v)
    message.success('添加成功'); setAddOpen(false); form.resetFields(); load()
  }

  return (<>
    <Button type="primary" icon={<PlusOutlined />} style={{ marginBottom: 16 }} onClick={() => setAddOpen(true)}>添加分销商</Button>
    <Table dataSource={list} rowKey="id" pagination={{ total, current: page, pageSize: 20, onChange: setPage }}
      columns={[
        { title: 'ID', dataIndex: 'id', width: 60 },
        { title: '用户', dataIndex: 'user_id', width: 160, render: (v: number) => userName(v) },
        { title: '上级', dataIndex: 'parent_id', width: 160, render: (v: number) => userName(v) },
        { title: '等级', dataIndex: 'level', width: 60 },
        { title: '累计佣金', dataIndex: 'total_commission', render: (v: number) => `¥${(v / 100).toFixed(2)}` },
        { title: '可提现', dataIndex: 'balance', render: (v: number) => `¥${(v / 100).toFixed(2)}` },
        { title: '推广订单', dataIndex: 'order_count', width: 80 },
        { title: '状态', dataIndex: 'status', width: 70, render: (v: number) => <Tag color={v === 1 ? 'green' : 'red'}>{v === 1 ? '正常' : '冻结'}</Tag> },
      ]} />
    <Modal title="添加分销商" open={addOpen} onCancel={() => setAddOpen(false)} onOk={() => form.submit()} destroyOnClose>
      <Form form={form} layout="vertical" onFinish={onAdd}>
        <Form.Item name="user_id" label="选择用户" rules={[{ required: true, message: '请选择用户' }]}>
          <Select showSearch placeholder="搜索用户名/手机号" filterOption={false} onSearch={searchUsers}
            options={users.map(u => ({ value: u.id, label: `${u.nickname || u.username} (${u.phone || u.id})` }))} />
        </Form.Item>
        <Form.Item name="parent_id" label="上级分销商(可选)">
          <Select showSearch allowClear placeholder="搜索上级用户" filterOption={false} onSearch={searchUsers}
            options={users.map(u => ({ value: u.id, label: `${u.nickname || u.username} (${u.phone || u.id})` }))} />
        </Form.Item>
      </Form>
    </Modal>
  </>)
}

function WithdrawList() {
  const [list, setList] = useState<Withdraw[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [, setTick] = useState(0)

  const load = useCallback(async () => {
    const r = await api.get<{ total: number; list: Withdraw[] }>(`/admin/withdraws?page=${page}&page_size=20`)
    const items = r?.list || []
    setList(items); setTotal(r?.total || 0)
    await resolveUsers(items.map(w => w.user_id))
    setTick(t => t + 1)
  }, [page])
  useEffect(() => { load() }, [load])

  return (
    <Table dataSource={list} rowKey="id" pagination={{ total, current: page, pageSize: 20, onChange: setPage }}
      columns={[
        { title: 'ID', dataIndex: 'id', width: 60 },
        { title: '用户', dataIndex: 'user_id', width: 160, render: (v: number) => userName(v) },
        { title: '金额', dataIndex: 'amount', render: (v: number) => `¥${(v / 100).toFixed(2)}` },
        { title: '账户', width: 200, render: (_: unknown, r: Withdraw) => `${r.account_type} ${r.account_name} ${r.account_no}` },
        { title: '状态', dataIndex: 'status', width: 80, render: (v: number) => <Tag color={['blue','green','red','cyan'][v]}>{['待审核','已通过','已拒绝','已打款'][v]}</Tag> },
        { title: '时间', dataIndex: 'created_at', width: 170 },
        { title: '操作', width: 140, render: (_: unknown, r: Withdraw) => r.status === 0 ? (
          <Space>
            <Popconfirm title="确认通过?" onConfirm={async () => { await api.put(`/admin/withdraws/${r.id}/audit`, { approve: true }); message.success('已通过'); load() }}><a>通过</a></Popconfirm>
            <Popconfirm title="确认拒绝?" onConfirm={async () => { await api.put(`/admin/withdraws/${r.id}/audit`, { approve: false, reason: '审核不通过' }); message.success('已拒绝'); load() }}><a style={{ color: 'red' }}>拒绝</a></Popconfirm>
          </Space>
        ) : '-' },
      ]} />
  )
}

function RateConfig() {
  const [form] = Form.useForm()
  useEffect(() => {
    api.get<{ key: string; value: string }[]>('/admin/config?group=distribution').then(cfg => {
      const map: Record<string, string> = {}
      if (Array.isArray(cfg)) cfg.forEach(c => { map[c.key] = c.value })
      form.setFieldsValue({ level1: parseInt(map['distribution_rate_level1'] || '10'), level2: parseInt(map['distribution_rate_level2'] || '5') })
    })
  }, [form])
  const onSave = async (v: { level1: number; level2: number }) => {
    await api.post('/admin/config', { key: 'distribution_rate_level1', value: String(v.level1), group: 'distribution', desc: '一级分销佣金比例(%)' })
    await api.post('/admin/config', { key: 'distribution_rate_level2', value: String(v.level2), group: 'distribution', desc: '二级分销佣金比例(%)' })
    message.success('保存成功')
  }
  return (
    <div style={{ padding: 24 }}>
      <Typography.Paragraph type="secondary">订单完成后自动按比例结算佣金给上级分销商</Typography.Paragraph>
      <Form form={form} layout="inline" onFinish={onSave}>
        <Form.Item name="level1" label="一级佣金比例(%)"><InputNumber min={0} max={100} /></Form.Item>
        <Form.Item name="level2" label="二级佣金比例(%)"><InputNumber min={0} max={100} /></Form.Item>
        <Form.Item><Button type="primary" htmlType="submit">保存</Button></Form.Item>
      </Form>
    </div>
  )
}

export default function DistributionPage() {
  return (<>
    <Typography.Title level={4}>分销管理</Typography.Title>
    <Tabs items={[
      { key: 'distributors', label: '分销商', children: <DistributorList /> },
      { key: 'withdraws', label: '提现管理', children: <WithdrawList /> },
      { key: 'config', label: '佣金设置', children: <RateConfig /> },
    ]} />
  </>)
}
