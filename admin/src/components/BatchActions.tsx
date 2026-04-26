'use client'
import { Button, Space, Popconfirm, message } from 'antd'
import { DeleteOutlined, ExportOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'

interface Props {
  selectedIds: number[]
  deleteUrl?: string
  statusUrl?: string
  /** 与 exportType 配合：POST /admin/export 下载 CSV（勾选行） */
  exportUrl?: string
  exportType?: 'orders' | 'users' | 'goods'
  /** 批量删除按钮文案，默认「批量删除」 */
  deleteButtonText?: string
  /** Popconfirm 标题，默认「确认删除 N 条?」 */
  deleteConfirmTitle?: string
  onDone: () => void
}

export default function BatchActions({ selectedIds, deleteUrl, statusUrl, exportUrl, exportType, deleteButtonText, deleteConfirmTitle, onDone }: Props) {
  if (!selectedIds.length) return null
  const count = selectedIds.length
  const delTitle = deleteConfirmTitle ?? `确认删除 ${count} 条?`
  const delBtn = deleteButtonText ?? '批量删除'

  const batchDelete = async () => {
    if (!deleteUrl) return
    for (const id of selectedIds) await api.del(`${deleteUrl}/${id}`)
    message.success(deleteButtonText ? `已处理 ${count} 条` : `已删除 ${count} 条`)
    onDone()
  }

  const batchStatus = async (status: number) => {
    if (!statusUrl) return
    for (const id of selectedIds) await api.put(`${statusUrl}/${id}/status`, { status })
    message.success(`已更新 ${count} 条`); onDone()
  }

  const doExport = async () => {
    if (!exportUrl || !exportType) return
    const token = typeof window !== 'undefined' ? localStorage.getItem('admin_token') : null
    try {
      const res = await fetch(`/api${exportUrl}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({ type: exportType, ids: selectedIds }),
      })
      if (!res.ok) {
        const j = await res.json().catch(() => ({}))
        message.error((j as { msg?: string }).msg || '导出失败')
        return
      }
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `export_${exportType}_${selectedIds.length}rows_${new Date().toISOString().slice(0, 10)}.csv`
      a.click()
      URL.revokeObjectURL(url)
      message.success('导出成功')
    } catch {
      message.error('导出失败')
    }
  }

  return (
    <Space style={{ marginBottom: 12 }}>
      <span>已选 {count} 项</span>
      {deleteUrl && <Popconfirm title={delTitle} onConfirm={batchDelete}><Button danger size="small" icon={<DeleteOutlined />}>{delBtn}</Button></Popconfirm>}
      {statusUrl && <><Button size="small" onClick={() => batchStatus(1)}>批量启用</Button><Button size="small" onClick={() => batchStatus(0)}>批量禁用</Button></>}
      {exportUrl && exportType && <Button size="small" icon={<ExportOutlined />} onClick={doExport}>导出选中</Button>}
    </Space>
  )
}
