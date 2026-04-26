'use client'
import { useState } from 'react'
import { Table, Input, Modal, Form, Typography, Rate, message } from 'antd'
import { api } from '@/lib/api'
import { useUserMap } from '@/lib/useIdMap'

interface Review { id: number; user_id: number; rating: number; content: string; reply: string; created_at: string }

export default function ReviewsPage() {
  const [list, setList] = useState<Review[]>([])
  const [goodsId, setGoodsId] = useState('')
  const [replyTarget, setReplyTarget] = useState<Review | null>(null)
  const [form] = Form.useForm()
  const userMap = useUserMap(list.map(r => r.user_id))

  const load = (gid: string) => {
    if (!gid) return
    api.get<Review[]>(`/goods/${gid}/reviews`).then(r => setList(Array.isArray(r) ? r : []))
  }

  return (
    <>
      <Typography.Title level={4}>评价管理</Typography.Title>
      <Input.Search placeholder="输入商品ID查询评价" style={{ width: 300, marginBottom: 16 }} value={goodsId}
        onChange={e => setGoodsId(e.target.value)} onSearch={v => load(v)} enterButton="查询" />
      <Table dataSource={list} rowKey="id"
        columns={[
          { title: 'ID', dataIndex: 'id', width: 60 },
          { title: '用户', dataIndex: 'user_id', width: 120, render: (v: number) => userMap[v] || `#${v}` },
          { title: '评分', dataIndex: 'rating', width: 160, render: (v: number) => <Rate disabled value={v} allowHalf /> },
          { title: '内容', dataIndex: 'content', ellipsis: true },
          { title: '回复', dataIndex: 'reply', ellipsis: true },
          { title: '操作', width: 80, render: (_: unknown, r: Review) => <a onClick={() => { setReplyTarget(r); form.setFieldsValue({ content: r.reply }) }}>回复</a> },
        ]}
      />
      <Modal title="回复评价" open={!!replyTarget} onCancel={() => setReplyTarget(null)} onOk={async () => {
        const v = form.getFieldsValue(); await api.put(`/admin/reviews/${replyTarget!.id}/reply`, v)
        message.success('已回复'); setReplyTarget(null); load(goodsId)
      }} forceRender>
        <Form form={form}><Form.Item name="content"><Input.TextArea rows={3} /></Form.Item></Form>
      </Modal>
    </>
  )
}
