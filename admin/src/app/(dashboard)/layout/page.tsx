'use client'
import CrudPage from '@/components/CrudPage'
import JsonConfigEditor from '@/components/JsonConfigEditor'
import { Form, Input, Select } from 'antd'
export default function LayoutPage() {
  return <CrudPage title="首页布局" listUrl="/admin/layouts" createUrl="/admin/layouts" searchable
    updateUrl={r => `/admin/layouts/${r.id}`}
    detailItems={r => { const d = r as Record<string,unknown>; return [{label:'ID',value:d.id as number},{label:'名称',value:d.name as string},{label:'类型',value:d.type as string},{label:'状态',value:(d.status as number)===1?'启用':'禁用'}] }}
    deleteUrl={r => `/admin/layouts/${r.id}`} batchDelete
    columns={[{title:'ID',dataIndex:'id',width:60},{title:'名称',dataIndex:'name'},{title:'类型',dataIndex:'type',width:80},{title:'状态',dataIndex:'status',width:60}]}
    formItems={() => (<><Form.Item name="name" label="名称" rules={[{required:true}]}><Input /></Form.Item><Form.Item name="type" label="类型"><Select options={[{value:'home',label:'首页'},{value:'category',label:'分类'},{value:'user',label:'用户中心'}]} /></Form.Item><Form.Item name="data" label="配置"><JsonConfigEditor /></Form.Item></>)} modalWidth={700} />
}
