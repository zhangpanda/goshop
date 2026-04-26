'use client'
import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'

interface Article { id: number; title: string; content: string; cover: string; created_at: string; article_category_name: string }

export default function ArticleDetailPage() {
  const { id } = useParams()
  const [article, setArticle] = useState<Article | null>(null)

  useEffect(() => {
    api.get<Article>(`/articles/${id}`).then(setArticle).catch(() => {})
  }, [id])

  if (!article) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">加载中...</div>

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-3xl mx-auto">
        <Link href="/articles" className="text-sm text-[#0071e3] hover:underline mb-6 inline-block">← 返回文章列表</Link>
        <h1 className="text-3xl font-semibold mb-4" style={{ color: '#1d1d1f' }}>{article.title}</h1>
        <div className="flex gap-3 text-sm text-[#86868b] mb-8">
          {article.article_category_name && <span>{article.article_category_name}</span>}
          <span>{new Date(article.created_at).toLocaleDateString()}</span>
        </div>
        <div className="prose prose-lg max-w-none" dangerouslySetInnerHTML={{ __html: article.content }} />
      </div>
    </section>
  )
}
