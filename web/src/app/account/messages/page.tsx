'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'

interface Message { id: number; title: string; content: string; is_read: number; created_at: string }

export default function MessagesPage() {
  const [list, setList] = useState<Message[]>([])
  const [loading, setLoading] = useState(true)

  const load = () => { api.get<{ list: Message[] }>('/messages').then(d => setList(d.list || d as any || [])).catch(() => {}).finally(() => setLoading(false)) }
  useEffect(load, [])

  const markRead = async (id: number) => {
    try { await api.put(`/messages/${id}/read`); load() } catch {}
  }

  const markAll = async () => {
    try { await api.put('/messages/read-all'); load() } catch {}
  }

  if (loading) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">加载中...</div>

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-2xl mx-auto">
        <Link href="/account" className="text-sm text-[#0071e3] hover:underline mb-6 inline-block">← 返回账户</Link>
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-semibold" style={{ color: '#1d1d1f' }}>我的消息</h1>
          {list.some(m => !m.is_read) && <button onClick={markAll} className="text-sm text-[#0071e3] hover:underline">全部已读</button>}
        </div>

        {list.length === 0 ? <p className="text-center py-20 text-[#86868b]">暂无消息</p> : (
          <div className="space-y-3">
            {list.map(m => (
              <div key={m.id} onClick={() => !m.is_read && markRead(m.id)}
                className={`p-4 rounded-2xl border-2 cursor-pointer transition-colors ${m.is_read ? 'border-gray-100 bg-white' : 'border-[#0071e3]/20 bg-blue-50'}`}>
                <div className="flex items-center justify-between mb-1">
                  <h3 className="text-sm font-medium" style={{ color: '#1d1d1f' }}>{m.title || '系统消息'}</h3>
                  {!m.is_read && <span className="w-2 h-2 rounded-full bg-[#0071e3]" />}
                </div>
                <p className="text-sm text-[#86868b]">{m.content}</p>
                <p className="text-xs text-[#86868b] mt-2">{new Date(m.created_at).toLocaleString()}</p>
              </div>
            ))}
          </div>
        )}
      </div>
    </section>
  )
}
