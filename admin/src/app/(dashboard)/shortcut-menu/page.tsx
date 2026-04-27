'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, InputNumber } from 'antd'
import ImageUpload from '@/components/ImageUpload'
export default function ShortcutMenuPage() {
  return <CrudPage title="快捷菜单" listUrl="/admin/shortcut-menus" createUrl="/admin/shortcut-menus" searchable searchClient searchPlaceholder="名称、链接" batchDelete
    updateUrl={r => `/admin/shortcut-menus/${r.id}`} deleteUrl={r => `/admin/shortcut-menus/${r.id}`}
    columns={[{title:'ID',dataIndex:'id',width:60},{title:'名称',dataIndex:'name'},{title:'图标',dataIndex:'icon',width:80,render:(v:string)=>v?<img src={v} style={{width:24,height:24}} />:'-'},{title:'链接',dataIndex:'url',ellipsis:true},{title:'排序',dataIndex:'sort',width:60}]}
    formItems={() => (<>
      <Form.Item name="name" label="名称" rules={[{required:true}]}><Input /></Form.Item>
      <Form.Item name="icon" label="图标"><ImageUpload /></Form.Item>
      <Form.Item name="url" label="链接"><Input /></Form.Item>
      <Form.Item name="sort" label="排序" initialValue={0}><InputNumber min={0} /></Form.Item>
    </>)} />
}
