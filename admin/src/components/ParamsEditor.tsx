'use client'
import { Button, Input, Space } from 'antd'
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons'
import { useEffect, useState } from 'react'

interface Param { name: string; value: string }
interface ParamsEditorProps { value?: string; onChange?: (v: string) => void }

export default function ParamsEditor({ value, onChange }: ParamsEditorProps) {
  const [items, setItems] = useState<Param[]>([])

  useEffect(() => {
    try { const arr = JSON.parse(value || '[]'); if (Array.isArray(arr)) setItems(arr) } catch { /* ignore */ }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const emit = (next: Param[]) => { setItems(next); onChange?.(JSON.stringify(next)) }
  const update = (i: number, field: keyof Param, v: string) => {
    const next = [...items]; next[i] = { ...next[i], [field]: v }; emit(next)
  }

  return (
    <div>
      {items.map((item, i) => (
        <Space key={i} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
          <Input value={item.name} onChange={e => update(i, 'name', e.target.value)} placeholder="参数名" style={{ width: 160 }} />
          <Input value={item.value} onChange={e => update(i, 'value', e.target.value)} placeholder="参数值" style={{ width: 260 }} />
          <MinusCircleOutlined onClick={() => emit(items.filter((_, j) => j !== i))} style={{ color: '#999' }} />
        </Space>
      ))}
      <Button type="dashed" onClick={() => emit([...items, { name: '', value: '' }])} icon={<PlusOutlined />} block>
        添加参数
      </Button>
    </div>
  )
}
