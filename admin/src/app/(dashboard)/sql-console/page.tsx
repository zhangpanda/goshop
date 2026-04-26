'use client'
import { useState } from 'react'
import { Typography, Input, Button, Table, Card, message } from 'antd'
import { api } from '@/lib/api'
export default function SqlConsolePage() {
  const [sql, setSql] = useState('')
  const [result, setResult] = useState<{columns:string[];rows:Record<string,unknown>[]}|null>(null)
  const exec = async () => {
    try { const r = await api.post<{columns:string[];rows:Record<string,unknown>[]}>('/admin/sql-console', { sql }); setResult(r) }
    catch (e: unknown) { message.error((e as Error).message) }
  }
  return (<><Typography.Title level={4}>SQL控制台</Typography.Title>
    <Card><Input.TextArea rows={4} value={sql} onChange={e=>setSql(e.target.value)} placeholder="输入SELECT语句..." />
    <Button type="primary" style={{marginTop:8}} onClick={exec}>执行</Button></Card>
    {result && <Table style={{marginTop:16}} dataSource={result.rows} rowKey={(_, i) => String(i)} pagination={{pageSize:50}}
      columns={result.columns.map(c => ({ title: c, dataIndex: c, ellipsis: true }))} scroll={{x:'max-content'}} />}
  </>)
}
