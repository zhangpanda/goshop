'use client'
import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { api } from '@/lib/api'

interface CustomView { id: number; title: string; content: string }

export default function CustomViewPage() {
  const { id } = useParams()
  const [page, setPage] = useState<CustomView | null>(null)

  useEffect(() => {
    // 获取自定义页面列表，找到对应id的
    api.get<{ list: CustomView[] }>('/custom-views').then(d => {
      const found = (d.list || d as any || []).find((p: CustomView) => String(p.id) === String(id))
      if (found) setPage(found)
    }).catch(() => {})
  }, [id])

  if (!page) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">加载中...</div>

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-3xl mx-auto">
        <Link href="/" className="text-sm text-[#0071e3] hover:underline mb-6 inline-block">← 返回首页</Link>
        <h1 className="text-3xl font-semibold mb-8" style={{ color: '#1d1d1f' }}>{page.title}</h1>
        <div className="prose prose-lg max-w-none" dangerouslySetInnerHTML={{ __html: page.content }} />
      </div>
    </section>
  )
}
