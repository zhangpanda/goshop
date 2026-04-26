'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, Select } from 'antd'
export default function AppMiniPage() {
  return <CrudPage title="小程序配置" listUrl="/admin/app-mini" createUrl="/admin/app-mini"
    deleteUrl={r => `/admin/app-mini/${r.id}`}
    columns={[{title:'ID',dataIndex:'id',width:60},{title:'平台',dataIndex:'platform',width:100},{title:'名称',dataIndex:'title'},{title:'AppID',dataIndex:'app_id',width:200}]}
    formItems={() => (<>
      <Form.Item name="platform" label="平台" rules={[{required:true}]}><Select options={[{value:'weixin',label:'微信'},{value:'alipay',label:'支付宝'},{value:'baidu',label:'百度'},{value:'toutiao',label:'头条'},{value:'qq',label:'QQ'},{value:'kuaishou',label:'快手'}]} /></Form.Item>
      <Form.Item name="title" label="名称"><Input /></Form.Item>
      <Form.Item name="describe" label="描述"><Input /></Form.Item>
      <Form.Item name="app_id" label="AppID"><Input /></Form.Item>
      <Form.Item name="app_secret" label="AppSecret"><Input.Password /></Form.Item>
    </>)} />
}
