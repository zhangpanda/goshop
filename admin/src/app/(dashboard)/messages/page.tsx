'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography, Tag, Card, Row, Col, Input, Button } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import DetailDrawer from '@/components/DetailDrawer'
type R = {id:number;user_id:number;title:string;content:string;type:string;is_read:boolean;created_at:string}
export default function MessagesPage() {
  const [list, setList] = useState<R[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [kw, setKw] = useState(''); const [detail, setDetail] = useState<R|null>(null)
  const load = useCallback(async (p=1) => { const params = new URLSearchParams({page:String(p),page_size:'20'}); if(kw) params.set('keyword',kw); const r = await api.get<{total:number;list:R[]}>(`/admin/messages?${params}`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [kw])
  useEffect(() => { load() }, [load])
  return (<><Typography.Title level={4}>消息管理</Typography.Title>
    <Card size="small" style={{marginBottom:16}}><Row gutter={12}><Col><Input placeholder="标题/用户ID" prefix={<SearchOutlined />} value={kw} onChange={e=>setKw(e.target.value)} onPressEnter={()=>load()} allowClear style={{width:220}} /></Col><Col><Button type="primary" onClick={()=>load()}>查询</Button></Col></Row></Card>
    <Table dataSource={list} rowKey="id" pagination={{current:page,total,pageSize:20,onChange:p=>load(p),showTotal:t=>`共 ${t} 条`}} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'用户ID',dataIndex:'user_id',width:80},
      {title:'标题',dataIndex:'title',render:(v:string,r:R)=><a onClick={()=>setDetail(r)}>{v}</a>},
      {title:'类型',dataIndex:'type',width:80},{title:'已读',dataIndex:'is_read',width:60,render:(v:boolean)=>v?<Tag color="green">是</Tag>:<Tag>否</Tag>},
      {title:'操作',width:80,render:(_:unknown,r:R)=><a style={{color:'red'}} onClick={async()=>{await api.del(`/admin/messages/${r.id}`);load(page)}}>删除</a>},
      {title:'时间',dataIndex:'created_at',width:170,render:(v:string)=>v?new Date(v).toLocaleString('zh-CN'):'-'},
    ]} />
    <DetailDrawer open={!!detail} onClose={()=>setDetail(null)} title={`消息详情 #${detail?.id}`} items={detail?[
      {label:'用户ID',value:detail.user_id},{label:'标题',value:detail.title},{label:'类型',value:detail.type},
      {label:'已读',value:detail.is_read?'是':'否'},{label:'内容',value:detail.content||'-'},
      {label:'时间',value:detail.created_at?new Date(detail.created_at).toLocaleString('zh-CN'):'-'},
    ]:[]} />
  </>)
}
