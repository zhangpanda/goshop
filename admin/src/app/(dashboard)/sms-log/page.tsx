'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography, Tag, Card, Row, Col, Input, Button } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import DetailDrawer from '@/components/DetailDrawer'
const S: Record<number,{t:string;c:string}> = {0:{t:'发送中',c:'blue'},1:{t:'成功',c:'green'},2:{t:'失败',c:'red'}}
type R = {id:number;phone:string;content:string;type:string;status:number;created_at:string}
export default function SmsLogPage() {
  const [list, setList] = useState<R[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [kw, setKw] = useState(''); const [detail, setDetail] = useState<R|null>(null)
  const load = useCallback(async (p=1) => { const params = new URLSearchParams({page:String(p),page_size:'20'}); if(kw) params.set('keyword',kw); const r = await api.get<{total:number;list:R[]}>(`/admin/sms-logs?${params}`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [kw])
  useEffect(() => { load() }, [load])
  return (<><Typography.Title level={4}>短信日志</Typography.Title>
    <Card size="small" style={{marginBottom:16}}><Row gutter={12}><Col><Input placeholder="手机号" prefix={<SearchOutlined />} value={kw} onChange={e=>setKw(e.target.value)} onPressEnter={()=>load()} allowClear style={{width:200}} /></Col><Col><Button type="primary" onClick={()=>load()}>查询</Button></Col></Row></Card>
    <Table dataSource={list} rowKey="id" pagination={{current:page,total,pageSize:20,onChange:p=>load(p),showTotal:t=>`共 ${t} 条`}} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'手机号',dataIndex:'phone',width:130,render:(v:string,r:R)=><a onClick={()=>setDetail(r)}>{v}</a>},
      {title:'内容',dataIndex:'content',ellipsis:true},{title:'类型',dataIndex:'type',width:80},
      {title:'状态',dataIndex:'status',width:80,render:(v:number)=>{const s=S[v];return s?<Tag color={s.c}>{s.t}</Tag>:v}},
      {title:'操作',width:60,render:(_:unknown,r:R)=><a style={{color:'red'}} onClick={async()=>{await api.del(`/admin/sms-logs/${r.id}`);load(page)}}>删除</a>},
      {title:'时间',dataIndex:'created_at',width:170,render:(v:string)=>v?new Date(v).toLocaleString('zh-CN'):'-'},
    ]} />
    <DetailDrawer open={!!detail} onClose={()=>setDetail(null)} title={`短信详情 #${detail?.id}`} items={detail?[
      {label:'手机号',value:detail.phone},{label:'类型',value:detail.type},{label:'状态',value:S[detail.status]?.t||detail.status},
      {label:'内容',value:detail.content},{label:'时间',value:detail.created_at?new Date(detail.created_at).toLocaleString('zh-CN'):'-'},
    ]:[]} />
  </>)
}
