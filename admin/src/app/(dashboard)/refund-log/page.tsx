'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography, Tag, Card, Row, Col, Input, Button } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import DetailDrawer from '@/components/DetailDrawer'
const S: Record<number,{t:string;c:string}> = {0:{t:'处理中',c:'blue'},1:{t:'成功',c:'green'},2:{t:'失败',c:'red'}}
type R = {id:number;order_id:number;refund_no:string;trade_no:string;refund_price:number;reason:string;status:number;created_at:string}
export default function RefundLogPage() {
  const [list, setList] = useState<R[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [kw, setKw] = useState(''); const [detail, setDetail] = useState<R|null>(null)
  const load = useCallback(async (p=1) => { const params = new URLSearchParams({page:String(p),page_size:'20'}); if(kw) params.set('keyword',kw); const r = await api.get<{total:number;list:R[]}>(`/admin/refund-logs?${params}`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [kw])
  useEffect(() => { load() }, [load])
  return (<><Typography.Title level={4}>退款日志</Typography.Title>
    <Card size="small" style={{marginBottom:16}}><Row gutter={12}><Col><Input placeholder="退款单号" prefix={<SearchOutlined />} value={kw} onChange={e=>setKw(e.target.value)} onPressEnter={()=>load()} allowClear style={{width:240}} /></Col><Col><Button type="primary" onClick={()=>load()}>查询</Button></Col></Row></Card>
    <Table dataSource={list} rowKey="id" pagination={{current:page,total,pageSize:20,onChange:p=>load(p),showTotal:t=>`共 ${t} 条`}} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'订单ID',dataIndex:'order_id',width:80},
      {title:'退款单号',dataIndex:'refund_no',width:180,render:(v:string,r:R)=><a onClick={()=>setDetail(r)}>{v}</a>},
      {title:'金额',dataIndex:'refund_price',width:100,render:(v:number)=>`¥${(v/100).toFixed(2)}`},
      {title:'原因',dataIndex:'reason',ellipsis:true},{title:'状态',dataIndex:'status',width:80,render:(v:number)=>{const s=S[v];return s?<Tag color={s.c}>{s.t}</Tag>:v}},
      {title:'时间',dataIndex:'created_at',width:170,render:(v:string)=>v?new Date(v).toLocaleString('zh-CN'):'-'},
    ]} />
    <DetailDrawer open={!!detail} onClose={()=>setDetail(null)} title={`退款详情 #${detail?.id}`} items={detail?[
      {label:'退款单号',value:detail.refund_no},{label:'订单ID',value:detail.order_id},{label:'金额',value:`¥${(detail.refund_price/100).toFixed(2)}`},
      {label:'第三方交易号',value:detail.trade_no||'-'},{label:'原因',value:detail.reason||'-'},{label:'状态',value:S[detail.status]?.t||detail.status},
      {label:'时间',value:detail.created_at?new Date(detail.created_at).toLocaleString('zh-CN'):'-'},
    ]:[]} />
  </>)
}
