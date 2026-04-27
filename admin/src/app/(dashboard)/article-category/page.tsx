'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, InputNumber } from 'antd'
export default function ArticleCategoryPage() {
  return <CrudPage title="文章分类" listUrl="/admin/article-categories" createUrl="/admin/article-categories"
    searchClient searchPlaceholder="名称"
    deleteUrl={r => `/admin/article-categories/${r.id}`} statusUrl={r => `/admin/article-categories/${r.id}/status`} updateUrl={r => `/admin/article-categories/${r.id}`} batchDelete
    columns={[{title:'ID',dataIndex:'id',width:60},{title:'名称',dataIndex:'name'},{title:'排序',dataIndex:'sort',width:60},{title:'状态',dataIndex:'status',width:60}]}
    formItems={() => (<><Form.Item name="name" label="名称" rules={[{required:true}]}><Input /></Form.Item><Form.Item name="sort" label="排序" initialValue={0}><InputNumber min={0} /></Form.Item></>)} />
}
