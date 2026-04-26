'use client'
import { useState, Suspense } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'

export default function LoginPage() {
  return <Suspense><LoginInner /></Suspense>
}

function LoginInner() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const redirect = searchParams.get('redirect') || '/account'
  const [isLogin, setIsLogin] = useState(true)
  const [form, setForm] = useState({ username: '', password: '', confirm: '' })
  const [loading, setLoading] = useState(false)

  const submit = async () => {
    if (!form.username || !form.password) { alert('请填写完整'); return }
    if (!isLogin && form.password !== form.confirm) { alert('两次密码不一致'); return }
    setLoading(true)
    try {
      const url = isLogin ? '/login' : '/register'
      const res = await api.post<{ token: string }>(url, { username: form.username, password: form.password })
      localStorage.setItem('token', res.token)
      router.push(redirect)
    } catch (e: any) { alert(e.message || '操作失败') }
    setLoading(false)
  }

  return (
    <section className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-sm">
        <h1 className="text-3xl font-semibold text-center mb-2" style={{ color: '#1d1d1f' }}>{isLogin ? '登录' : '注册'}</h1>
        <p className="text-center text-sm mb-8" style={{ color: '#86868b' }}>{isLogin ? '登录你的 GoShop 账户' : '创建一个新的 GoShop 账户'}</p>

        <div className="space-y-4">
          <input placeholder="用户名" value={form.username} onChange={e => setForm({ ...form, username: e.target.value })}
            className="w-full px-4 py-3 rounded-xl border border-gray-300 focus:border-[#0071e3] focus:outline-none" />
          <input type="password" placeholder="密码" value={form.password} onChange={e => setForm({ ...form, password: e.target.value })}
            onKeyDown={e => e.key === 'Enter' && isLogin && submit()}
            className="w-full px-4 py-3 rounded-xl border border-gray-300 focus:border-[#0071e3] focus:outline-none" />
          {!isLogin && (
            <input type="password" placeholder="确认密码" value={form.confirm} onChange={e => setForm({ ...form, confirm: e.target.value })}
              onKeyDown={e => e.key === 'Enter' && submit()}
              className="w-full px-4 py-3 rounded-xl border border-gray-300 focus:border-[#0071e3] focus:outline-none" />
          )}
        </div>

        <button onClick={submit} disabled={loading}
          className="mt-6 w-full py-3 bg-[#0071e3] text-white rounded-full font-medium hover:bg-blue-600 transition-colors disabled:opacity-50">
          {loading ? '处理中...' : isLogin ? '登录' : '注册'}
        </button>

        <p className="text-center mt-6 text-sm" style={{ color: '#86868b' }}>
          {isLogin ? '还没有账户？' : '已有账户？'}
          <button onClick={() => setIsLogin(!isLogin)} className="text-[#0071e3] hover:underline ml-1">
            {isLogin ? '立即注册' : '去登录'}
          </button>
        </p>

        <div className="text-center mt-4">
          <Link href="/" className="text-sm text-[#86868b] hover:text-[#1d1d1f]">← 返回首页</Link>
        </div>
      </div>
    </section>
  )
}
