'use client'

import { useEffect } from 'react'
import { Typography, Card, Form, Select, Input, Button, InputNumber, message } from 'antd'
import { api } from '@/lib/api'

/**
 * 多语言与货币：对接 /admin/multilingual、/admin/currency。
 */
export default function LocalePage() {
  const [mlForm] = Form.useForm()
  const [curForm] = Form.useForm()

  useEffect(() => {
    api.get<{ default_lang: string; available: string[] }>('/admin/multilingual').then(r => {
      mlForm.setFieldsValue({ default_lang: r.default_lang, available: r.available })
    }).catch(() => {})
    api.get<{ symbol: string; code: string; name: string; rate: number }>('/admin/currency').then(r => {
      curForm.setFieldsValue(r)
    }).catch(() => {})
  }, [mlForm, curForm])

  return (
    <>
      <Typography.Title level={4}>语言与货币</Typography.Title>
      <Card title="多语言" style={{ marginBottom: 16 }}>
        <Form form={mlForm} layout="vertical" onFinish={async v => {
          await api.post('/admin/multilingual', { default_lang: v.default_lang, available: v.available || [] })
          message.success('已保存')
        }}>
          <Form.Item name="default_lang" label="默认语言" rules={[{ required: true }]}>
            <Select options={[
              { value: 'zh', label: '中文' },
              { value: 'en', label: 'English' },
              { value: 'cht', label: '繁体' },
            ]} />
          </Form.Item>
          <Form.Item name="available" label="可用语言" rules={[{ required: true }]}>
            <Select
              mode="multiple"
              options={[
                { value: 'zh', label: '中文' },
                { value: 'en', label: 'English' },
                { value: 'cht', label: '繁体' },
              ]}
              placeholder="至少选择一种"
            />
          </Form.Item>
          <Button type="primary" htmlType="submit">保存多语言</Button>
        </Form>
      </Card>
      <Card title="货币">
        <Form form={curForm} layout="vertical" onFinish={async v => {
          await api.post('/admin/currency', v)
          message.success('已保存')
        }}>
          <Form.Item name="symbol" label="符号" rules={[{ required: true }]}><Input placeholder="如 ¥" /></Form.Item>
          <Form.Item name="code" label="货币代码" rules={[{ required: true }]}><Input placeholder="如 CNY" /></Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input placeholder="如 人民币" /></Form.Item>
          <Form.Item name="rate" label="汇率（相对展示基准）" rules={[{ required: true }]}>
            <InputNumber min={0} step={0.01} style={{ width: '100%' }} />
          </Form.Item>
          <Button type="primary" htmlType="submit">保存货币</Button>
        </Form>
      </Card>
    </>
  )
}
