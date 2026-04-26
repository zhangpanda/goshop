'use client'
import { Button, Input, Select, Space, Switch } from 'antd'
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons'
import { useEffect, useState } from 'react'

interface ConfigItem { key: string; value: string; type: 'string' | 'number' | 'boolean' | 'color' }
interface Props { value?: string; onChange?: (v: string) => void }

export default function JsonConfigEditor({ value, onChange }: Props) {
  const [items, setItems] = useState<ConfigItem[]>([])

  useEffect(() => {
    try {
      const obj = JSON.parse(value || '{}')
      if (typeof obj === 'object' && !Array.isArray(obj)) {
        setItems(Object.entries(obj).map(([k, v]) => ({
          key: k, value: String(v),
          type: typeof v === 'boolean' ? 'boolean' : typeof v === 'number' ? 'number' : (String(v).match(/^#[0-9a-fA-F]{3,8}$/) ? 'color' : 'string'),
        })))
      }
    } catch { /* ignore */ }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const emit = (next: ConfigItem[]) => {
    setItems(next)
    const obj: Record<string, unknown> = {}
    for (const item of next) {
      if (!item.key) continue
      if (item.type === 'boolean') obj[item.key] = item.value === 'true'
      else if (item.type === 'number') obj[item.key] = Number(item.value) || 0
      else obj[item.key] = item.value
    }
    onChange?.(JSON.stringify(obj, null, 2))
  }

  const update = (i: number, field: string, v: string) => {
    const next = [...items]; next[i] = { ...next[i], [field]: v }; emit(next)
  }

  return (
    <div>
      {items.map((item, i) => (
        <Space key={i} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
          <Input value={item.key} onChange={e => update(i, 'key', e.target.value)} placeholder="键名" style={{ width: 140 }} />
          <Select value={item.type} onChange={v => update(i, 'type', v)} style={{ width: 90 }}
            options={[{ value: 'string', label: '文本' }, { value: 'number', label: '数字' }, { value: 'boolean', label: '开关' }, { value: 'color', label: '颜色' }]} />
          {item.type === 'boolean' ? (
            <Switch checked={item.value === 'true'} onChange={v => update(i, 'value', String(v))} />
          ) : item.type === 'color' ? (
            <Input type="color" value={item.value || '#000000'} onChange={e => update(i, 'value', e.target.value)} style={{ width: 60, padding: 2 }} />
          ) : (
            <Input value={item.value} onChange={e => update(i, 'value', e.target.value)} placeholder="值" style={{ width: 220 }}
              type={item.type === 'number' ? 'number' : 'text'} />
          )}
          <MinusCircleOutlined onClick={() => emit(items.filter((_, j) => j !== i))} style={{ color: '#999' }} />
        </Space>
      ))}
      <Button type="dashed" onClick={() => emit([...items, { key: '', value: '', type: 'string' }])} icon={<PlusOutlined />} block>
        添加配置项
      </Button>
    </div>
  )
}
