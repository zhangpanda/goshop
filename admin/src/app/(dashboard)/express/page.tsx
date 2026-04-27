'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, InputNumber } from 'antd'
import ImageUpload from '@/components/ImageUpload'
export default function ExpressPage() {
  return <CrudPage title="快递管理" listUrl="/admin/express" createUrl="/admin/express"
    searchClient searchPlaceholder="名称、编码"
    deleteUrl={r => `/admin/express/${r.id}`} statusUrl={r => `/admin/express/${r.id}/status`} updateUrl={r => `/admin/express/${r.id}`} batchDelete
    columns={[{title:'ID',dataIndex:'id',width:60},{title:'名称',dataIndex:'name'},{title:'编码',dataIndex:'code',width:100},{title:'图标',dataIndex:'icon',width:80,render:(v:string)=>v?<img src={v} style={{width:24,height:24}} />:'-'},{title:'排序',dataIndex:'sort',width:60}]}
    formItems={() => (<>
      <Form.Item name="name" label="名称" rules={[{required:true}]}><Input /></Form.Item>
      <Form.Item name="code" label="编码" rules={[{required:true}]}><Input placeholder="如 sf/yt/yd" /></Form.Item>
      <Form.Item name="icon" label="图标"><ImageUpload /></Form.Item>
      <Form.Item name="sort" label="排序" initialValue={0}><InputNumber min={0} /></Form.Item>
    </>)} />
}
