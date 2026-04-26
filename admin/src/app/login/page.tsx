'use client'
import { useState } from 'react'
import { Button, Card, Form, Input, message, Typography } from 'antd'
import { UserOutlined, LockOutlined, SafetyOutlined } from '@ant-design/icons'
import { useAdmin, AdminAuthProvider } from '@/lib/admin-auth'

function LoginForm() {
  const { login } = useAdmin()
  const [captchaKey, setCaptchaKey] = useState(() => `captcha_${Date.now()}`)
  const [captchaUrl, setCaptchaUrl] = useState(`/api/admin/captcha?key=captcha_${Date.now()}&t=${Date.now()}`)

  const refreshCaptcha = () => {
    const k = `captcha_${Date.now()}`
    setCaptchaKey(k)
    setCaptchaUrl(`/api/admin/captcha?key=${k}&t=${Date.now()}`)
  }

  const onFinish = async (v: { username: string; password: string; captcha: string }) => {
    try { await login(v.username, v.password, captchaKey, v.captcha) } catch (e: unknown) { message.error((e as Error).message); refreshCaptcha() }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' }}>
      <Card style={{ width: 400, borderRadius: 12, boxShadow: '0 8px 32px rgba(0,0,0,0.2)' }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <Typography.Title level={3} style={{ margin: 0 }}>GoShop</Typography.Title>
          <Typography.Text type="secondary">管理后台</Typography.Text>
        </div>
        <Form onFinish={onFinish} size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item name="captcha" rules={[{ required: true, message: '请输入验证码' }]}>
            <Input prefix={<SafetyOutlined />} placeholder="验证码"
              suffix={<img src={captchaUrl} alt="验证码" onClick={refreshCaptcha}
                style={{ height: 30, borderRadius: 4, cursor: 'pointer' }} title="点击刷新" />} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block style={{ height: 44, borderRadius: 8, fontSize: 16 }}>登录</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

export default function LoginPage() {
  return <AdminAuthProvider><LoginForm /></AdminAuthProvider>
}
