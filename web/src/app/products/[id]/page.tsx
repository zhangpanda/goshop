'use client'
import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import { api, formatPrice } from '@/lib/api'

interface SKU { id: number; name: string; price: number; stock: number }
interface Goods { id: number; title: string; subtitle: string; main_image: string; detail: string; skus: SKU[] }

export default function ProductDetail() {
  const { id } = useParams()
  const [goods, setGoods] = useState<Goods | null>(null)
  const [selectedSku, setSelectedSku] = useState<SKU | null>(null)
  const [adding, setAdding] = useState(false)
  const [isFav, setIsFav] = useState(false)

  useEffect(() => {
    api.get<Goods>(`/goods/${id}`).then(d => { setGoods(d); setSelectedSku(d.skus?.[0] || null) }).catch(() => {})
  }, [id])

  const addToCart = async () => {
    if (!selectedSku || !goods) return
    setAdding(true)
    try { await api.post('/cart', { goods_id: goods.id, sku_id: selectedSku.id, quantity: 1 }); alert('已加入购物袋') }
    catch (e: any) { alert(e.message || '请先登录') }
    setAdding(false)
  }

  const toggleFav = async () => {
    if (!goods) return
    try {
      const res = await api.post<{ is_favorite: boolean }>(`/favorites/${goods.id}`)
      setIsFav(res.is_favorite)
    } catch { alert('请先登录') }
  }

  if (!goods) return <div className="min-h-screen flex items-center justify-center" style={{ color: '#86868b' }}>加载中...</div>

  return (
    <>
      <section className="min-h-screen flex flex-col items-center justify-center text-center px-4 bg-[#f5f5f7]">
        <h1 className="text-4xl md:text-7xl font-semibold tracking-tight" style={{ color: '#1d1d1f' }}>{goods.title}</h1>
        <p className="mt-3 text-xl" style={{ color: '#86868b' }}>{goods.subtitle || '探索全新体验'}</p>
        <p className="mt-4 text-2xl font-medium" style={{ color: '#1d1d1f' }}>¥{formatPrice(selectedSku?.price || 0)}</p>
        <div className="mt-12 w-full max-w-2xl aspect-[4/3] rounded-3xl flex items-center justify-center overflow-hidden" style={{ background: goods.main_image ? undefined : '#e8e8ed' }}>
          {goods.main_image ? <img src={goods.main_image} alt={goods.title} className="w-full h-full object-cover" /> : (
            <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="#c7c7cc" strokeWidth="1" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-5-5L5 21"/></svg>
          )}
        </div>
      </section>

      <section className="py-20 px-4">
        <div className="max-w-xl mx-auto">
          <h2 className="text-2xl font-semibold mb-6" style={{ color: '#1d1d1f' }}>选择型号</h2>
          <div className="space-y-3">
            {goods.skus?.map(sku => (
              <button key={sku.id} onClick={() => setSelectedSku(sku)}
                className={`w-full flex items-center justify-between p-4 rounded-2xl border-2 transition-all ${selectedSku?.id === sku.id ? 'border-[#0071e3] bg-blue-50' : 'border-gray-200 hover:border-gray-400'}`}>
                <span className="font-medium" style={{ color: '#1d1d1f' }}>{sku.name}</span>
                <span style={{ color: '#86868b' }}>¥{formatPrice(sku.price)}</span>
              </button>
            ))}
          </div>
          <button onClick={addToCart} disabled={adding || !selectedSku?.stock}
            className="mt-8 w-full py-4 bg-[#0071e3] text-white rounded-full text-lg font-medium hover:bg-blue-600 transition-colors disabled:opacity-50">
            {!selectedSku?.stock ? '暂时缺货' : adding ? '添加中...' : '加入购物袋'}
          </button>
          <button onClick={toggleFav}
            className="mt-3 w-full py-3 border-2 border-gray-200 rounded-full text-base font-medium hover:border-gray-400 transition-colors">
            {isFav ? '♥ 已收藏' : '♡ 收藏'}
          </button>
          {selectedSku && <p className="mt-3 text-center text-sm" style={{ color: '#86868b' }}>库存 {selectedSku.stock} 件</p>}
        </div>
      </section>

      {goods.detail && (
        <section className="py-20 px-4 bg-[#f5f5f7]">
          <div className="max-w-3xl mx-auto prose prose-lg" dangerouslySetInnerHTML={{ __html: goods.detail }} />
        </section>
      )}
    </>
  )
}
