'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'

interface Answer { id: number; title: string; content: string; reply: string; created_at: string }

export default function AnswersPage() {
  const [list, setList] = useState<Answer[]>([])
  const [form, setForm] = useState({ title: '', content: '' })
  const [saving, setSaving] = useState(false)
  const [loading, setLoading] = useState(true)

  const load = () => { api.get<{ list: Answer[] }>('/answers').then(d => setList(d.list || d as any || [])).catch(() => {}).finally(() => setLoading(false)) }
  useEffect(load, [])

  const submit = async () => {
    if (!form.content) { alert('请输入内容'); return }
    setSaving(true)
    try { await api.post('/answers', form); setForm({ title: '', content: '' }); load() } catch (e: any) { alert(e.message) }
    setSaving(false)
  }

  if (loading) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">加载中...</div>

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-2xl mx-auto">
        <Link href="/account" className="text-sm text-[#0071e3] hover:underline mb-6 inline-block">← 返回账户</Link>
        <h1 className="text-2xl font-semibold mb-8" style={{ color: '#1d1d1f' }}>问答留言</h1>

        <div className="mb-8 space-y-3">
          <input placeholder="标题（选填）" value={form.title} onChange={e => setForm({ ...form, title: e.target.value })}
            className="w-full px-4 py-3 rounded-xl border border-gray-300 focus:border-[#0071e3] focus:outline-none" />
          <textarea placeholder="请输入留言内容" value={form.content} onChange={e => setForm({ ...form, content: e.target.value })} rows={3}
            className="w-full px-4 py-3 rounded-xl border border-gray-300 focus:border-[#0071e3] focus:outline-none resize-none" />
          <button onClick={submit} disabled={saving}
            className="w-full py-3 bg-[#0071e3] text-white rounded-full font-medium hover:bg-blue-600 transition-colors disabled:opacity-50">
            {saving ? '提交中...' : '提交留言'}
          </button>
        </div>

        {list.length === 0 ? <p className="text-center py-10 text-[#86868b]">暂无留言</p> : (
          <div className="space-y-4">
            {list.map(item => (
              <div key={item.id} className="border border-gray-200 rounded-2xl p-5">
                {item.title && <h3 className="text-sm font-medium mb-1" style={{ color: '#1d1d1f' }}>{item.title}</h3>}
                <p className="text-sm text-[#86868b]">{item.content}</p>
                {item.reply && (
                  <div className="mt-3 pl-3 border-l-2 border-[#0071e3]">
                    <p className="text-xs text-[#0071e3] mb-0.5">商家回复</p>
                    <p className="text-sm" style={{ color: '#1d1d1f' }}>{item.reply}</p>
                  </div>
                )}
                <p className="text-xs text-[#86868b] mt-2">{new Date(item.created_at).toLocaleString()}</p>
              </div>
            ))}
          </div>
        )}
      </div>
    </section>
  )
}
