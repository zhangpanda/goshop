'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography, Card, Row, Col, Input, Button } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
type R = {id:number;admin_id:number;username:string;action:string;method:string;path:string;ip:string;created_at:string}
export default function OperationLogPage() {
  const [list, setList] = useState<R[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [kw, setKw] = useState('')
  const load = useCallback(async (p=1) => { const params = new URLSearchParams({page:String(p),page_size:'20'}); if(kw) params.set('keyword',kw); const r = await api.get<{total:number;list:R[]}>(`/admin/operation-logs?${params}`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [kw])
  useEffect(() => { load() }, [load])
  return (<><Typography.Title level={4}>操作审计日志</Typography.Title>
    <Card size="small" style={{marginBottom:16}}><Row gutter={12}><Col><Input placeholder="管理员/路径" prefix={<SearchOutlined />} value={kw} onChange={e=>setKw(e.target.value)} onPressEnter={()=>load()} allowClear style={{width:240}} /></Col><Col><Button type="primary" onClick={()=>load()}>查询</Button></Col></Row></Card>
    <Table dataSource={list} rowKey="id" pagination={{current:page,total,pageSize:20,onChange:p=>load(p),showTotal:t=>`共 ${t} 条`}} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'管理员',dataIndex:'username',width:100},{title:'操作',dataIndex:'action',width:200},
      {title:'路径',dataIndex:'path',ellipsis:true},{title:'IP',dataIndex:'ip',width:120},
      {title:'时间',dataIndex:'created_at',width:170,render:(v:string)=>v?new Date(v).toLocaleString('zh-CN'):'-'},
    ]} />
  </>)
}
