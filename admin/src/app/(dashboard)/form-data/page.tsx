'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography, Card, Row, Col, Input, Button } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import { useUserMap } from '@/lib/useIdMap'
import DetailDrawer from '@/components/DetailDrawer'
type R = {id:number;form_id:number;user_id:number;data:string;created_at:string}
export default function FormDataPage() {
  const [list, setList] = useState<R[]>([]); const [total, setTotal] = useState(0); const [page, setPage] = useState(1)
  const [kw, setKw] = useState(''); const [detail, setDetail] = useState<R|null>(null)
  const userMap = useUserMap(list.map(r => r.user_id))
  const load = useCallback(async (p=1) => { const params = new URLSearchParams({page:String(p),page_size:'20'}); if(kw) params.set('keyword',kw); const r = await api.get<{total:number;list:R[]}>(`/admin/form-data?${params}`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [kw])
  useEffect(() => { load() }, [load])
  return (<><Typography.Title level={4}>表单数据</Typography.Title>
    <Card size="small" style={{marginBottom:16}}><Row gutter={12}><Col><Input placeholder="表单ID" prefix={<SearchOutlined />} value={kw} onChange={e=>setKw(e.target.value)} onPressEnter={()=>load()} allowClear style={{width:200}} /></Col><Col><Button type="primary" onClick={()=>load()}>查询</Button></Col></Row></Card>
    <Table dataSource={list} rowKey="id" pagination={{current:page,total,pageSize:20,onChange:p=>load(p),showTotal:t=>`共 ${t} 条`}} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'表单ID',dataIndex:'form_id',width:80,render:(v:number,r:R)=><a onClick={()=>setDetail(r)}>{v}</a>},
      {title:'用户',dataIndex:'user_id',width:120,render:(v:number)=>userMap[v]||`#${v}`},{title:'数据',dataIndex:'data',ellipsis:true,render:(v:string)=>{try{const d=JSON.parse(v||'{}');const keys=Object.keys(d);return keys.slice(0,3).map(k=>`${k}:${typeof d[k]==='object'?'...':d[k]}`).join(' | ')+(keys.length>3?' ...':'')}catch{return v}}},
      {title:'操作',width:60,render:(_:unknown,r:R)=><a style={{color:'red'}} onClick={async()=>{await api.del(`/admin/form-data/${r.id}`);load(page)}}>删除</a>},
      {title:'时间',dataIndex:'created_at',width:170,render:(v:string)=>v?new Date(v).toLocaleString('zh-CN'):'-'},
    ]} />
    <DetailDrawer open={!!detail} onClose={()=>setDetail(null)} title={`表单数据 #${detail?.id}`} items={detail?[
      {label:'表单ID',value:detail.form_id},{label:'用户',value:userMap[detail.user_id]||`#${detail.user_id}`},
      {label:'提交数据',value:<pre style={{whiteSpace:'pre-wrap',background:'#f5f5f5',padding:8,borderRadius:4,fontSize:12}}>{JSON.stringify(JSON.parse(detail.data||'{}'),null,2)}</pre>},
      {label:'时间',value:detail.created_at?new Date(detail.created_at).toLocaleString('zh-CN'):'-'},
    ]:[]} />
  </>)
}
