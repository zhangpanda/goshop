'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Input, Switch, Avatar, Typography, Space, Button, Card, Row, Col, Modal, Form, InputNumber, message } from 'antd'
import { SearchOutlined, UserOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import BatchActions from '@/components/BatchActions'
import ExportButton from '@/components/ExportButton'
import DetailDrawer from '@/components/DetailDrawer'

interface User { id: number; username: string; nickname: string; phone: string; email: string; avatar: string; points: number; wallet_balance: number; status: number; created_at: string }

export default function UsersPage() {
  const [list, setList] = useState<User[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [detail, setDetail] = useState<User | null>(null)
  const [editUser, setEditUser] = useState<User | null>(null)
  const [form] = Form.useForm()

  const load = useCallback(async (p = 1) => {
    const params = new URLSearchParams({ page: String(p), page_size: '20' })
    if (keyword) params.set('keyword', keyword)
    const res = await api.get<{ total: number; list: User[] }>(`/admin/users?${params}`)
    setList(res.list || []); setTotal(res.total); setPage(p); setSelectedIds([])
  }, [keyword])

  useEffect(() => { load() }, [load])

  return (
    <>
      <Typography.Title level={4}>用户管理</Typography.Title>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={12}>
          <Col><Input placeholder="用户名/昵称/手机号" prefix={<SearchOutlined />} value={keyword} onChange={e => setKeyword(e.target.value)} onPressEnter={() => load()} allowClear style={{ width: 240 }} /></Col>
          <Col><Button type="primary" onClick={() => load()}>查询</Button></Col>
          <Col><Button onClick={() => setKeyword('')}>重置</Button></Col>
          <Col flex="auto" style={{ textAlign: 'right' }}><ExportButton type="users" /></Col>
        </Row>
      </Card>

      <BatchActions selectedIds={selectedIds} deleteUrl="/admin/users" statusUrl="/admin/users"
        deleteButtonText="批量禁用" deleteConfirmTitle="确认将选中用户设为禁用？（不删除数据，保留订单关联）"
        exportUrl="/admin/export" exportType="users" onDone={() => load(page)} />

      <Table dataSource={list} rowKey="id"
        rowSelection={{ selectedRowKeys: selectedIds, onChange: keys => setSelectedIds(keys as number[]) }}
        pagination={{ current: page, total, pageSize: 20, onChange: p => load(p), showTotal: t => `共 ${t} 条` }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '头像', dataIndex: 'avatar', width: 50, render: (v: string) => <Avatar src={v || undefined} icon={<UserOutlined />} size="small" /> },
          { title: '用户名', dataIndex: 'username', render: (v: string, r: User) => <a onClick={() => setDetail(r)}>{v}</a> },
          { title: '昵称', dataIndex: 'nickname', width: 100 },
          { title: '手机', dataIndex: 'phone', width: 120 },
          { title: '邮箱', dataIndex: 'email', width: 160 },
          { title: '积分', dataIndex: 'points', width: 70 },
          { title: '余额', dataIndex: 'wallet_balance', width: 90, render: (v: number) => `¥${(v/100).toFixed(2)}` },
          { title: '状态', dataIndex: 'status', width: 70, render: (v: number, r: User) => <Switch size="small" checked={v === 1} onChange={async s => { await api.put(`/admin/users/${r.id}/status`, { status: s ? 1 : 0 }); load(page) }} /> },
          { title: '注册时间', dataIndex: 'created_at', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
          { title: '操作', width: 120, render: (_: unknown, r: User) => (
            <Space>
              <a onClick={() => { setEditUser(r); form.setFieldsValue(r) }}>编辑</a>
              <a onClick={() => setDetail(r)}>详情</a>
            </Space>
          )},
        ]}
      />

      <DetailDrawer open={!!detail} onClose={() => setDetail(null)} title={`用户详情 #${detail?.id}`}
        items={detail ? [
          { label: 'ID', value: detail.id },
          { label: '用户名', value: detail.username },
          { label: '昵称', value: detail.nickname },
          { label: '手机', value: detail.phone || '-' },
          { label: '邮箱', value: detail.email || '-' },
          { label: '头像', value: detail.avatar ? <Avatar src={detail.avatar} size={64} /> : '-' },
          { label: '积分', value: detail.points },
          { label: '钱包余额', value: `¥${(detail.wallet_balance / 100).toFixed(2)}` },
          { label: '状态', value: detail.status === 1 ? '正常' : '禁用' },
          { label: '注册时间', value: new Date(detail.created_at).toLocaleString('zh-CN') },
        ] : []}
      />

      <Modal title="编辑用户" open={!!editUser} onCancel={() => setEditUser(null)} onOk={() => form.submit()} forceRender width={500}>
        <Form form={form} layout="vertical" onFinish={async v => {
          await api.put(`/admin/users/${editUser!.id}/save`, v)
          message.success('已保存'); setEditUser(null); load(page)
        }}>
          <Form.Item name="nickname" label="昵称"><Input /></Form.Item>
          <Form.Item name="phone" label="手机"><Input /></Form.Item>
          <Form.Item name="email" label="邮箱"><Input /></Form.Item>
          <Form.Item name="password" label="新密码(留空不修改)"><Input.Password /></Form.Item>
          <Form.Item name="points" label="积分"><InputNumber style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="wallet_balance" label="钱包余额(分)"><InputNumber style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}
