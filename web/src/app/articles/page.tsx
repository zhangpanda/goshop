'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'

interface Article { id: number; title: string; content: string; cover: string; created_at: string; article_category_name: string }
interface Category { id: number; name: string }

export default function ArticlesPage() {
  const [articles, setArticles] = useState<Article[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [catId, setCatId] = useState('')

  useEffect(() => {
    api.get<{ list: Category[] }>('/article-categories').then(d => setCategories(d.list || d as any || [])).catch(() => {})
  }, [])

  useEffect(() => {
    const q = catId ? `?category_id=${catId}` : ''
    api.get<{ list: Article[] }>(`/articles${q}`).then(d => setArticles(d.list || d as any || [])).catch(() => {})
  }, [catId])

  const img = (src: string) => src && !src.startsWith('http') && !src.startsWith('/') ? `/${src}` : src

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-semibold text-center mb-10" style={{ color: '#1d1d1f' }}>文章资讯</h1>

        {categories.length > 0 && (
          <div className="flex gap-3 justify-center mb-10 flex-wrap">
            <button onClick={() => setCatId('')} className={`px-4 py-1.5 rounded-full text-sm ${!catId ? 'bg-[#0071e3] text-white' : 'bg-[#f5f5f7] text-[#86868b] hover:text-[#1d1d1f]'}`}>全部</button>
            {categories.map(c => (
              <button key={c.id} onClick={() => setCatId(String(c.id))} className={`px-4 py-1.5 rounded-full text-sm ${catId === String(c.id) ? 'bg-[#0071e3] text-white' : 'bg-[#f5f5f7] text-[#86868b] hover:text-[#1d1d1f]'}`}>{c.name}</button>
            ))}
          </div>
        )}

        {articles.length === 0 ? <p className="text-center py-20 text-[#86868b]">暂无文章</p> : (
          <div className="space-y-6">
            {articles.map(a => (
              <Link key={a.id} href={`/articles/${a.id}`} className="flex gap-5 group">
                {a.cover && (
                  <div className="w-40 h-28 flex-shrink-0 bg-[#f5f5f7] rounded-xl overflow-hidden">
                    <img src={img(a.cover)} alt="" className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
                  </div>
                )}
                <div className="flex-1 min-w-0">
                  <h2 className="text-lg font-medium group-hover:text-[#0071e3] transition-colors" style={{ color: '#1d1d1f' }}>{a.title}</h2>
                  <p className="text-sm text-[#86868b] mt-1 line-clamp-2">{a.content?.replace(/<[^>]+>/g, '').slice(0, 120)}</p>
                  <div className="flex gap-3 mt-2 text-xs text-[#86868b]">
                    {a.article_category_name && <span>{a.article_category_name}</span>}
                    <span>{new Date(a.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </section>
  )
}
