'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api, formatPrice } from '@/lib/api'

interface Fav { id: number; goods: { id: number; title: string; main_image: string; skus: { price: number }[] } }

export default function FavoritesPage() {
  const [list, setList] = useState<Fav[]>([])
  useEffect(() => { api.get<{ list: Fav[] }>('/favorites').then(d => setList(d.list || [])).catch(() => {}) }, [])

  const removeFav = async (goodsId: number) => {
    await api.post(`/favorites/${goodsId}`)
    setList(prev => prev.filter(f => f.goods.id !== goodsId))
  }

  return (
    <section className="min-h-screen py-20 px-4">
        <Link href="/account" className="inline-flex items-center text-sm text-[#0071e3] hover:underline mb-6">← 返回账户</Link>
      <div className="max-w-2xl mx-auto">
        <h1 className="text-3xl font-semibold mb-10">我的收藏</h1>
        {list.length === 0 ? (
          <p className="text-center text-[var(--muted)] py-20">暂无收藏</p>
        ) : (
          <div className="space-y-4">
            {list.map(f => (
              <div key={f.id} className="flex gap-4 p-4 rounded-2xl border border-gray-100 hover:border-gray-300 transition-colors">
                <Link href={`/products/${f.goods.id}`} className="flex gap-4 flex-1 min-w-0">
                  <div className="w-16 h-16 bg-[var(--surface)] rounded-xl flex-shrink-0 flex items-center justify-center overflow-hidden">
                    {f.goods.main_image ? <img src={f.goods.main_image} alt="" className="w-full h-full object-cover" /> : <span className="text-xs text-gray-400">图</span>}
                  </div>
                  <div>
                    <p className="font-medium">{f.goods.title}</p>
                    {f.goods.skus?.[0] && <p className="text-sm text-[var(--muted)] mt-1">¥{formatPrice(f.goods.skus[0].price)} 起</p>}
                  </div>
                </Link>
                <button onClick={() => removeFav(f.goods.id)}
                  className="self-center text-xs text-red-500 hover:text-red-700 flex-shrink-0">取消收藏</button>
              </div>
            ))}
          </div>
        )}
      </div>
    </section>
  )
}
