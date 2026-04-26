'use client'
import { useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'

export default function SecurityPage() {
  const [form, setForm] = useState({ old_password: '', new_password: '', confirm: '' })
  const [saving, setSaving] = useState(false)

  const save = async () => {
    if (!form.old_password || !form.new_password) { alert('请填写完整'); return }
    if (form.new_password !== form.confirm) { alert('两次密码不一致'); return }
    setSaving(true)
    try {
      await api.put('/user/password', { old_password: form.old_password, new_password: form.new_password })
      alert('密码修改成功')
      setForm({ old_password: '', new_password: '', confirm: '' })
    } catch (e: any) { alert(e.message) }
    setSaving(false)
  }

  const inputCls = "w-full px-4 py-3 rounded-xl border border-gray-300 focus:border-[#0071e3] focus:outline-none"

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-lg mx-auto">
        <Link href="/account" className="text-sm text-[#0071e3] hover:underline mb-6 inline-block">← 返回账户</Link>
        <h1 className="text-2xl font-semibold mb-8" style={{ color: '#1d1d1f' }}>安全设置</h1>

        <div className="space-y-4">
          <div>
            <label className="text-sm text-[#86868b] mb-1 block">当前密码</label>
            <input type="password" value={form.old_password} onChange={e => setForm({ ...form, old_password: e.target.value })} className={inputCls} />
          </div>
          <div>
            <label className="text-sm text-[#86868b] mb-1 block">新密码</label>
            <input type="password" value={form.new_password} onChange={e => setForm({ ...form, new_password: e.target.value })} className={inputCls} />
          </div>
          <div>
            <label className="text-sm text-[#86868b] mb-1 block">确认新密码</label>
            <input type="password" value={form.confirm} onChange={e => setForm({ ...form, confirm: e.target.value })}
              onKeyDown={e => e.key === 'Enter' && save()} className={inputCls} />
          </div>
        </div>

        <button onClick={save} disabled={saving}
          className="mt-8 w-full py-3 bg-[#0071e3] text-white rounded-full font-medium hover:bg-blue-600 transition-colors disabled:opacity-50">
          {saving ? '保存中...' : '修改密码'}
        </button>
      </div>
    </section>
  )
}
