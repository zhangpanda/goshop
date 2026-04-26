'use client'
import { useState, useEffect } from 'react'
import { Card, Input, InputNumber, Button, Tag, Space, Table, Typography, message } from 'antd'
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons'

interface SpecType { name: string; values: string[] }
interface SKURow { specs: string; price: number; stock: number; coding: string }
interface Props { value?: { types: SpecType[]; skus: SKURow[] }; onChange?: (v: { types: SpecType[]; skus: SKURow[] }) => void }

// 笛卡尔积
function cartesian(arrays: string[][]): string[][] {
  if (!arrays.length) return [[]]
  return arrays.reduce((acc, arr) => acc.flatMap(a => arr.map(v => [...a, v])), [[]] as string[][])
}

export default function SpecEditor({ value, onChange }: Props) {
  const [types, setTypes] = useState<SpecType[]>(value?.types || [])
  const [skus, setSkus] = useState<SKURow[]>(value?.skus || [])
  const [newType, setNewType] = useState('')
  const [newValues, setNewValues] = useState<Record<number, string>>({})

  // 规格变化时自动重新组合SKU
  useEffect(() => {
    if (!types.length || types.some(t => !t.values.length)) { setSkus([]); return }
    const combos = cartesian(types.map(t => t.values))
    setSkus(prev => {
      return combos.map(combo => {
        const specs = combo.join(',')
        const existing = prev.find(s => s.specs === specs)
        return existing || { specs, price: 0, stock: 0, coding: '' }
      })
    })
  }, [types])

  // 通知父组件
  useEffect(() => { onChange?.({ types, skus }) }, [types, skus])

  const addType = () => {
    if (!newType.trim()) return
    if (types.find(t => t.name === newType.trim())) { message.warning('规格类型已存在'); return }
    setTypes([...types, { name: newType.trim(), values: [] }])
    setNewType('')
  }

  const removeType = (i: number) => setTypes(types.filter((_, idx) => idx !== i))

  const addValue = (ti: number) => {
    const v = (newValues[ti] || '').trim()
    if (!v) return
    if (types[ti].values.includes(v)) { message.warning('规格值已存在'); return }
    const nt = [...types]; nt[ti] = { ...nt[ti], values: [...nt[ti].values, v] }
    setTypes(nt); setNewValues({ ...newValues, [ti]: '' })
  }

  const removeValue = (ti: number, vi: number) => {
    const nt = [...types]; nt[ti] = { ...nt[ti], values: nt[ti].values.filter((_, i) => i !== vi) }
    setTypes(nt)
  }

  const updateSku = (idx: number, field: keyof SKURow, val: string | number) => {
    const ns = [...skus]; ns[idx] = { ...ns[idx], [field]: val }; setSkus(ns)
  }

  const batchSet = (field: 'price' | 'stock', val: number) => {
    setSkus(skus.map(s => ({ ...s, [field]: val })))
  }

  return (
    <div>
      {/* 规格类型列表 */}
      {types.map((t, ti) => (
        <Card key={ti} size="small" style={{ marginBottom: 12 }}
          title={<Space><span>规格: {t.name}</span><Button type="link" danger size="small" icon={<DeleteOutlined />} onClick={() => removeType(ti)}>删除</Button></Space>}>
          <div style={{ marginBottom: 8 }}>
            {t.values.map((v, vi) => (
              <Tag key={vi} closable onClose={() => removeValue(ti, vi)} style={{ marginBottom: 4 }}>{v}</Tag>
            ))}
          </div>
          <Space>
            <Input size="small" placeholder="添加规格值" value={newValues[ti] || ''} style={{ width: 140 }}
              onChange={e => setNewValues({ ...newValues, [ti]: e.target.value })}
              onPressEnter={() => addValue(ti)} />
            <Button size="small" onClick={() => addValue(ti)} icon={<PlusOutlined />}>添加</Button>
          </Space>
        </Card>
      ))}

      {/* 添加规格类型 */}
      <Space style={{ marginBottom: 16 }}>
        <Input placeholder="规格类型名称(如颜色/尺码)" value={newType} onChange={e => setNewType(e.target.value)}
          onPressEnter={addType} style={{ width: 200 }} />
        <Button type="dashed" icon={<PlusOutlined />} onClick={addType}>添加规格类型</Button>
      </Space>

      {/* SKU矩阵表格 */}
      {skus.length > 0 && (
        <>
          <div style={{ marginBottom: 8 }}>
            <Typography.Text strong>SKU列表 ({skus.length}个组合)</Typography.Text>
            <Space style={{ marginLeft: 16 }}>
              <span>批量设置:</span>
              <InputNumber size="small" placeholder="价格(分)" min={0} style={{ width: 110 }}
                onPressEnter={e => batchSet('price', Number((e.target as HTMLInputElement).value))} />
              <InputNumber size="small" placeholder="库存" min={0} style={{ width: 90 }}
                onPressEnter={e => batchSet('stock', Number((e.target as HTMLInputElement).value))} />
            </Space>
          </div>
          <Table dataSource={skus} rowKey="specs" pagination={false} size="small"
            columns={[
              { title: '规格组合', dataIndex: 'specs', render: (v: string) => v.split(',').map((s, i) => <Tag key={i}>{types[i]?.name}: {s}</Tag>) },
              { title: '价格(分)', dataIndex: 'price', width: 120, render: (v: number, _: SKURow, i: number) => <InputNumber size="small" value={v} min={0} style={{ width: 100 }} onChange={val => updateSku(i, 'price', val || 0)} /> },
              { title: '库存', dataIndex: 'stock', width: 100, render: (v: number, _: SKURow, i: number) => <InputNumber size="small" value={v} min={0} style={{ width: 80 }} onChange={val => updateSku(i, 'stock', val || 0)} /> },
              { title: '编码', dataIndex: 'coding', width: 140, render: (v: string, _: SKURow, i: number) => <Input size="small" value={v} style={{ width: 120 }} onChange={e => updateSku(i, 'coding', e.target.value)} /> },
            ]}
          />
        </>
      )}

      {!types.length && <Typography.Text type="secondary">点击"添加规格类型"开始配置商品规格，如：颜色（红/蓝）、尺码（S/M/L）</Typography.Text>}
    </div>
  )
}
