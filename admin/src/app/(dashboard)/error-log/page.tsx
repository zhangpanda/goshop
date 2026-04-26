'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography, Card, Row, Col, Input, Button } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import DetailDrawer from '@/components/DetailDrawer'
type R = {id:number;type:string;content:string;url:string;ip:string;created_at:string}
export default function ErrorLogPage() {
  const [list, setList] = useState<R[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [kw, setKw] = useState(''); const [detail, setDetail] = useState<R|null>(null)
  const load = useCallback(async (p=1) => { const params = new URLSearchParams({page:String(p),page_size:'20'}); if(kw) params.set('keyword',kw); const r = await api.get<{total:number;list:R[]}>(`/admin/error-logs?${params}`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [kw])
  useEffect(() => { load() }, [load])
  return (<><Typography.Title level={4}>错误日志</Typography.Title>
    <Card size="small" style={{marginBottom:16}}><Row gutter={12}><Col><Input placeholder="URL/IP" prefix={<SearchOutlined />} value={kw} onChange={e=>setKw(e.target.value)} onPressEnter={()=>load()} allowClear style={{width:240}} /></Col><Col><Button type="primary" onClick={()=>load()}>查询</Button></Col></Row></Card>
    <Table dataSource={list} rowKey="id" pagination={{current:page,total,pageSize:20,onChange:p=>load(p),showTotal:t=>`共 ${t} 条`}} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'类型',dataIndex:'type',width:80,render:(v:string,r:R)=><a onClick={()=>setDetail(r)}>{v}</a>},
      {title:'内容',dataIndex:'content',ellipsis:true},{title:'URL',dataIndex:'url',ellipsis:true},{title:'IP',dataIndex:'ip',width:120},
      {title:'操作',width:60,render:(_:unknown,r:R)=><a style={{color:'red'}} onClick={async()=>{await api.del(`/admin/error-logs/${r.id}`);load(page)}}>删除</a>},
      {title:'时间',dataIndex:'created_at',width:170,render:(v:string)=>v?new Date(v).toLocaleString('zh-CN'):'-'},
    ]} />
    <DetailDrawer open={!!detail} onClose={()=>setDetail(null)} title={`错误详情 #${detail?.id}`} width={700} items={detail?[
      {label:'类型',value:detail.type},{label:'URL',value:detail.url},{label:'IP',value:detail.ip},
      {label:'内容',value:<pre style={{whiteSpace:'pre-wrap',background:'#f5f5f5',padding:8,borderRadius:4,fontSize:12,maxHeight:400,overflow:'auto'}}>{detail.content}</pre>},
      {label:'时间',value:detail.created_at?new Date(detail.created_at).toLocaleString('zh-CN'):'-'},
    ]:[]} />
  </>)
}
