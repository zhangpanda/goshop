'use client'
import { useEffect, useState } from 'react'
import { Tabs, Table, Button, Modal, Form, Input, Select, Space, Tree, TreeSelect, Typography, Switch, Popconfirm, message, Checkbox } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'

interface Role { id: number; name: string; description: string }
interface PluginRow { id: number; name: string; title: string; status: number }
interface Power { id: number; name: string; parent_id: number; action: string; sort: number; children?: Power[] }
interface Admin { id: number; username: string; role_id: number; status: number; created_at: string }
interface TreeNode { key: number; title: string; children?: TreeNode[] }

const toTreeData = (arr: Power[]): TreeNode[] =>
  arr.map(p => ({ key: p.id, title: `${p.name} (${p.action || '-'})`, ...(p.children?.length ? { children: toTreeData(p.children) } : {}) }))

export default function RBACPage() {
  const [roles, setRoles] = useState<Role[]>([])
  const [powers, setPowers] = useState<Power[]>([])
  const [admins, setAdmins] = useState<Admin[]>([])
  const [adminTotal, setAdminTotal] = useState(0)
  const [adminPage, setAdminPage] = useState(1)
  const [roleOpen, setRoleOpen] = useState(false)
  const [powerOpen, setPowerOpen] = useState(false)
  const [adminOpen, setAdminOpen] = useState(false)
  const [permOpen, setPermOpen] = useState<Role | null>(null)
  const [checkedKeys, setCheckedKeys] = useState<number[]>([])
  const [pluginOpen, setPluginOpen] = useState<Role | null>(null)
  const [plugins, setPlugins] = useState<PluginRow[]>([])
  const [pluginIds, setPluginIds] = useState<number[]>([])
  const [roleForm] = Form.useForm()
  const [powerForm] = Form.useForm()
  const [adminForm] = Form.useForm()

  const loadRoles = () => api.get<Role[]>('/admin/roles').then(r => setRoles(Array.isArray(r) ? r : []))
  const loadPowers = () => api.get<Power[]>('/admin/powers').then(r => setPowers(Array.isArray(r) ? r : []))
  const loadAdmins = async (p = 1) => {
    const res = await api.get<{ total: number; list: Admin[] }>(`/admin/admins?page=${p}&page_size=20`)
    setAdmins(res.list || []); setAdminTotal(res.total); setAdminPage(p)
  }

  useEffect(() => { loadRoles(); loadPowers(); loadAdmins() }, [])

  const openPerm = async (r: Role) => {
    setPermOpen(r)
    try { const ids = await api.get<number[]>(`/admin/roles/${r.id}/powers`); setCheckedKeys(ids || []) }
    catch { setCheckedKeys([]) }
  }

  const openPlugins = async (r: Role) => {
    setPluginOpen(r)
    try {
      const [plist, ids] = await Promise.all([
        api.get<PluginRow[]>('/admin/plugins'),
        api.get<number[]>(`/admin/roles/${r.id}/plugins`),
      ])
      setPlugins(Array.isArray(plist) ? plist : [])
      setPluginIds(Array.isArray(ids) ? ids : [])
    } catch {
      setPlugins([])
      setPluginIds([])
    }
  }

  return (
    <>
      <Typography.Title level={4}>权限管理</Typography.Title>
      <Tabs items={[
        { key: 'admins', label: '管理员', children: (
          <>
            <Button type="primary" icon={<PlusOutlined />} style={{ marginBottom: 16 }} onClick={() => { adminForm.resetFields(); setAdminOpen(true) }}>新增管理员</Button>
            <Table dataSource={admins} rowKey="id" pagination={{ current: adminPage, total: adminTotal, pageSize: 20, onChange: p => loadAdmins(p) }}
              columns={[
                { title: 'ID', dataIndex: 'id', width: 60 },
                { title: '用户名', dataIndex: 'username' },
                { title: '角色', dataIndex: 'role_id', width: 120, render: (v: number) => roles.find(r => r.id === v)?.name || v },
                { title: '状态', dataIndex: 'status', width: 80, render: (v: number, r: Admin) => <Switch size="small" checked={v === 1} onChange={async s => { await api.put(`/admin/admins/${r.id}/status`, { status: s ? 1 : 0 }); loadAdmins(adminPage) }} /> },
                { title: '创建时间', dataIndex: 'created_at', width: 180, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
              ]}
            />
          </>
        )},
        { key: 'roles', label: '角色管理', children: (
          <>
            <Button type="primary" icon={<PlusOutlined />} style={{ marginBottom: 16 }} onClick={() => { roleForm.resetFields(); setRoleOpen(true) }}>新增角色</Button>
            <Table dataSource={roles} rowKey="id" pagination={false}
              columns={[
                { title: 'ID', dataIndex: 'id', width: 60 },
                { title: '名称', dataIndex: 'name' },
                { title: '描述', dataIndex: 'description' },
                { title: '操作', width: 280, render: (_: unknown, r: Role) => (
                  <Space wrap>
                    <a onClick={() => openPerm(r)}>分配权限</a>
                    <a onClick={() => openPlugins(r)}>分配插件</a>
                    <Popconfirm title="确认删除?" onConfirm={async () => { await api.del(`/admin/roles/${r.id}`); loadRoles() }}>
                      <a style={{ color: 'red' }}>删除</a>
                    </Popconfirm>
                  </Space>
                )},
              ]}
            />
          </>
        )},
        { key: 'powers', label: '权限管理', children: (
          <>
            <Button type="primary" icon={<PlusOutlined />} style={{ marginBottom: 16 }} onClick={() => { powerForm.resetFields(); setPowerOpen(true) }}>新增权限</Button>
            <Tree treeData={toTreeData(powers)} defaultExpandAll selectable={false} />
          </>
        )},
      ]} />
      <Modal title="新增角色" open={roleOpen} onCancel={() => setRoleOpen(false)} onOk={() => roleForm.submit()} forceRender>
        <Form form={roleForm} layout="vertical" onFinish={async v => { await api.post('/admin/roles', v); message.success('已创建'); setRoleOpen(false); loadRoles() }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="description" label="描述"><Input /></Form.Item>
        </Form>
      </Modal>
      <Modal title="新增权限" open={powerOpen} onCancel={() => setPowerOpen(false)} onOk={() => powerForm.submit()} forceRender>
        <Form form={powerForm} layout="vertical" onFinish={async v => { await api.post('/admin/powers', v); message.success('已创建'); setPowerOpen(false); loadPowers() }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="action" label="操作标识"><Input placeholder="如 goods.list" /></Form.Item>
          <Form.Item name="parent_id" label="上级节点" initialValue={0}><TreeSelect treeData={[{value:0,title:'顶级节点'},...toTreeData(powers).map(function fix(n: any): any { return {value:n.key,title:n.title,children:n.children?.map(fix)} })]} placeholder="选择上级" treeDefaultExpandAll /></Form.Item>
          <Form.Item name="sort" label="排序" initialValue={0}><Input type="number" /></Form.Item>
        </Form>
      </Modal>
      <Modal title="新增管理员" open={adminOpen} onCancel={() => setAdminOpen(false)} onOk={() => adminForm.submit()} forceRender>
        <Form form={adminForm} layout="vertical" onFinish={async v => { await api.post('/admin/admins', v); message.success('已创建'); setAdminOpen(false); loadAdmins() }}>
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true }]}><Input.Password /></Form.Item>
          <Form.Item name="role_id" label="角色"><Select options={roles.map(r => ({ value: r.id, label: r.name }))} /></Form.Item>
        </Form>
      </Modal>
      <Modal title={`分配权限 - ${permOpen?.name}`} open={!!permOpen} onCancel={() => setPermOpen(null)} onOk={async () => {
        await api.put(`/admin/roles/${permOpen!.id}/powers`, { power_ids: checkedKeys })
        message.success('已保存'); setPermOpen(null)
      }} forceRender>
        <Tree checkable treeData={toTreeData(powers)} defaultExpandAll
          checkedKeys={checkedKeys} onCheck={keys => setCheckedKeys(keys as number[])} />
      </Modal>
      <Modal title={`分配插件 - ${pluginOpen?.name}`} open={!!pluginOpen} onCancel={() => setPluginOpen(null)} onOk={async () => {
        await api.put(`/admin/roles/${pluginOpen!.id}/plugins`, { plugin_ids: pluginIds })
        message.success('已保存'); setPluginOpen(null)
      }} forceRender>
        <Typography.Paragraph type="secondary" style={{ marginBottom: 12 }}>
          勾选该角色可使用的应用插件（与 ShopXO 角色插件概念对齐；具体能力仍依赖业务是否读取此关联）。
        </Typography.Paragraph>
        <Checkbox.Group
          value={pluginIds}
          onChange={v => setPluginIds(v as number[])}
          style={{ display: 'flex', flexDirection: 'column', gap: 8 }}
          options={plugins.map(p => ({
            value: p.id,
            label: `${p.title || p.name} (${p.name})${p.status === 1 ? '' : ' · 未启用'}`,
          }))}
        />
      </Modal>
    </>
  )
}
