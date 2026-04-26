'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, InputNumber } from 'antd'
import ImageUpload from '@/components/ImageUpload'

export default function LinksPage() {
  return <CrudPage title="友情链接" listUrl="/admin/links" createUrl="/admin/links"
    deleteUrl={r => `/admin/links/${r.id}`} statusUrl={r => `/admin/links/${r.id}/status`} updateUrl={r => `/admin/links/${r.id}`} batchDelete
    columns={[
      { title: 'ID', dataIndex: 'id', width: 60 },
      { title: '名称', dataIndex: 'name' },
      { title: 'Logo', dataIndex: 'logo', width: 80, render: (v: string) => v ? <img src={v} style={{height:24}} /> : '-' },
      { title: '链接', dataIndex: 'url', ellipsis: true },
      { title: '排序', dataIndex: 'sort', width: 60 },
    ]}
    formItems={() => (<>
      <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="url" label="链接"><Input /></Form.Item>
      <Form.Item name="logo" label="Logo"><ImageUpload /></Form.Item>
      <Form.Item name="sort" label="排序" initialValue={0}><InputNumber min={0} /></Form.Item>
    </>)}
  />
}
