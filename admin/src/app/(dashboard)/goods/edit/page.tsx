'use client'
import { useEffect, useState } from 'react'
import { Card, Form, Input, InputNumber, Select, Tabs, Button, Typography, Space, message } from 'antd'
import { useRouter, useSearchParams } from 'next/navigation'
import { api } from '@/lib/api'
import ImageUpload from '@/components/ImageUpload'
import SpecEditor from '@/components/SpecEditor'
import RichEditor from '@/components/RichEditor'

interface Cat { id: number; name: string; children?: Cat[] }

export default function GoodsEditPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const id = searchParams.get('id')
  const [form] = Form.useForm()
  const [cats, setCats] = useState<Cat[]>([])
  const [brands, setBrands] = useState<{value:number;label:string}[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => { api.get<Cat[]>('/categories').then(setCats) }, [])
  useEffect(() => { api.get<{id:number;name:string}[]>('/brands').then(r => setBrands((r||[]).map(b => ({value:b.id,label:b.name})))) }, [])
  useEffect(() => {
    if (id) {
      api.get<Record<string, unknown>>(`/goods/${id}`).then(data => {
        // Reconstruct spec_data from flat SKUs for SpecEditor
        const skus = data.skus as Array<{ name: string; specs: string; price: number; stock: number; coding?: string }> | undefined
        if (skus?.length && skus[0].specs) {
          // Parse spec types from SKU specs strings (e.g. "红色,XL")
          const specNames = (skus[0].name || skus[0].specs).split(',')
          // Try to infer type names from the pattern - use generic names
          const typeValues: string[][] = specNames.map(() => [] as string[])
          for (const sku of skus) {
            const vals = sku.specs.split(',')
            vals.forEach((v, i) => { if (v && !typeValues[i]?.includes(v)) typeValues[i]?.push(v) })
          }
          const types = typeValues.map((values, i) => ({ name: `规格${i + 1}`, values }))
          const skuRows = skus.map(s => ({ specs: s.specs, price: s.price, stock: s.stock, coding: s.coding || '' }))
          data.spec_data = { types, skus: skuRows }
          data.default_price = skus[0]?.price
          data.default_stock = skus[0]?.stock
        } else if (skus?.length) {
          data.default_price = skus[0].price
          data.default_stock = skus[0].stock
        }
        form.setFieldsValue(data)
      })
    }
  }, [id, form])

  const flatCats = (arr: Cat[], prefix = ''): { value: number; label: string }[] =>
    arr.flatMap(c => [{ value: c.id, label: prefix + c.name }, ...(c.children ? flatCats(c.children, prefix + c.name + '/') : [])])

  const onSave = async () => {
    const values = await form.validateFields()
    setLoading(true)
    try {
      // 把 SpecEditor 数据转成后端需要的 skus 数组
      const specData = values.spec_data
      let skus = []
      if (specData?.skus?.length) {
        skus = specData.skus.map((s: { specs: string; price: number; stock: number; coding: string }) => ({
          name: s.specs, price: s.price, stock: s.stock, specs: s.specs, coding: s.coding || '', image: '',
        }))
      }
      if (!skus.length) {
        // 无规格时创建默认SKU
        skus = [{ name: '默认', price: values.default_price || 0, stock: values.default_stock || 0, specs: '', image: '' }]
      }
      const payload = { ...values, skus, spec_data: undefined, default_price: undefined, default_stock: undefined }
      if (id) await api.put(`/admin/goods/${id}`, payload)
      else await api.post('/admin/goods', payload)
      message.success('保存成功'); router.push('/goods')
    } catch (e) { message.error((e as Error).message) }
    setLoading(false)
  }

  return (
    <>
      <Space style={{ marginBottom: 16 }}>
        <Button onClick={() => router.push('/goods')}>← 返回列表</Button>
        <Typography.Title level={4} style={{ margin: 0 }}>{id ? '编辑商品' : '新增商品'}</Typography.Title>
      </Space>

      <Form form={form} layout="vertical">
        <Tabs items={[
          { key: 'base', label: '基础信息', children: (
            <Card>
              <Form.Item name="title" label="商品标题" rules={[{ required: true }]}><Input /></Form.Item>
              <Form.Item name="subtitle" label="副标题"><Input /></Form.Item>
              <Form.Item name="category_id" label="商品分类" rules={[{ required: true }]}>
                <Select options={flatCats(cats)} placeholder="选择分类" showSearch optionFilterProp="label" />
              </Form.Item>
              <Form.Item name="main_image" label="主图"><ImageUpload /></Form.Item>
              <Form.Item name="images" label="商品图片(多图)"><ImageUpload max={9} /></Form.Item>
              <Form.Item name="brand_id" label="品牌"><Select options={brands} placeholder="选择品牌" allowClear showSearch optionFilterProp="label" /></Form.Item>
              <Form.Item name="sort" label="排序" initialValue={0}><InputNumber min={0} /></Form.Item>
              <Form.Item name="give_integral" label="赠送积分" initialValue={0}><InputNumber min={0} /></Form.Item>
              <Form.Item name="default_price" label="默认价格(分，无规格时使用)"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
              <Form.Item name="default_stock" label="默认库存(无规格时使用)" initialValue={0}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
            </Card>
          )},
          { key: 'spec', label: '商品规格', children: (
            <Card>
              <Form.Item name="spec_data" label="规格配置">
                <SpecEditor />
              </Form.Item>
            </Card>
          )},
          { key: 'params', label: '商品参数', children: (
            <Card>
              <Typography.Text type="secondary">商品参数将展示在商品详情页的参数Tab中</Typography.Text>
              <Form.Item name="params_json" label="参数(JSON格式)">
                <Input.TextArea rows={6} placeholder='[{"name":"品牌","value":"Apple"},{"name":"产地","value":"中国"}]' />
              </Form.Item>
            </Card>
          )},
          { key: 'photo', label: '相册/视频', children: (
            <Card>
              <Form.Item name="photos" label="商品相册"><ImageUpload max={20} /></Form.Item>
              <Form.Item name="video_url" label="视频URL"><Input placeholder="支持mp4格式" /></Form.Item>
            </Card>
          )},
          { key: 'app', label: 'APP详情', children: (
            <Card>
              <Form.Item name="content_app" label="APP端商品详情"><RichEditor height={400} /></Form.Item>
            </Card>
          )},
          { key: 'web', label: 'Web详情', children: (
            <Card>
              <Form.Item name="detail" label="Web端商品详情"><RichEditor height={400} /></Form.Item>
            </Card>
          )},
          { key: 'seo', label: 'SEO', children: (
            <Card>
              <Form.Item name="seo_title" label="SEO标题"><Input /></Form.Item>
              <Form.Item name="seo_keywords" label="SEO关键词"><Input /></Form.Item>
              <Form.Item name="seo_description" label="SEO描述"><Input.TextArea rows={3} /></Form.Item>
            </Card>
          )},
        ]} />

        <Card style={{ marginTop: 16, textAlign: 'center' }}>
          <Space>
            <Button onClick={() => router.push('/goods')}>取消</Button>
            <Button type="primary" loading={loading} onClick={onSave}>保存</Button>
          </Space>
        </Card>
      </Form>
    </>
  )
}
