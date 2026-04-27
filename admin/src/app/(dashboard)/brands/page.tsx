'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, InputNumber } from 'antd'
import ImageUpload from '@/components/ImageUpload'

export default function BrandsPage() {
  return <CrudPage title="品牌管理" listUrl="/admin/brands" createUrl="/admin/brands"
    updateUrl={r => `/admin/brands/${r.id}`} deleteUrl={r => `/admin/brands/${r.id}`} statusUrl={r => `/admin/brands/${r.id}/status`}
    searchable searchClient searchPlaceholder="品牌名称" batchDelete
    columns={[
      { title: 'ID', dataIndex: 'id', width: 60 },
      { title: '名称', dataIndex: 'name' },
      { title: 'Logo', dataIndex: 'logo', width: 80, render: (v: string) => v ? <img src={v} alt="" style={{ height: 30 }} /> : '-' },
      { title: '描述', dataIndex: 'desc', ellipsis: true },
      { title: '排序', dataIndex: 'sort', width: 60 },
    ]}
    detailItems={r => [
      { label: 'ID', value: r.id },
      { label: '名称', value: (r as Record<string, unknown>).name as string },
      { label: 'Logo', value: (r as Record<string, unknown>).logo ? <img src={(r as Record<string, unknown>).logo as string} alt="" style={{ maxWidth: 200 }} /> : '-' },
      { label: '描述', value: (r as Record<string, unknown>).desc as string || '-' },
      { label: '排序', value: (r as Record<string, unknown>).sort as number },
    ]}
    formItems={() => (<>
      <Form.Item name="name" label="品牌名称" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="logo" label="品牌Logo"><ImageUpload /></Form.Item>
      <Form.Item name="desc" label="品牌描述"><Input.TextArea rows={3} /></Form.Item>
      <Form.Item name="sort" label="排序" initialValue={0}><InputNumber min={0} /></Form.Item>
    </>)}
  />
}
