'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input } from 'antd'
export default function ThemeDataPage() {
  return <CrudPage title="主题数据" listUrl="/admin/themes" createUrl="/admin/themes"
    deleteUrl={r => `/admin/themes/${r.id}`} updateUrl={r => `/admin/themes/${r.id}`} batchDelete
    columns={[{ title: 'ID', dataIndex: 'id', width: 60 }, { title: '名称', dataIndex: 'name' }]}
    formItems={() => (<><Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item></>)} />
}
