'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api, formatPrice } from '@/lib/api'

interface HistoryItem { id: number; goods_id: number; title: string; main_image: string; price: number; created_at: string }

export default function HistoryPage() {
  const [list, setList] = useState<HistoryItem[]>([])
  const [loading, setLoading] = useState(true)

  const load = () => { api.get<{ list: HistoryItem[] }>('/history').then(d => setList(d.list || d as any || [])).catch(() => {}).finally(() => setLoading(false)) }
  useEffect(load, [])

  const clear = async () => {
    if (!confirm('确定清空浏览记录？')) return
    try { await api.del('/history'); setList([]) } catch {}
  }

  const img = (src: string) => src && !src.startsWith('http') && !src.startsWith('/') ? `/${src}` : src

  if (loading) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">加载中...</div>

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-2xl mx-auto">
        <Link href="/account" className="text-sm text-[#0071e3] hover:underline mb-6 inline-block">← 返回账户</Link>
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-semibold" style={{ color: '#1d1d1f' }}>浏览记录</h1>
          {list.length > 0 && <button onClick={clear} className="text-sm text-red-400 hover:text-red-600">清空记录</button>}
        </div>

        {list.length === 0 ? <p className="text-center py-20 text-[#86868b]">暂无浏览记录</p> : (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {list.map(item => (
              <Link key={item.id} href={`/products/${item.goods_id || item.id}`} className="group">
                <div className="aspect-square bg-[#f5f5f7] rounded-2xl overflow-hidden mb-2">
                  {item.main_image ? <img src={img(item.main_image)} alt="" className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" /> : <div className="w-full h-full flex items-center justify-center text-gray-300 text-xs">暂无图片</div>}
                </div>
                <h3 className="text-xs font-medium text-[#1d1d1f] truncate">{item.title}</h3>
                {item.price > 0 && <p className="text-xs text-[#86868b]">¥{formatPrice(item.price)}</p>}
              </Link>
            ))}
          </div>
        )}
      </div>
    </section>
  )
}
