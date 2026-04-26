'use client'
import { useEffect, useState } from 'react'
import { Table, Typography, Image, message } from 'antd'
import { api } from '@/lib/api'
type R = {id:number;name:string;path:string;size:number;ext:string;created_at:string}
export default function AttachmentPage() {
  const [list, setList] = useState<R[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const load = async (p = 1) => { const r = await api.get<{total:number;list:R[]}>(`/admin/attachments?page=${p}&page_size=20`); setList(r.list||[]); setTotal(r.total); setPage(p) }
  useEffect(() => { load() }, [])
  return (<>
    <Typography.Title level={4}>附件管理</Typography.Title>
    <Table dataSource={list} rowKey="id" pagination={{ current: page, total, pageSize: 20, onChange: p => load(p), showTotal: t => `共 ${t} 条` }}
      columns={[
        { title: 'ID', dataIndex: 'id', width: 60 },
        { title: '预览', dataIndex: 'path', width: 80, render: (v: string) => v && /\.(jpg|jpeg|png|gif|webp)$/i.test(v) ? <Image src={v} width={40} height={40} style={{objectFit:'cover'}} alt="" /> : '-' },
        { title: '文件名', dataIndex: 'name', ellipsis: true },
        { title: '大小', dataIndex: 'size', width: 100, render: (v: number) => v > 1048576 ? `${(v/1048576).toFixed(1)}MB` : `${(v/1024).toFixed(0)}KB` },
        { title: '类型', dataIndex: 'ext', width: 60 },
        { title: '时间', dataIndex: 'created_at', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
        { title: '操作', width: 60, render: (_: unknown, r: R) => <a style={{color:'red'}} onClick={async () => { await api.del(`/admin/attachments/${r.id}`); message.success('已删除'); load(page) }}>删除</a> },
      ]} />
  </>)
}
