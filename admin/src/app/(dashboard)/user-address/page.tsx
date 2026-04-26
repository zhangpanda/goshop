'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography } from 'antd'
import { api } from '@/lib/api'
import { useUserMap } from '@/lib/useIdMap'
export default function UserAddressPage() {
  const [list, setList] = useState<{id:number;user_id:number;name:string;phone:string;province:string;city:string;district:string;detail:string;is_default:boolean}[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const load = useCallback(async (p=1) => { const r = await api.get<{total:number;list:typeof list}>(`/admin/user-address?page=${p}&page_size=20`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [])
  useEffect(() => { load() }, [load])
  const userMap = useUserMap(list.map(r => r.user_id))
  return (<><Typography.Title level={4}>用户地址</Typography.Title>
    <Table dataSource={list} rowKey="id" pagination={{current:page,total,pageSize:20,onChange:p=>load(p)}} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'用户',dataIndex:'user_id',width:120,render:(v:number)=>userMap[v]||`#${v}`},{title:'姓名',dataIndex:'name',width:80},
      {title:'手机',dataIndex:'phone',width:120},{title:'地区',key:'area',render:(_:unknown,r:typeof list[0])=>`${r.province}${r.city}${r.district}`},
      {title:'详细地址',dataIndex:'detail',ellipsis:true},{title:'默认',dataIndex:'is_default',width:60,render:(v:boolean)=>v?'是':'否'},
      {title:'操作',width:60,render:(_:unknown,r:typeof list[0])=><a style={{color:'red'}} onClick={async()=>{await api.del(`/admin/user-address/${r.id}`);load(page)}}>删除</a>},
    ]} /></>)
}
