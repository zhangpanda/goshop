'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useSiteConfig } from '@/lib/site-config'
import { api } from '@/lib/api'

interface Nav { id: number; name: string; url: string }
interface FLink { id: number; name: string; url: string }

export default function Footer() {
  const { home_site_name, home_footer_info, common_app_customer_service_tel } = useSiteConfig()
  const [footerNav, setFooterNav] = useState<Nav[]>([])
  const [friendLinks, setFriendLinks] = useState<FLink[]>([])

  useEffect(() => {
    api.get<Nav[]>('/navigations?type=footer').then(d => setFooterNav(d || [])).catch(() => {})
    api.get<FLink[]>('/links').then(d => setFriendLinks((d || []).filter((l: FLink) => l.url))).catch(() => {})
  }, [])

  return (
    <footer className="bg-[#f5f5f7] border-t border-[#d2d2d7]">
      <div className="max-w-[980px] mx-auto px-4 py-12">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-8">
          <div>
            <h4 className="text-xs font-semibold text-[var(--muted)] uppercase tracking-wider mb-3">选购</h4>
            <ul className="space-y-2">
              <li><Link href="/products" className="text-sm text-[var(--fg)] hover:text-[var(--accent)] transition-colors">全部产品</Link></li>
              <li><Link href="/articles" className="text-sm text-[var(--fg)] hover:text-[var(--accent)] transition-colors">文章资讯</Link></li>
            </ul>
          </div>
          <div>
            <h4 className="text-xs font-semibold text-[var(--muted)] uppercase tracking-wider mb-3">服务</h4>
            <ul className="space-y-2">
              <li><Link href="/account/orders" className="text-sm text-[var(--fg)] hover:text-[var(--accent)] transition-colors">订单查询</Link></li>
              <li><Link href="/account/aftersale" className="text-sm text-[var(--fg)] hover:text-[var(--accent)] transition-colors">售后服务</Link></li>
            </ul>
          </div>
          {footerNav.length > 0 ? (
            <div>
              <h4 className="text-xs font-semibold text-[var(--muted)] uppercase tracking-wider mb-3">更多</h4>
              <ul className="space-y-2">
                {footerNav.map(n => (
                  <li key={n.id}><Link href={n.url} className="text-sm text-[var(--fg)] hover:text-[var(--accent)] transition-colors">{n.name}</Link></li>
                ))}
              </ul>
            </div>
          ) : (
            <div>
              <h4 className="text-xs font-semibold text-[var(--muted)] uppercase tracking-wider mb-3">关于</h4>
              <ul className="space-y-2">
                <li><Link href="/story" className="text-sm text-[var(--fg)] hover:text-[var(--accent)] transition-colors">品牌故事</Link></li>
                <li><Link href="/support" className="text-sm text-[var(--fg)] hover:text-[var(--accent)] transition-colors">联系我们</Link></li>
              </ul>
            </div>
          )}
          <div>
            <h4 className="text-xs font-semibold text-[var(--muted)] uppercase tracking-wider mb-3">联系</h4>
            {common_app_customer_service_tel
              ? <p className="text-sm text-[var(--fg)]">客服电话: {common_app_customer_service_tel}</p>
              : <p className="text-sm text-[var(--fg)]">hi@zhangpanda.com</p>}
          </div>
        </div>

        {friendLinks.length > 0 && (
          <div className="border-t border-[#d2d2d7] pt-4 pb-2 flex flex-wrap gap-4">
            <span className="text-xs text-[var(--muted)]">友情链接:</span>
            {friendLinks.map(l => (
              <a key={l.id} href={l.url} target="_blank" rel="noopener noreferrer" className="text-xs text-[var(--fg)] hover:text-[var(--accent)]">{l.name}</a>
            ))}
          </div>
        )}

        {home_footer_info && (
          <div className="border-t border-[#d2d2d7] pt-4 pb-2 text-xs text-[var(--muted)]" dangerouslySetInnerHTML={{ __html: home_footer_info }} />
        )}
        <div className={`${home_footer_info || friendLinks.length > 0 ? '' : 'border-t border-[#d2d2d7] pt-6'} text-xs text-[var(--muted)]`}>
          Copyright © {new Date().getFullYear()} {home_site_name}. All rights reserved.
        </div>
      </div>
    </footer>
  )
}
