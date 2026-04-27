'use client'

import { useEffect, useRef, useState } from 'react'
import { Form, Select } from 'antd'
import type { FormListFieldData } from 'antd/es/form'
import { api } from '@/lib/api'

type Option = { value: number; label: string }

interface PromoGoodsSkuFieldsProps {
  /**
   * Form.List 根字段名（秒杀/拼团为 items）。
   */
  listName?: string
  /** Form.List 行下标 */
  rowName: number
  /** Form.List 解构出的 field 透传（除 key、name 外） */
  fieldRest: Partial<Omit<FormListFieldData, 'key' | 'name'>>
  /** 商品下拉宽度 */
  goodsSelectWidth?: number
}

/**
 * 秒杀/拼团等活动：按标题搜索选商品，再选 SKU，写入 goods_id、sku_id。
 */
export function PromoGoodsSkuFields({
  listName = 'items',
  rowName,
  fieldRest,
  goodsSelectWidth = 280,
}: PromoGoodsSkuFieldsProps) {
  const form = Form.useFormInstance()
  const goodsId = Form.useWatch([listName, rowName, 'goods_id']) as number | undefined
  const [goodsOpts, setGoodsOpts] = useState<Option[]>([])
  const [skuOpts, setSkuOpts] = useState<Option[]>([])
  const lastGoodsRef = useRef<number | undefined>(undefined)
  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  const fetchGoods = async (q: string) => {
    const params = new URLSearchParams({ page: '1', page_size: '30' })
    if (q.trim()) params.set('keyword', q.trim())
    const res = await api.get<{ list: { id: number; title: string }[] }>(`/goods?${params}`)
    setGoodsOpts((res.list || []).map(g => ({ value: g.id, label: `#${g.id} ${g.title}` })))
  }

  const scheduleGoodsSearch = (q: string) => {
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
    searchTimerRef.current = setTimeout(() => {
      void fetchGoods(q)
    }, 300)
  }

  useEffect(() => () => {
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
  }, [])

  useEffect(() => {
    if (!goodsId) {
      setSkuOpts([])
      lastGoodsRef.current = undefined
      form.setFieldValue([listName, rowName, 'sku_id'], undefined)
      return
    }
    const changed = lastGoodsRef.current !== goodsId
    lastGoodsRef.current = goodsId
    if (changed) form.setFieldValue([listName, rowName, 'sku_id'], undefined)

    let cancelled = false
    void api.get<{ skus?: { id: number; name: string }[] }>(`/goods/${goodsId}`).then(d => {
      if (cancelled) return
      setSkuOpts((d.skus || []).map(s => ({ value: s.id, label: `#${s.id} ${s.name || '默认规格'}` })))
    })
    return () => {
      cancelled = true
    }
  }, [goodsId, form, listName, rowName])

  return (
    <>
      <Form.Item
        {...fieldRest}
        name={[rowName, 'goods_id']}
        rules={[{ required: true, message: '请选择商品' }]}
      >
        <Select
          showSearch
          filterOption={false}
          placeholder="搜索商品名称"
          style={{ width: goodsSelectWidth }}
          options={goodsOpts}
          onSearch={scheduleGoodsSearch}
          onFocus={() => { void fetchGoods('') }}
          allowClear
        />
      </Form.Item>
      <Form.Item
        {...fieldRest}
        name={[rowName, 'sku_id']}
        rules={[{ required: true, message: '请选择 SKU' }]}
      >
        <Select
          placeholder="选择 SKU"
          style={{ width: 200 }}
          options={skuOpts}
          disabled={!goodsId}
          allowClear
        />
      </Form.Item>
    </>
  )
}
