'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography, Tag, Card, Row, Col, Input, Button } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import DetailDrawer from '@/components/DetailDrawer'
const S: Record<number,{t:string;c:string}> = {0:{t:'发送中',c:'blue'},1:{t:'成功',c:'green'},2:{t:'失败',c:'red'}}
type R = {id:number;email:string;title:string;content:string;status:number;created_at:string}
export default function EmailLogPage() {
  const [list, setList] = useState<R[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [kw, setKw] = useState(''); const [detail, setDetail] = useState<R|null>(null)
  const load = useCallback(async (p=1) => { const params = new URLSearchParams({page:String(p),page_size:'20'}); if(kw) params.set('keyword',kw); const r = await api.get<{total:number;list:R[]}>(`/admin/email-logs?${params}`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [kw])
  useEffect(() => { load() }, [load])
  return (<><Typography.Title level={4}>邮件日志</Typography.Title>
    <Card size="small" style={{marginBottom:16}}><Row gutter={12}><Col><Input placeholder="收件邮箱" prefix={<SearchOutlined />} value={kw} onChange={e=>setKw(e.target.value)} onPressEnter={()=>load()} allowClear style={{width:240}} /></Col><Col><Button type="primary" onClick={()=>load()}>查询</Button></Col></Row></Card>
    <Table dataSource={list} rowKey="id" pagination={{current:page,total,pageSize:20,onChange:p=>load(p),showTotal:t=>`共 ${t} 条`}} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'收件邮箱',dataIndex:'email',width:200,render:(v:string,r:R)=><a onClick={()=>setDetail(r)}>{v}</a>},
      {title:'标题',dataIndex:'title',ellipsis:true},{title:'状态',dataIndex:'status',width:80,render:(v:number)=>{const s=S[v];return s?<Tag color={s.c}>{s.t}</Tag>:v}},
      {title:'操作',width:60,render:(_:unknown,r:R)=><a style={{color:'red'}} onClick={async()=>{await api.del(`/admin/email-logs/${r.id}`);load(page)}}>删除</a>},
      {title:'时间',dataIndex:'created_at',width:170,render:(v:string)=>v?new Date(v).toLocaleString('zh-CN'):'-'},
    ]} />
    <DetailDrawer open={!!detail} onClose={()=>setDetail(null)} title={`邮件详情 #${detail?.id}`} items={detail?[
      {label:'收件邮箱',value:detail.email},{label:'标题',value:detail.title},{label:'状态',value:S[detail.status]?.t||detail.status},
      {label:'内容',value:<div dangerouslySetInnerHTML={{__html:detail.content||'-'}} />},
      {label:'时间',value:detail.created_at?new Date(detail.created_at).toLocaleString('zh-CN'):'-'},
    ]:[]} />
  </>)
}
