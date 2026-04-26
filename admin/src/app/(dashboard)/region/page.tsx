'use client'
import { useEffect, useState } from 'react'
import { Table, Typography, Select } from 'antd'
import { api } from '@/lib/api'

interface Region { id: number; name: string; parent_id: number; level: number }

export default function RegionPage() {
  const [list, setList] = useState<Region[]>([])
  const [pid, setPid] = useState(0)
  const load = (p: number) => { setPid(p); api.get<Region[]>(`/regions?parent_id=${p}`).then(r => setList(Array.isArray(r) ? r : [])) }
  useEffect(() => { load(0) }, [])
  return (
    <>
      <Typography.Title level={4}>地区管理</Typography.Title>
      <Select style={{ width: 200, marginBottom: 16 }} value={pid} onChange={load}
        options={[{ value: 0, label: '顶级(省)' }, ...list.map(r => ({ value: r.id, label: r.name }))]} />
      <Table dataSource={list} rowKey="id" pagination={false}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '名称', dataIndex: 'name' },
          { title: '层级', dataIndex: 'level', width: 80, render: (v: number) => ['', '省', '市', '区'][v] || v },
          { title: '操作', width: 80, render: (_: unknown, r: Region) => r.level < 3 ? <a onClick={() => load(r.id)}>下级</a> : '-' },
        ]}
      />
    </>
  )
}
