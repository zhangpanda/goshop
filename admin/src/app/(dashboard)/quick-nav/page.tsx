'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, InputNumber } from 'antd'
import ImageUpload from '@/components/ImageUpload'
export default function QuickNavPage() {
  return <CrudPage title="快捷导航" listUrl="/admin/quick-nav" createUrl="/admin/quick-nav"
    searchClient searchPlaceholder="名称、链接"
    deleteUrl={r => `/admin/quick-nav/${r.id}`} statusUrl={r => `/admin/quick-nav/${r.id}/status`} updateUrl={r => `/admin/quick-nav/${r.id}`} batchDelete
    columns={[{title:'ID',dataIndex:'id',width:60},{title:'名称',dataIndex:'name'},{title:'图标',dataIndex:'icon',width:80,render:(v:string)=>v?<img src={v} style={{width:24,height:24}} />:'-'},{title:'链接',dataIndex:'url',ellipsis:true},{title:'排序',dataIndex:'sort',width:60}]}
    formItems={() => (<>
      <Form.Item name="name" label="名称" rules={[{required:true}]}><Input /></Form.Item>
      <Form.Item name="icon" label="图标"><ImageUpload /></Form.Item>
      <Form.Item name="url" label="链接"><Input /></Form.Item>
      <Form.Item name="sort" label="排序" initialValue={0}><InputNumber min={0} /></Form.Item>
    </>)} />
}
