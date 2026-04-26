'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography, Select } from 'antd'
import { api } from '@/lib/api'
import { useGoodsMap } from '@/lib/useIdMap'
export default function WarehouseGoodsPage() {
  const [whs, setWhs] = useState<{id:number;name:string}[]>([])
  const [whId, setWhId] = useState<number|undefined>()
  const [list, setList] = useState<{id:number;goods_id:number;inventory:number;is_enable:number}[]>([])
  useEffect(() => { api.get<{id:number;name:string}[]>('/admin/warehouses').then(r => { const a = Array.isArray(r)?r:[]; setWhs(a); if(a.length) { setWhId(a[0].id) } }) }, [])
  const load = useCallback(async () => { if(!whId) return; const r = await api.get<typeof list>(`/admin/warehouses/${whId}/goods`); setList(Array.isArray(r)?r:[]) }, [whId])
  useEffect(() => { load() }, [load])
  const goodsMap = useGoodsMap(list.map(r => r.goods_id))
  return (<><Typography.Title level={4}>仓库商品管理</Typography.Title>
    <Select style={{width:200,marginBottom:16}} value={whId} onChange={setWhId} placeholder="选择仓库" options={whs.map(w=>({value:w.id,label:w.name}))} />
    <Table dataSource={list} rowKey="id" pagination={false} columns={[
      {title:'ID',dataIndex:'id',width:60},{title:'商品',dataIndex:'goods_id',render:(v:number)=>goodsMap[v]||`#${v}`},{title:'库存',dataIndex:'inventory',width:80},
      {title:'启用',dataIndex:'is_enable',width:60,render:(v:number)=>v?'是':'否'},
      {title:'操作',width:60,render:(_:unknown,r:typeof list[0])=><a style={{color:'red'}} onClick={async()=>{await api.del(`/admin/warehouse-goods/${r.id}`);load()}}>删除</a>},
    ]} /></>)
}
