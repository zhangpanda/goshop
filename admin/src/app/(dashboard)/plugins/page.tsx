'use client'
import { useEffect, useState } from 'react'
import { Table, Button, Tag, Space, Typography, Modal, Form, Input, Tabs, Card, message, Drawer } from 'antd'
import { PlusOutlined, CloudDownloadOutlined, SettingOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'

type Plugin = { id: number; name: string; title: string; desc: string; author: string; version: string; config: string; status: number }

export default function PluginsPage() {
  const [list, setList] = useState<Plugin[]>([])
  const [installOpen, setInstallOpen] = useState(false)
  const [configPlugin, setConfigPlugin] = useState<Plugin | null>(null)
  const [form] = Form.useForm()
  const [configForm] = Form.useForm()

  const load = () => api.get<Plugin[]>('/admin/plugins').then(r => setList(Array.isArray(r) ? r : []))
  useEffect(() => { load() }, [])

  const openConfig = (p: Plugin) => {
    setConfigPlugin(p)
    try { configForm.setFieldsValue(JSON.parse(p.config || '{}')) } catch { configForm.resetFields() }
  }

  return (
    <>
      <Typography.Title level={4}>应用管理</Typography.Title>
      <Tabs items={[
        { key: 'installed', label: '已安装', children: (
          <>
            <Table dataSource={list.filter(p => p.status === 1)} rowKey="id" pagination={false}
              columns={[
                { title: '名称', dataIndex: 'title', render: (v: string, r: Plugin) => v || r.name },
                { title: '版本', dataIndex: 'version', width: 80 },
                { title: '作者', dataIndex: 'author', width: 100 },
                { title: '描述', dataIndex: 'desc', ellipsis: true },
                { title: '操作', width: 160, render: (_: unknown, r: Plugin) => (
                  <Space>
                    <a onClick={() => openConfig(r)}><SettingOutlined /> 配置</a>
                    <a style={{ color: 'red' }} onClick={async () => { await api.put(`/admin/plugins/${r.id}/uninstall`, {}); message.success('已卸载'); load() }}>卸载</a>
                  </Space>
                )},
              ]}
            />
          </>
        )},
        { key: 'all', label: '全部插件', children: (
          <>
            <Button type="primary" icon={<PlusOutlined />} style={{ marginBottom: 16 }} onClick={() => { form.resetFields(); setInstallOpen(true) }}>安装插件</Button>
            <Table dataSource={list} rowKey="id" pagination={false}
              columns={[
                { title: '标识', dataIndex: 'name', width: 120 },
                { title: '名称', dataIndex: 'title' },
                { title: '版本', dataIndex: 'version', width: 80 },
                { title: '状态', dataIndex: 'status', width: 80, render: (v: number) => v === 1 ? <Tag color="green">已安装</Tag> : <Tag>未安装</Tag> },
                { title: '操作', width: 100, render: (_: unknown, r: Plugin) => r.status !== 1 ? (
                  <a onClick={async () => { await api.post('/admin/plugins', { id: r.id }); message.success('已安装'); load() }}>安装</a>
                ) : <Tag color="green">已安装</Tag> },
              ]}
            />
          </>
        )},
        { key: 'store', label: '应用商店', children: (
          <Card>
            <div style={{ textAlign: 'center', padding: 40 }}>
              <CloudDownloadOutlined style={{ fontSize: 48, color: '#999' }} />
              <Typography.Title level={5} style={{ marginTop: 16 }}>应用商店</Typography.Title>
              <Typography.Text type="secondary">在线浏览和安装插件，敬请期待</Typography.Text>
            </div>
          </Card>
        )},
      ]} />

      <Modal title="安装插件" open={installOpen} onCancel={() => setInstallOpen(false)} onOk={() => form.submit()} forceRender>
        <Form form={form} layout="vertical" onFinish={async v => { await api.post('/admin/plugins', v); message.success('已安装'); setInstallOpen(false); load() }}>
          <Form.Item name="name" label="插件标识" rules={[{ required: true }]}><Input placeholder="如 coupon_share" /></Form.Item>
          <Form.Item name="title" label="插件名称"><Input /></Form.Item>
          <Form.Item name="version" label="版本号" initialValue="1.0.0"><Input /></Form.Item>
          <Form.Item name="author" label="作者"><Input /></Form.Item>
          <Form.Item name="desc" label="描述"><Input.TextArea rows={2} /></Form.Item>
        </Form>
      </Modal>

      <Drawer title={`${configPlugin?.title || configPlugin?.name} - 插件配置`} open={!!configPlugin} onClose={() => setConfigPlugin(null)} width={500}
        extra={<Button type="primary" onClick={async () => {
          const values = configForm.getFieldsValue()
          await api.post('/admin/plugin-config', { plugin_id: configPlugin!.id, config: JSON.stringify(values) })
          message.success('配置已保存'); setConfigPlugin(null); load()
        }}>保存</Button>}>
        <Form form={configForm} layout="vertical">
          <Typography.Text type="secondary">插件配置以JSON格式存储，可自定义字段</Typography.Text>
          {configPlugin && Object.entries(JSON.parse(configPlugin.config || '{}')).map(([k, v]) => (
            <Form.Item key={k} name={k} label={k}><Input defaultValue={String(v)} /></Form.Item>
          ))}
        </Form>
      </Drawer>
    </>
  )
}
