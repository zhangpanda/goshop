'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input } from 'antd'
import RichEditor from '@/components/RichEditor'
export default function CustomViewPage() {
  return <CrudPage title="自定义页面" listUrl="/admin/custom-views" createUrl="/admin/custom-views"
    searchClient searchPlaceholder="标题"
    deleteUrl={r => `/admin/custom-views/${r.id}`} statusUrl={r => `/admin/custom-views/${r.id}/status`} updateUrl={r => `/admin/custom-views/${r.id}`} batchDelete
    columns={[{ title: 'ID', dataIndex: 'id', width: 60 }, { title: '标题', dataIndex: 'title' }, { title: '状态', dataIndex: 'status', width: 60 }]}
    formItems={() => (<>
      <Form.Item name="title" label="标题" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="content" label="内容"><RichEditor /></Form.Item>
    </>)} modalWidth={900} />
}
