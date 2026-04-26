'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input } from 'antd'
export default function ThemePage() {
  return <CrudPage title="主题管理" listUrl="/admin/themes" createUrl="/admin/themes"
    deleteUrl={r => `/admin/themes/${r.id}`} updateUrl={r => `/admin/themes/${r.id}`} batchDelete
    columns={[{ title: 'ID', dataIndex: 'id', width: 60 }, { title: '名称', dataIndex: 'name' }]}
    formItems={() => (<><Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name="data" label="配置(JSON)"><Input.TextArea rows={6} /></Form.Item></>)} />
}
