'use client'
import { useEffect, useState } from 'react'
import { Table, Button, Modal, Form, Input, Typography, message } from 'antd'
import { api } from '@/lib/api'
import RichEditor from '@/components/RichEditor'
export default function AgreementPage() {
  const [list, setList] = useState<{id:number;name:string;content:string}[]>([])
  const [editing, setEditing] = useState<typeof list[0]|null>(null)
  const [form] = Form.useForm()
  const load = () => api.get<{id:number;name:string;content:string}>('/agreement').then(r => setList(r ? [r] : []))
  useEffect(() => { load() }, [])
  return (<><Typography.Title level={4}>协议管理</Typography.Title>
    <Button type="primary" style={{marginBottom:16}} onClick={() => { form.resetFields(); setEditing({id:0,name:'',content:''})}}>新增/编辑</Button>
    <Table dataSource={list} rowKey="id" pagination={false} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'名称',dataIndex:'name'},{title:'操作',width:80,render:(_:unknown,r:typeof list[0])=><a onClick={()=>{setEditing(r);form.setFieldsValue(r)}}>编辑</a>},
    ]} />
    <Modal title="编辑协议" open={!!editing} onCancel={()=>setEditing(null)} onOk={()=>form.submit()} forceRender width={900}>
      <Form form={form} layout="vertical" onFinish={async v => { await api.post('/admin/agreement', v); message.success('保存成功'); setEditing(null); load() }}>
        <Form.Item name="name" label="名称" rules={[{required:true}]}><Input /></Form.Item>
        <Form.Item name="content" label="内容"><RichEditor /></Form.Item>
      </Form>
    </Modal>
  </>)
}
