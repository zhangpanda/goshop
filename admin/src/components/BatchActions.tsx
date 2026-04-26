'use client'
import { Button, Space, Popconfirm, message } from 'antd'
import { DeleteOutlined, ExportOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'

interface Props {
  selectedIds: number[]
  deleteUrl?: string
  statusUrl?: string
  exportUrl?: string
  onDone: () => void
}

export default function BatchActions({ selectedIds, deleteUrl, statusUrl, exportUrl, onDone }: Props) {
  if (!selectedIds.length) return null
  const count = selectedIds.length

  const batchDelete = async () => {
    if (!deleteUrl) return
    for (const id of selectedIds) await api.del(`${deleteUrl}/${id}`)
    message.success(`已删除 ${count} 条`); onDone()
  }

  const batchStatus = async (status: number) => {
    if (!statusUrl) return
    for (const id of selectedIds) await api.put(`${statusUrl}/${id}/status`, { status })
    message.success(`已更新 ${count} 条`); onDone()
  }

  const doExport = async () => {
    if (!exportUrl) return
    try {
      const res = await api.post<{ url: string }>(exportUrl, { ids: selectedIds })
      window.open(res.url)
    } catch { message.info('导出功能需要后端支持') }
  }

  return (
    <Space style={{ marginBottom: 12 }}>
      <span>已选 {count} 项</span>
      {deleteUrl && <Popconfirm title={`确认删除 ${count} 条?`} onConfirm={batchDelete}><Button danger size="small" icon={<DeleteOutlined />}>批量删除</Button></Popconfirm>}
      {statusUrl && <><Button size="small" onClick={() => batchStatus(1)}>批量启用</Button><Button size="small" onClick={() => batchStatus(0)}>批量禁用</Button></>}
      {exportUrl && <Button size="small" icon={<ExportOutlined />} onClick={doExport}>导出</Button>}
    </Space>
  )
}
