'use client'
import { Button, message } from 'antd'
import { ExportOutlined } from '@ant-design/icons'

interface Props { type: 'orders' | 'users' | 'goods'; label?: string }

export default function ExportButton({ type, label }: Props) {
  const doExport = async () => {
    const token = localStorage.getItem('admin_token')
    try {
      const res = await fetch('/api/admin/export', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
        body: JSON.stringify({ type }),
      })
      if (!res.ok) { message.error('导出失败'); return }
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url; a.download = `export_${type}_${new Date().toISOString().slice(0, 10)}.csv`
      a.click(); URL.revokeObjectURL(url)
      message.success('导出成功')
    } catch { message.error('导出失败') }
  }

  return <Button icon={<ExportOutlined />} onClick={doExport}>{label || '导出'}</Button>
}
