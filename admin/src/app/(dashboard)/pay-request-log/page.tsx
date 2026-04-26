'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography, Card, Row, Col, Input, Button } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import DetailDrawer from '@/components/DetailDrawer'
type R = {id:number;pay_log_id:number;business:string;request:string;response:string;created_at:string}
export default function PayRequestLogPage() {
  const [list, setList] = useState<R[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [kw, setKw] = useState(''); const [detail, setDetail] = useState<R|null>(null)
  const load = useCallback(async (p=1) => { const params = new URLSearchParams({page:String(p),page_size:'20'}); if(kw) params.set('keyword',kw); const r = await api.get<{total:number;list:R[]}>(`/admin/pay-request-logs?${params}`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [kw])
  useEffect(() => { load() }, [load])
  return (<><Typography.Title level={4}>支付请求日志</Typography.Title>
    <Card size="small" style={{marginBottom:16}}><Row gutter={12}><Col><Input placeholder="支付日志ID" prefix={<SearchOutlined />} value={kw} onChange={e=>setKw(e.target.value)} onPressEnter={()=>load()} allowClear style={{width:200}} /></Col><Col><Button type="primary" onClick={()=>load()}>查询</Button></Col></Row></Card>
    <Table dataSource={list} rowKey="id" pagination={{current:page,total,pageSize:20,onChange:p=>load(p),showTotal:t=>`共 ${t} 条`}} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'支付日志ID',dataIndex:'pay_log_id',width:100,render:(v:number,r:R)=><a onClick={()=>setDetail(r)}>{v}</a>},
      {title:'业务类型',dataIndex:'business',width:80},{title:'请求',dataIndex:'request',ellipsis:true},{title:'响应',dataIndex:'response',ellipsis:true},
      {title:'时间',dataIndex:'created_at',width:170,render:(v:string)=>v?new Date(v).toLocaleString('zh-CN'):'-'},
    ]} />
    <DetailDrawer open={!!detail} onClose={()=>setDetail(null)} title={`请求详情 #${detail?.id}`} width={800} items={detail?[
      {label:'支付日志ID',value:detail.pay_log_id},{label:'业务类型',value:detail.business},
      {label:'请求内容',value:<pre style={{whiteSpace:'pre-wrap',maxHeight:300,overflow:'auto',background:'#f5f5f5',padding:8,borderRadius:4,fontSize:12}}>{detail.request}</pre>},
      {label:'响应内容',value:<pre style={{whiteSpace:'pre-wrap',maxHeight:300,overflow:'auto',background:'#f5f5f5',padding:8,borderRadius:4,fontSize:12}}>{detail.response}</pre>},
      {label:'时间',value:detail.created_at?new Date(detail.created_at).toLocaleString('zh-CN'):'-'},
    ]:[]} />
  </>)
}
