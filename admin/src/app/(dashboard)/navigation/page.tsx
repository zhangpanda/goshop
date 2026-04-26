'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, InputNumber, Select } from 'antd'

export default function NavigationPage() {
  return <CrudPage title="导航管理" listUrl="/admin/navigations" createUrl="/admin/navigations"
    updateUrl={r => `/admin/navigations/${r.id}`} deleteUrl={r => `/admin/navigations/${r.id}`} statusUrl={r => `/admin/navigations/${r.id}/status`}
    searchable searchPlaceholder="导航名称" batchDelete
    detailItems={r => {
      const d = r as Record<string, unknown>
      return [{ label: 'ID', value: d.id as number }, { label: '名称', value: d.name as string }, { label: '链接', value: d.url as string }, { label: '类型', value: d.type === 'header' ? '顶部' : d.type === 'footer' ? '底部' : d.type as string }, { label: '排序', value: d.sort as number }]
    }}
    columns={[
      { title: 'ID', dataIndex: 'id', width: 60 },
      { title: '名称', dataIndex: 'name' },
      { title: '链接', dataIndex: 'url', ellipsis: true },
      { title: '类型', dataIndex: 'type', width: 80, render: (v: string) => v === 'header' ? '顶部' : v === 'footer' ? '底部' : v },
      { title: '排序', dataIndex: 'sort', width: 60 },
    ]}
    formItems={() => (<>
      <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="url" label="链接"><Input /></Form.Item>
      <Form.Item name="type" label="类型" initialValue="header"><Select options={[{ value: 'header', label: '顶部' }, { value: 'footer', label: '底部' }]} /></Form.Item>
      <Form.Item name="sort" label="排序" initialValue={0}><InputNumber min={0} /></Form.Item>
    </>)}
  />
}
