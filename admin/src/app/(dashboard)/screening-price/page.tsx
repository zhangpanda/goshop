'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, InputNumber } from 'antd'

export default function ScreeningPricePage() {
  return <CrudPage title="筛选价格" listUrl="/admin/screening-prices" createUrl="/admin/screening-prices"
    searchClient searchPlaceholder="名称"
    deleteUrl={r => `/admin/screening-prices/${r.id}`} updateUrl={r => `/admin/screening-prices/${r.id}`} batchDelete
    columns={[
      { title: 'ID', dataIndex: 'id', width: 60 },
      { title: '名称', dataIndex: 'name' },
      { title: '最低价(分)', dataIndex: 'min_price', width: 120 },
      { title: '最高价(分)', dataIndex: 'max_price', width: 120 },
      { title: '排序', dataIndex: 'sort', width: 60 },
    ]}
    formItems={() => (<>
      <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input placeholder="如 0-100元" /></Form.Item>
      <Form.Item name="min_price" label="最低价(分)" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
      <Form.Item name="max_price" label="最高价(分)" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
      <Form.Item name="sort" label="排序" initialValue={0}><InputNumber min={0} /></Form.Item>
    </>)}
  />
}
