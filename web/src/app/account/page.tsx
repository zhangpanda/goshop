'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'

const menuItems = [
  { name: '我的订单', href: '/account/orders', icon: '📦' },
  { name: '订单售后', href: '/account/aftersale', icon: '🔄' },
  { name: '收货地址', href: '/account/address', icon: '📍' },
  { name: '我的收藏', href: '/account/favorites', icon: '❤️' },
  { name: '我的积分', href: '/account/points', icon: '⭐' },
  { name: '浏览记录', href: '/account/history', icon: '👁️' },
  { name: '我的消息', href: '/account/messages', icon: '💬' },
  { name: '个人资料', href: '/account/profile', icon: '👤' },
  { name: '安全设置', href: '/account/security', icon: '🔒' },
  { name: '问答留言', href: '/account/answers', icon: '✉️' },
]

export default function AccountPage() {
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) { setChecking(false); return }
    api.get('/user/profile').then(setUser).catch(() => localStorage.removeItem('token')).finally(() => setChecking(false))
  }, [])

  const logout = () => {
    localStorage.removeItem('token')
    setUser(null)
    router.push('/')
  }

  if (checking) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">加载中...</div>

  if (!user) {
    return (
      <section className="min-h-screen flex items-center justify-center px-4">
        <div className="text-center">
          <p className="text-[#86868b] mb-4">请先登录</p>
          <a href="/login" className="px-8 py-3 bg-[#0071e3] text-white rounded-full font-medium hover:bg-blue-600 transition-colors inline-block">去登录</a>
        </div>
      </section>
    )
  }

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-lg mx-auto">
        <h1 className="text-3xl font-semibold mb-8" style={{ color: '#1d1d1f' }}>我的账户</h1>

        <div className="rounded-2xl p-6 mb-8" style={{ background: '#f5f5f7' }}>
          <p className="text-lg font-medium" style={{ color: '#1d1d1f' }}>{user.nickname || user.username}</p>
          <p className="text-sm mt-1" style={{ color: '#86868b' }}>积分: {user.integral || 0}</p>
        </div>

        <div className="space-y-2">
          {menuItems.map(item => (
            <a key={item.href} href={item.href} className="flex items-center gap-3 p-4 rounded-2xl border border-gray-200 hover:border-gray-400 transition-colors">
              <span className="text-lg">{item.icon}</span>
              <span className="flex-1" style={{ color: '#1d1d1f' }}>{item.name}</span>
              <span style={{ color: '#86868b' }}>→</span>
            </a>
          ))}
        </div>

        <button onClick={logout}
          className="mt-8 w-full py-3 border border-gray-300 rounded-full text-red-400 hover:text-red-600 hover:border-red-400 transition-colors">
          退出登录
        </button>
      </div>
    </section>
  )
}
