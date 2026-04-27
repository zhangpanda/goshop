'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, InputNumber } from 'antd'
import ImageUpload from '@/components/ImageUpload'

export default function SlidesPage() {
  return <CrudPage title="首页轮播" listUrl="/admin/slides" createUrl="/admin/slides"
    searchClient searchPlaceholder="名称"
    deleteUrl={r => `/admin/slides/${r.id}`} statusUrl={r => `/admin/slides/${r.id}/status`} updateUrl={r => `/admin/slides/${r.id}`} batchDelete
    columns={[
      { title: 'ID', dataIndex: 'id', width: 60 },
      { title: '名称', dataIndex: 'name' },
      { title: '图片', dataIndex: 'images', width: 120, render: (v: string) => {
        try { const arr = JSON.parse(v); return arr[0] ? <img src={arr[0]} style={{height:40}} /> : '-' } catch { return '-' }
      }},
      { title: '排序', dataIndex: 'sort', width: 60 },
    ]}
    formItems={(form) => (<>
      <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="images" label="轮播图片"
        getValueProps={(v) => ({ value: (() => { try { return JSON.parse(v) } catch { return [] } })() })}
        normalize={(v) => JSON.stringify(v || [])}
      ><ImageUpload max={10} /></Form.Item>
      <Form.Item name="url" label="跳转链接"><Input placeholder="点击图片跳转的链接" /></Form.Item>
      <Form.Item name="sort" label="排序" initialValue={0}><InputNumber min={0} /></Form.Item>
    </>)} modalWidth={600}
  />
}
