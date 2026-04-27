'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input } from 'antd'
export default function AttachmentCategoryPage() {
  return <CrudPage title="附件分类" listUrl="/admin/attachment-categories" createUrl="/admin/attachment-categories"
    searchClient searchPlaceholder="名称"
    deleteUrl={r => `/admin/attachment-categories/${r.id}`} updateUrl={r => `/admin/attachment-categories/${r.id}`} batchDelete
    columns={[{title:'ID',dataIndex:'id',width:60},{title:'名称',dataIndex:'name'},{title:'创建时间',dataIndex:'created_at',width:170,render:(v:string)=>v?new Date(v).toLocaleString('zh-CN'):'-'}]}
    formItems={() => (<><Form.Item name="name" label="分类名称" rules={[{required:true}]}><Input /></Form.Item></>)} />
}
