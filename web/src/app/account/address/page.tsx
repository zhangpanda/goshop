'use client'
import Link from 'next/link'
import { useEffect, useState } from 'react'
import { api } from '@/lib/api'

interface Address { id: number; name: string; phone: string; province: string; city: string; district: string; detail: string; is_default: boolean }

const emptyForm = { name: '', phone: '', province: '', city: '', district: '', detail: '' }

export default function AddressPage() {
  const [list, setList] = useState<Address[]>([])
  const [form, setForm] = useState(emptyForm)
  const [editId, setEditId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)

  const load = () => { api.get<Address[]>('/address').then(setList).catch(() => {}) }
  useEffect(load, [])

  const startEdit = (a: Address) => {
    setEditId(a.id)
    setForm({ name: a.name, phone: a.phone, province: a.province, city: a.city, district: a.district, detail: a.detail })
  }

  const cancelEdit = () => { setEditId(null); setForm(emptyForm) }

  const save = async () => {
    if (!form.name || !form.phone || !form.detail) { alert('请填写完整'); return }
    setSaving(true)
    try {
      if (editId) {
        await api.put(`/address/${editId}`, { ...form })
      } else {
        await api.post('/address', { ...form, is_default: list.length === 0 })
      }
      cancelEdit()
      load()
    } catch (e: any) { alert(e.message) }
    setSaving(false)
  }

  const del = async (id: number) => {
    if (!confirm('确定删除？')) return
    await api.del(`/address/${id}`)
    load()
  }

  const setDefault = async (id: number) => {
    try { await api.put(`/address/${id}`, { is_default: true }); load() } catch (e: any) { alert(e.message) }
  }

  const inputCls = "px-4 py-3 rounded-xl border border-gray-300 focus:border-[var(--accent)] focus:outline-none"

  return (
    <section className="min-h-screen py-20 px-4">
        <Link href="/account" className="inline-flex items-center text-sm text-[#0071e3] hover:underline mb-6">← 返回账户</Link>
      <div className="max-w-2xl mx-auto">
        <h1 className="text-3xl font-semibold mb-10">收货地址</h1>

        <div className="space-y-3 mb-8">
          {list.map(a => (
            <div key={a.id} className={`p-4 rounded-2xl border-2 ${a.is_default ? 'border-[var(--accent)] bg-blue-50' : 'border-gray-200'}`}>
              <div className="flex justify-between">
                <p className="font-medium">{a.name} <span className="text-[var(--muted)] font-normal ml-2">{a.phone}</span></p>
                {a.is_default && <span className="text-xs text-[var(--accent)]">默认</span>}
              </div>
              <p className="text-sm text-[var(--muted)] mt-1">{a.province}{a.city}{a.district} {a.detail}</p>
              <div className="flex gap-3 mt-2">
                <button onClick={() => startEdit(a)} className="text-xs text-[#0071e3] hover:underline">编辑</button>
                {!a.is_default && <button onClick={() => setDefault(a.id)} className="text-xs text-[#86868b] hover:text-[#1d1d1f]">设为默认</button>}
                <button onClick={() => del(a.id)} className="text-xs text-red-400 hover:text-red-600">删除</button>
              </div>
            </div>
          ))}
        </div>

        <h2 className="text-lg font-medium mb-4">{editId ? '编辑地址' : '添加新地址'}</h2>
        <div className="grid grid-cols-2 gap-3">
          <input placeholder="姓名" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} className={inputCls} />
          <input placeholder="手机号" value={form.phone} onChange={e => setForm({ ...form, phone: e.target.value })} className={inputCls} />
          <input placeholder="省" value={form.province} onChange={e => setForm({ ...form, province: e.target.value })} className={inputCls} />
          <input placeholder="市" value={form.city} onChange={e => setForm({ ...form, city: e.target.value })} className={inputCls} />
          <input placeholder="区" value={form.district} onChange={e => setForm({ ...form, district: e.target.value })} className={"col-span-2 " + inputCls} />
          <input placeholder="详细地址" value={form.detail} onChange={e => setForm({ ...form, detail: e.target.value })} className={"col-span-2 " + inputCls} />
        </div>
        <div className="flex gap-3 mt-4">
          <button onClick={save} disabled={saving} className="flex-1 py-3 bg-[var(--accent)] text-white rounded-full font-medium hover:bg-blue-600 transition-colors disabled:opacity-50">
            {saving ? '保存中...' : editId ? '更新地址' : '保存地址'}
          </button>
          {editId && (
            <button onClick={cancelEdit} className="px-6 py-3 border-2 border-gray-200 rounded-full font-medium hover:border-gray-400 transition-colors">取消</button>
          )}
        </div>
      </div>
    </section>
  )
}
