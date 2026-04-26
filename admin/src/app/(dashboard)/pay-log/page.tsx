'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography, Tag, Card, Row, Col, Input, Button } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import { useUserMap } from '@/lib/useIdMap'
import DetailDrawer from '@/components/DetailDrawer'
const S: Record<number,{t:string;c:string}> = {0:{t:'待支付',c:'default'},1:{t:'已支付',c:'green'},2:{t:'已关闭',c:'red'}}
type R = {id:number;pay_no:string;user_id:number;total_price:number;trade_no:string;status:number;client_type:string;created_at:string;paid_at:string}
export default function PayLogPage() {
  const [list, setList] = useState<R[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [kw, setKw] = useState(''); const [detail, setDetail] = useState<R|null>(null)
  const userMap = useUserMap(list.map(r => r.user_id))
  const load = useCallback(async (p=1) => { const params = new URLSearchParams({page:String(p),page_size:'20'}); if(kw) params.set('keyword',kw); const r = await api.get<{total:number;list:R[]}>(`/admin/pay-logs?${params}`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [kw])
  useEffect(() => { load() }, [load])
  return (<><Typography.Title level={4}>支付日志</Typography.Title>
    <Card size="small" style={{marginBottom:16}}><Row gutter={12}><Col><Input placeholder="支付单号" prefix={<SearchOutlined />} value={kw} onChange={e=>setKw(e.target.value)} onPressEnter={()=>load()} allowClear style={{width:240}} /></Col><Col><Button type="primary" onClick={()=>load()}>查询</Button></Col></Row></Card>
    <Table dataSource={list} rowKey="id" pagination={{current:page,total,pageSize:20,onChange:p=>load(p),showTotal:t=>`共 ${t} 条`}} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'支付单号',dataIndex:'pay_no',width:200,render:(v:string,r:R)=><a onClick={()=>setDetail(r)}>{v}</a>},
      {title:'用户',dataIndex:'user_id',width:120,render:(v:number)=>userMap[v]||`#${v}`},{title:'金额',dataIndex:'total_price',width:100,render:(v:number)=>`¥${(v/100).toFixed(2)}`},
      {title:'状态',dataIndex:'status',width:80,render:(v:number)=>{const s=S[v];return s?<Tag color={s.c}>{s.t}</Tag>:v}},
      {title:'客户端',dataIndex:'client_type',width:80},{title:'时间',dataIndex:'created_at',width:170,render:(v:string)=>v?new Date(v).toLocaleString('zh-CN'):'-'},
    ]} />
    <DetailDrawer open={!!detail} onClose={()=>setDetail(null)} title={`支付详情 #${detail?.id}`} items={detail?[
      {label:'支付单号',value:detail.pay_no},{label:'用户',value:userMap[detail.user_id]||`#${detail.user_id}`},{label:'金额',value:`¥${(detail.total_price/100).toFixed(2)}`},
      {label:'第三方交易号',value:detail.trade_no||'-'},{label:'状态',value:S[detail.status]?.t||detail.status},
      {label:'客户端',value:detail.client_type||'-'},{label:'创建时间',value:detail.created_at?new Date(detail.created_at).toLocaleString('zh-CN'):'-'},
      {label:'支付时间',value:detail.paid_at?new Date(detail.paid_at).toLocaleString('zh-CN'):'-'},
    ]:[]} />
  </>)
}
