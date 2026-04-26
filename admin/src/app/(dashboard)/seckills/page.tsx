'use client'

import { useEffect, useState, useCallback } from 'react'
import {
  Table, Button, Modal, Form, Input, InputNumber, DatePicker, Typography, Card, Row, Col, Space, message,
} from 'antd'
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import dayjs from 'dayjs'

type PromoItem = { goods_id: number; sku_id: number; promo_price: number; promo_stock: number; per_limit?: number }
type SeckillRow = { id: number; name: string; start_time: string; end_time: string; status: number; items?: PromoItem[] }

/**
 * 秒杀活动管理（对接 /api/admin/seckills）。
 */
export default function SeckillsPage() {
  const [list, setList] = useState<SeckillRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()

  const load = useCallback(async (p = 1) => {
    const r = await api.get<{ total: number; list: SeckillRow[] }>(`/admin/seckills?page=${p}&page_size=20`)
    setList(r.list || [])
    setTotal(r.total || 0)
    setPage(p)
  }, [])

  useEffect(() => { load(1) }, [load])

  return (
    <>
      <Typography.Title level={4}>秒杀活动</Typography.Title>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row justify="end">
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              form.resetFields()
              form.setFieldsValue({
                items: [{ goods_id: undefined, sku_id: undefined, promo_price: 1, promo_stock: 1, per_limit: 1 }],
              })
              setOpen(true)
            }}
          >
            新建秒杀
          </Button>
        </Row>
      </Card>
      <Table
        dataSource={list}
        rowKey="id"
        pagination={{ current: page, total, pageSize: 20, onChange: p => load(p), showTotal: t => `共 ${t} 条` }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 64 },
          { title: '名称', dataIndex: 'name' },
          { title: '开始', dataIndex: 'start_time', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
          { title: '结束', dataIndex: 'end_time', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
          { title: '状态', dataIndex: 'status', width: 72, render: (v: number) => (v === 1 ? '启用' : '禁用') },
          { title: 'SKU数', width: 72, render: (_: unknown, r: SeckillRow) => r.items?.length ?? 0 },
        ]}
      />
      <Modal title="新建秒杀" open={open} onCancel={() => setOpen(false)} onOk={() => form.submit()} width={640} destroyOnClose>
        <Form
          form={form}
          layout="vertical"
          onFinish={async v => {
            const items = (v.items as Record<string, unknown>[]).map(x => ({
              goods_id: Number(x.goods_id),
              sku_id: Number(x.sku_id),
              promo_price: Number(x.promo_price),
              promo_stock: Number(x.promo_stock),
              per_limit: Number(x.per_limit ?? 0),
            }))
            await api.post('/admin/seckills', {
              name: v.name,
              start_time: (v.start_time as dayjs.Dayjs).format('YYYY-MM-DD HH:mm:ss'),
              end_time: (v.end_time as dayjs.Dayjs).format('YYYY-MM-DD HH:mm:ss'),
              items,
            })
            message.success('已创建')
            setOpen(false)
            load(page)
          }}
        >
          <Form.Item name="name" label="活动名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Row gutter={12}>
            <Col span={12}><Form.Item name="start_time" label="开始时间" rules={[{ required: true }]}><DatePicker showTime style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={12}><Form.Item name="end_time" label="结束时间" rules={[{ required: true }]}><DatePicker showTime style={{ width: '100%' }} /></Form.Item></Col>
          </Row>
          <Typography.Text type="secondary">活动商品（价格单位：分）</Typography.Text>
          <Form.List name="items">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...rest }) => (
                  <Space key={key} align="baseline" style={{ display: 'flex', marginBottom: 8 }} wrap>
                    <Form.Item {...rest} name={[name, 'goods_id']} rules={[{ required: true, message: '商品ID' }]}><InputNumber placeholder="商品ID" min={1} style={{ width: 100 }} /></Form.Item>
                    <Form.Item {...rest} name={[name, 'sku_id']} rules={[{ required: true, message: 'SKU' }]}><InputNumber placeholder="SKU ID" min={1} style={{ width: 100 }} /></Form.Item>
                    <Form.Item {...rest} name={[name, 'promo_price']} rules={[{ required: true }]}><InputNumber placeholder="秒杀价(分)" min={1} style={{ width: 120 }} /></Form.Item>
                    <Form.Item {...rest} name={[name, 'promo_stock']} rules={[{ required: true }]}><InputNumber placeholder="库存" min={1} style={{ width: 88 }} /></Form.Item>
                    <Form.Item {...rest} name={[name, 'per_limit']}><InputNumber placeholder="每人限购(0不限)" min={0} style={{ width: 130 }} /></Form.Item>
                    <MinusCircleOutlined onClick={() => remove(name)} />
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>添加 SKU</Button>
                </Form.Item>
              </>
            )}
          </Form.List>
        </Form>
      </Modal>
    </>
  )
}
