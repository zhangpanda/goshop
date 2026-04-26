'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography } from 'antd'
import { api } from '@/lib/api'
export default function SearchHistoryPage() {
  const [list, setList] = useState<{id:number;user_id:number;keyword:string;created_at:string}[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const load = useCallback(async (p=1) => { const r = await api.get<{total:number;list:typeof list}>(`/admin/search-history?page=${p}&page_size=20`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [])
  useEffect(() => { load() }, [load])
  return (<><Typography.Title level={4}>搜索记录</Typography.Title>
    <Table dataSource={list} rowKey="id" pagination={{current:page,total,pageSize:20,onChange:p=>load(p)}} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'用户ID',dataIndex:'user_id',width:80},{title:'关键词',dataIndex:'keyword'},
      {title:'操作',width:80,render:(_:unknown,r:typeof list[0])=><a style={{color:'red'}} onClick={async()=>{await api.del(`/admin/search-history/${r.id}`);load(page)}}>删除</a>},
      {title:'时间',dataIndex:'created_at',width:180,render:(v:string)=>v?new Date(v).toLocaleString('zh-CN'):'-'},
    ]} /></>)
}
