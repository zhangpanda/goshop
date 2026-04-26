'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'

interface Profile { username: string; nickname: string; mobile: string; email: string; avatar: string; gender: number }

export default function ProfilePage() {
  const [profile, setProfile] = useState<Profile | null>(null)
  const [form, setForm] = useState({ nickname: '', mobile: '', email: '', gender: 0 })
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    api.get<Profile>('/user/profile').then(d => {
      setProfile(d)
      setForm({ nickname: d.nickname || '', mobile: d.mobile || '', email: d.email || '', gender: d.gender || 0 })
    }).catch(() => {})
  }, [])

  const save = async () => {
    setSaving(true)
    try {
      await api.put('/user/profile', form)
      alert('保存成功')
    } catch (e: any) { alert(e.message) }
    setSaving(false)
  }

  if (!profile) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">加载中...</div>

  const inputCls = "w-full px-4 py-3 rounded-xl border border-gray-300 focus:border-[#0071e3] focus:outline-none"

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-lg mx-auto">
        <Link href="/account" className="text-sm text-[#0071e3] hover:underline mb-6 inline-block">← 返回账户</Link>
        <h1 className="text-2xl font-semibold mb-8" style={{ color: '#1d1d1f' }}>个人资料</h1>

        <div className="space-y-4">
          <div>
            <label className="text-sm text-[#86868b] mb-1 block">用户名</label>
            <input value={profile.username} disabled className={inputCls + " bg-[#f5f5f7]"} />
          </div>
          <div>
            <label className="text-sm text-[#86868b] mb-1 block">昵称</label>
            <input value={form.nickname} onChange={e => setForm({ ...form, nickname: e.target.value })} className={inputCls} />
          </div>
          <div>
            <label className="text-sm text-[#86868b] mb-1 block">手机号</label>
            <input value={form.mobile} onChange={e => setForm({ ...form, mobile: e.target.value })} className={inputCls} />
          </div>
          <div>
            <label className="text-sm text-[#86868b] mb-1 block">邮箱</label>
            <input value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} className={inputCls} />
          </div>
          <div>
            <label className="text-sm text-[#86868b] mb-1 block">性别</label>
            <div className="flex gap-4">
              {[{ v: 0, l: '保密' }, { v: 1, l: '男' }, { v: 2, l: '女' }].map(g => (
                <button key={g.v} onClick={() => setForm({ ...form, gender: g.v })}
                  className={`px-5 py-2 rounded-full text-sm border-2 ${form.gender === g.v ? 'border-[#0071e3] bg-blue-50 text-[#0071e3]' : 'border-gray-200'}`}>{g.l}</button>
              ))}
            </div>
          </div>
        </div>

        <button onClick={save} disabled={saving}
          className="mt-8 w-full py-3 bg-[#0071e3] text-white rounded-full font-medium hover:bg-blue-600 transition-colors disabled:opacity-50">
          {saving ? '保存中...' : '保存资料'}
        </button>
      </div>
    </section>
  )
}
