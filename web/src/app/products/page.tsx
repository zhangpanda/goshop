'use client'
import { useEffect, useState, useCallback, Suspense } from 'react'
import Link from 'next/link'
import { useRouter, useSearchParams, usePathname } from 'next/navigation'
import { api, formatPrice } from '@/lib/api'

interface Goods { id: number; title: string; main_image: string; skus: { price: number }[] }
interface Category { id: number; name: string; children?: Category[] }

const sortOptions = [
  { label: '综合', value: '' },
  { label: '价格↑', value: 'price_asc' },
  { label: '价格↓', value: 'price_desc' },
  { label: '最新', value: 'new' },
]

function ProductsInner() {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const keyword = searchParams.get('keyword') || ''
  const categoryId = searchParams.get('category_id') || ''
  const orderBy = searchParams.get('order_by') || ''
  const page = parseInt(searchParams.get('page') || '1')

  const [goods, setGoods] = useState<Goods[]>([])
  const [total, setTotal] = useState(0)
  const [categories, setCategories] = useState<Category[]>([])
  const [expandedCat, setExpandedCat] = useState<number | null>(null)
  const limit = 12

  const updateParams = useCallback((updates: Record<string, string>) => {
    const params = new URLSearchParams(searchParams.toString())
    Object.entries(updates).forEach(([k, v]) => { v ? params.set(k, v) : params.delete(k) })
    if (!updates.page) params.delete('page')
    router.push(`${pathname}?${params.toString()}`)
  }, [searchParams, pathname, router])

  useEffect(() => {
    api.get<Category[]>('/categories').then(d => setCategories(d || [])).catch(() => {})
  }, [])

  useEffect(() => {
    const params = new URLSearchParams()
    if (keyword) params.set('keyword', keyword)
    if (categoryId) params.set('category_id', categoryId)
    if (orderBy) params.set('order_by', orderBy)
    params.set('page', String(page))
    params.set('limit', String(limit))
    api.get<{ list: Goods[]; total: number }>(`/goods?${params}`).then(d => {
      setGoods(d.list || [])
      setTotal(d.total || 0)
    }).catch(() => {})
  }, [keyword, categoryId, orderBy, page])

  const totalPages = Math.ceil(total / limit) || 1
  const img = (src: string) => src && !src.startsWith('http') && !src.startsWith('/') ? `/${src}` : src

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-3xl md:text-4xl font-semibold text-center mb-4" style={{ color: '#1d1d1f' }}>
          {keyword ? `搜索: ${keyword}` : '全部产品'}
        </h1>
        <p className="text-center text-sm mb-10" style={{ color: '#86868b' }}>共 {total} 件商品</p>

        <div className="flex gap-8">
          {/* 左侧分类 */}
          <aside className="hidden md:block w-48 flex-shrink-0">
            <h3 className="text-sm font-semibold mb-3" style={{ color: '#1d1d1f' }}>分类</h3>
            <button onClick={() => updateParams({ category_id: '' })} className={`block w-full text-left text-sm py-1.5 ${!categoryId ? 'text-[#0071e3] font-medium' : 'text-[#86868b] hover:text-[#1d1d1f]'}`}>全部</button>
            {categories.map(c => (
              <div key={c.id}>
                <button onClick={() => { updateParams({ category_id: String(c.id) }); setExpandedCat(expandedCat === c.id ? null : c.id) }}
                  className={`block w-full text-left text-sm py-1.5 ${categoryId === String(c.id) ? 'text-[#0071e3] font-medium' : 'text-[#86868b] hover:text-[#1d1d1f]'}`}>
                  {c.name}
                </button>
                {expandedCat === c.id && c.children?.map(sub => (
                  <button key={sub.id} onClick={() => updateParams({ category_id: String(sub.id) })}
                    className={`block w-full text-left text-sm py-1 pl-4 ${categoryId === String(sub.id) ? 'text-[#0071e3] font-medium' : 'text-[#86868b] hover:text-[#1d1d1f]'}`}>
                    {sub.name}
                  </button>
                ))}
              </div>
            ))}
          </aside>

          {/* 右侧内容 */}
          <div className="flex-1">
            {/* 排序栏 */}
            <div className="flex gap-4 mb-6 border-b pb-3">
              {sortOptions.map(s => (
                <button key={s.value} onClick={() => updateParams({ order_by: s.value })}
                  className={`text-sm pb-1 ${orderBy === s.value || (!orderBy && !s.value) ? 'text-[#0071e3] font-medium border-b-2 border-[#0071e3]' : 'text-[#86868b] hover:text-[#1d1d1f]'}`}>
                  {s.label}
                </button>
              ))}
            </div>

            {/* 商品网格 */}
            <div className="grid grid-cols-2 md:grid-cols-3 gap-5">
              {goods.map(g => (
                <Link key={g.id} href={`/products/${g.id}`} className="group">
                  <div className="aspect-square bg-[#f5f5f7] rounded-2xl overflow-hidden mb-3">
                    {g.main_image ? <img src={img(g.main_image)} alt={g.title} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" /> : <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-[#f5f5f7] to-[#e8e8ed]"><svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#c7c7cc" strokeWidth="1" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-5-5L5 21"/></svg></div>}
                  </div>
                  <h3 className="text-sm font-medium text-[#1d1d1f] truncate">{g.title}</h3>
                  {g.skus?.[0] && <p className="text-sm text-[#86868b] mt-0.5">¥{formatPrice(g.skus[0].price)}</p>}
                </Link>
              ))}
            </div>

            {goods.length === 0 && <p className="text-center py-20 text-[#86868b]">暂无商品</p>}

            {/* 分页 */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center gap-2 mt-10">
                <button disabled={page <= 1} onClick={() => updateParams({ page: String(page - 1) })}
                  className="px-4 py-2 rounded-full text-sm border border-gray-200 disabled:opacity-30 hover:bg-[#f5f5f7]">上一页</button>
                {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
                  let p: number
                  if (totalPages <= 7) p = i + 1
                  else if (page <= 4) p = i + 1
                  else if (page >= totalPages - 3) p = totalPages - 6 + i
                  else p = page - 3 + i
                  return (
                    <button key={p} onClick={() => updateParams({ page: String(p) })}
                      className={`w-9 h-9 rounded-full text-sm ${p === page ? 'bg-[#0071e3] text-white' : 'hover:bg-[#f5f5f7]'}`}>{p}</button>
                  )
                })}
                <button disabled={page >= totalPages} onClick={() => updateParams({ page: String(page + 1) })}
                  className="px-4 py-2 rounded-full text-sm border border-gray-200 disabled:opacity-30 hover:bg-[#f5f5f7]">下一页</button>
              </div>
            )}
          </div>
        </div>
      </div>
    </section>
  )
}

export default function Products() {
  return <Suspense><ProductsInner /></Suspense>
}
