'use client'

import { useEffect, useState, useCallback } from 'react'
import {
  Table, Button, Modal, Form, Input, InputNumber, DatePicker, Typography, Card, Row, Col, Space, message,
} from 'antd'
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons'
import { api } from '@/lib/api'
import dayjs from 'dayjs'
import { PromoGoodsSkuFields } from '@/components/PromoGoodsSkuFields'

type PromoItem = { goods_id: number; sku_id: number; promo_price: number; promo_stock: number }
type GroupRow = {
  id: number
  name: string
  start_time: string
  end_time: string
  status: number
  group_size?: number
  group_time?: number
  items?: PromoItem[]
}

/**
 * 拼团活动管理（对接 /api/admin/group-buys）。
 */
export default function GroupBuysPage() {
  const [list, setList] = useState<GroupRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()

  const load = useCallback(async (p = 1) => {
    const r = await api.get<{ total: number; list: GroupRow[] }>(`/admin/group-buys?page=${p}&page_size=20`)
    setList(r.list || [])
    setTotal(r.total || 0)
    setPage(p)
  }, [])

  useEffect(() => { load(1) }, [load])

  return (
    <>
      <Typography.Title level={4}>拼团活动</Typography.Title>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row justify="end">
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>
            新建拼团
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
          { title: '成团人数', dataIndex: 'group_size', width: 88 },
          { title: '限时(分)', dataIndex: 'group_time', width: 88 },
          { title: '开始', dataIndex: 'start_time', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
          { title: '结束', dataIndex: 'end_time', width: 170, render: (v: string) => v ? new Date(v).toLocaleString('zh-CN') : '-' },
          { title: 'SKU数', width: 72, render: (_: unknown, r: GroupRow) => r.items?.length ?? 0 },
        ]}
      />
      <Modal
        title="新建拼团"
        open={open}
        onCancel={() => setOpen(false)}
        onOk={() => form.submit()}
        width={640}
        forceRender
        afterOpenChange={visible => {
          if (!visible) return
          form.resetFields()
          form.setFieldsValue({
            group_size: 2,
            group_time: 1440,
            items: [{ goods_id: undefined, sku_id: undefined, promo_price: 1, promo_stock: 100 }],
          })
        }}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={async v => {
            const items = (v.items as Record<string, unknown>[]).map(x => ({
              goods_id: Number(x.goods_id),
              sku_id: Number(x.sku_id),
              promo_price: Number(x.promo_price),
              promo_stock: Number(x.promo_stock),
            }))
            await api.post('/admin/group-buys', {
              name: v.name,
              start_time: (v.start_time as dayjs.Dayjs).format('YYYY-MM-DD HH:mm:ss'),
              end_time: (v.end_time as dayjs.Dayjs).format('YYYY-MM-DD HH:mm:ss'),
              group_size: Number(v.group_size),
              group_time: Number(v.group_time),
              items,
            })
            message.success('已创建')
            setOpen(false)
            load(page)
          }}
        >
          <Form.Item name="name" label="活动名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Row gutter={12}>
            <Col span={8}><Form.Item name="group_size" label="成团人数" rules={[{ required: true }]}><InputNumber min={2} style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={8}><Form.Item name="group_time" label="成团时限(分钟)" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Col>
          </Row>
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
                    <PromoGoodsSkuFields rowName={name} fieldRest={rest} />
                    <Form.Item {...rest} name={[name, 'promo_price']} rules={[{ required: true }]}><InputNumber placeholder="拼团价(分)" min={1} style={{ width: 120 }} /></Form.Item>
                    <Form.Item {...rest} name={[name, 'promo_stock']} rules={[{ required: true }]}><InputNumber placeholder="库存" min={1} style={{ width: 88 }} /></Form.Item>
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
