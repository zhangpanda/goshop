'use client'
import Link from 'next/link'
import { useEffect, useState, useCallback } from 'react'
import { api } from '@/lib/api'
import { useSiteConfig } from '@/lib/site-config'

interface Slide { id: number; name: string; images: string; url: string }
interface Goods { id: number; title: string; subtitle: string; main_image: string; skus: { price: number }[] }

const BG = '#f5f5f7'

export default function Home() {
  const [slides, setSlides] = useState<Slide[]>([])
  const [goods, setGoods] = useState<Goods[]>([])
  const [cur, setCur] = useState(0)
  const { home_site_name } = useSiteConfig()

  useEffect(() => {
    api.get<Slide[]>('/slides').then(d => setSlides(d || [])).catch(() => {})
    api.get<{ list: Goods[] }>('/goods?page_size=6').then(d => setGoods(d.list || [])).catch(() => {})
  }, [])

  useEffect(() => {
    if (slides.length < 2) return
    const t = setInterval(() => setCur(c => (c + 1) % slides.length), 4000)
    return () => clearInterval(t)
  }, [slides.length])

  const parseImg = useCallback((s: Slide) => {
    try { const arr = JSON.parse(s.images); return arr[0] || '' } catch { return s.images || '' }
  }, [])

  const img = (src: string) => src && !src.startsWith('http') && !src.startsWith('/') ? `/${src}` : src

  const hero = goods.slice(0, 2)
  const cards = goods.slice(2, 6)

  return (
    <div>
      {/* 轮播 / 大屏 hero */}
      {slides.length > 0 ? (
        <section className="relative w-full h-[580px] overflow-hidden bg-[#f5f5f7]">
          {slides.map((s, i) => (
            <Link key={s.id} href={s.url || '/products'} className={`absolute inset-0 transition-opacity duration-700 ${i === cur ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}>
              <img src={img(parseImg(s))} alt={s.name} className="w-full h-full object-cover" />
              <div className="absolute inset-0 flex flex-col items-center justify-end pb-16 bg-gradient-to-t from-black/40 to-transparent">
                <h2 className="text-3xl md:text-5xl font-semibold text-white">{s.name}</h2>
              </div>
            </Link>
          ))}
          <div className="absolute bottom-5 left-1/2 -translate-x-1/2 flex gap-2 z-10">
            {slides.map((_, i) => (
              <button key={i} onClick={() => setCur(i)} className={`w-2 h-2 rounded-full transition-all ${i === cur ? 'bg-white w-6' : 'bg-white/50'}`} />
            ))}
          </div>
        </section>
      ) : (
        <section className="h-[580px] flex flex-col items-center justify-center text-center px-4 bg-[#f5f5f7]">
          <h1 className="text-[56px] md:text-[80px] font-semibold tracking-tight leading-tight">{home_site_name}</h1>
          <p className="mt-4 text-[21px] text-[#6e6e73]">为品质而生的购物体验</p>
          <div className="mt-6 flex gap-4">
            <Link href="/products" className="px-6 py-2.5 bg-[#0071e3] text-white rounded-full text-sm font-medium hover:bg-[#0077ED] transition-colors">立即选购</Link>
            <Link href="/story" className="px-6 py-2.5 rounded-full text-sm font-medium border border-[#0071e3] text-[#0071e3] hover:bg-[#0071e3] hover:text-white transition-colors">了解更多</Link>
          </div>
        </section>
      )}

      {/* 大区块：每个产品独占全宽 */}
      {hero.map((g) => (
        <section key={g.id} className="mt-3" style={{ background: BG }}>
          <div className="text-center pt-10 pb-3">
            <h2 className="text-[40px] md:text-[56px] font-semibold tracking-tight leading-tight text-[#1d1d1f]">{g.title}</h2>
            <p className="mt-1 text-[17px] text-[#6e6e73]">{g.subtitle || '探索全新体验'}</p>
            <div className="mt-3 flex justify-center gap-4 text-sm">
              <Link href={`/products/${g.id}`} className="px-5 py-2 bg-[#0071e3] text-white rounded-full hover:bg-[#0077ED] transition-colors">进一步了解</Link>
              <Link href={`/products/${g.id}`} className="px-5 py-2 rounded-full border border-[#0071e3] text-[#0071e3] hover:bg-[#0071e3] hover:text-white transition-colors">购买</Link>
            </div>
          </div>
          {g.main_image ? (
            <div className="flex justify-center overflow-hidden">
              <img src={img(g.main_image)} alt={g.title} className="w-full max-w-[1060px] h-[500px] object-cover object-top" />
            </div>
          ) : <div className="h-8" />}
        </section>
      ))}

      {/* 2x2 小卡片网格，紧密排列 */}
      {cards.length > 0 && (
        <div className="grid md:grid-cols-2 gap-3 mt-3">
          {cards.map((g) => (
            <div key={g.id} style={{ background: BG }}>
              <div className="text-center pt-10 pb-4 px-4">
                <h3 className="text-[28px] font-semibold">{g.title}</h3>
                <p className="mt-1 text-[14px] text-[#6e6e73]">{g.subtitle || '探索更多'}</p>
                <div className="mt-2 flex justify-center gap-3 text-sm">
                  <Link href={`/products/${g.id}`} className="text-[#0071e3] hover:underline">进一步了解 {'>'}</Link>
                  <Link href={`/products/${g.id}`} className="text-[#0071e3] hover:underline">购买 {'>'}</Link>
                </div>
              </div>
              {g.main_image ? (
                <div className="flex justify-center overflow-hidden">
                  <img src={img(g.main_image)} alt={g.title} className="w-full max-w-[500px] h-[300px] object-cover object-top" />
                </div>
              ) : <div className="h-6" />}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
