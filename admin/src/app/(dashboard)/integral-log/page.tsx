'use client'
import { useEffect, useState, useCallback } from 'react'
import { Table, Typography } from 'antd'
import { api } from '@/lib/api'
export default function IntegralLogPage() {
  const [list, setList] = useState<{id:number;user_id:number;points:number;balance:number;type:string;remark:string;created_at:string}[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const load = useCallback(async (p = 1) => { const r = await api.get<{total:number;list:typeof list}>(`/admin/integral-logs?page=${p}&page_size=20`); setList(r.list||[]); setTotal(r.total); setPage(p) }, [])
  useEffect(() => { load() }, [load])
  return (<>
    <Typography.Title level={4}>积分日志</Typography.Title>
    <Table dataSource={list} rowKey="id" pagination={{ current: page, total, pageSize: 20, onChange: p => load(p) }}
      columns={[
        { title: 'ID', dataIndex: 'id', width: 60 },
        { title: '用户ID', dataIndex: 'user_id', width: 80 },
        { title: '变动', dataIndex: 'points', width: 80 },
        { title: '余额', dataIndex: 'balance', width: 80 },
        { title: '类型', dataIndex: 'type', width: 100 },
        { title: '备注', dataIndex: 'remark', ellipsis: true },
        { title: '时间', dataIndex: 'created_at', width: 180, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
      ]} />
  </>)
}
