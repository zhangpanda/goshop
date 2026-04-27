'use client'
import CrudPage from '@/components/CrudPage'
import JsonConfigEditor from '@/components/JsonConfigEditor'
import { Form, Input } from 'antd'
export default function ThemePage() {
  return <CrudPage title="主题管理" listUrl="/admin/themes" createUrl="/admin/themes"
    searchClient searchPlaceholder="名称"
    deleteUrl={r => `/admin/themes/${r.id}`} updateUrl={r => `/admin/themes/${r.id}`} batchDelete
    columns={[{ title: 'ID', dataIndex: 'id', width: 60 }, { title: '名称', dataIndex: 'name' }]}
    formItems={() => (<><Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name="data" label="配置"><JsonConfigEditor /></Form.Item></>)} modalWidth={700} />
}
