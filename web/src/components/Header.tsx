'use client'
import Link from 'next/link'
import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useSiteConfig } from '@/lib/site-config'
import { api } from '@/lib/api'

interface Nav { id: number; name: string; url: string }

const defaultNav = [
  { name: '产品', url: '/products' },
  { name: '文章', url: '/articles' },
  { name: '支持', url: '/support' },
]

export default function Header() {
  const [open, setOpen] = useState(false)
  const [showSearch, setShowSearch] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [noticeClosed, setNoticeClosed] = useState(false)
  const [navItems, setNavItems] = useState(defaultNav)
  const { home_site_name, home_site_logo, common_shop_notice, common_app_is_enable_search } = useSiteConfig()

  useEffect(() => {
    api.get<Nav[]>('/navigations?type=header').then(d => {
      const list = d || []
      if (list.length > 0) setNavItems(list.map((n: Nav) => ({ name: n.name, url: n.url })))
    }).catch(() => {})
  }, [])

  const img = (src: string) => src && !src.startsWith('http') && !src.startsWith('/') ? `/${src}` : src
  const searchEnabled = common_app_is_enable_search !== '0'

  const doSearch = (val?: string) => {
    const q = (val ?? keyword).trim()
    if (!q) return
    setKeyword('')
    setShowSearch(false)
    window.location.href = `/products?keyword=${encodeURIComponent(q)}`
  }

  return (
    <>
      {/* 公告条 */}
      {common_shop_notice && !noticeClosed && (
        <div className="fixed top-0 w-full z-[60] bg-[#1d1d1f] text-white text-center text-xs py-2 px-8">
          <span dangerouslySetInnerHTML={{ __html: common_shop_notice }} />
          <button onClick={() => setNoticeClosed(true)} className="absolute right-3 top-1/2 -translate-y-1/2 text-white/60 hover:text-white">✕</button>
        </div>
      )}

      <header className={`fixed w-full z-50 bg-[rgba(251,251,253,0.8)] backdrop-blur-xl border-b border-[#d2d2d7]/60`} style={{ top: common_shop_notice && !noticeClosed ? 28 : 0 }}>
        <nav className="max-w-[980px] mx-auto flex items-center justify-between h-11 px-4">
          <Link href="/" className="text-sm font-semibold tracking-tight flex items-center gap-2 text-[#1d1d1f]">
            {home_site_logo ? <img src={img(home_site_logo)} alt={home_site_name} className="h-4" /> : home_site_name}
          </Link>

          <div className="hidden md:flex items-center gap-7 text-xs text-[#1d1d1f]/80">
            {navItems.map(item => (
              <Link key={item.url} href={item.url} className="hover:text-[#1d1d1f] transition-colors">{item.name}</Link>
            ))}
          </div>

          <div className="hidden md:flex items-center gap-4 text-xs">
            {searchEnabled && (
              <button onClick={() => setShowSearch(s => !s)} className="text-[#1d1d1f]/80 hover:text-[#1d1d1f] transition-colors" aria-label="搜索">
                <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
              </button>
            )}
            <Link href="/cart" className="text-[#1d1d1f]/80 hover:text-[#1d1d1f] transition-colors">
              <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>
            </Link>
            <Link href="/account" className="text-[#1d1d1f]/80 hover:text-[#1d1d1f] transition-colors">账户</Link>
          </div>

          <button className="md:hidden p-2 text-[#1d1d1f]" onClick={() => setOpen(!open)} aria-label="菜单">
            <div className="w-4 flex flex-col gap-1">
              <span className={`block h-[1.5px] bg-current transition-transform ${open ? 'rotate-45 translate-y-[3.5px]' : ''}`} />
              <span className={`block h-[1.5px] bg-current transition-transform ${open ? '-rotate-45 -translate-y-[2px]' : ''}`} />
            </div>
          </button>
        </nav>

        {/* 搜索栏 */}
        <AnimatePresence>
          {showSearch && searchEnabled && (
            <motion.div initial={{ height: 0, opacity: 0 }} animate={{ height: 'auto', opacity: 1 }} exit={{ height: 0, opacity: 0 }} className="overflow-hidden bg-white border-b">
              <div className="max-w-[600px] mx-auto px-4 py-3">
                <div className="flex items-center gap-2 bg-[#f5f5f7] rounded-full px-4 py-2">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#86868b" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
                  <input autoFocus value={keyword} onChange={e => setKeyword(e.target.value)} onKeyDown={e => e.key === 'Enter' && doSearch((e.target as HTMLInputElement).value)} placeholder="搜索商品" className="flex-1 bg-transparent outline-none text-sm" />
                  {keyword && <button onClick={() => setKeyword('')} className="text-[#86868b] text-xs">✕</button>}
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* 移动端菜单 */}
        <AnimatePresence>
          {open && (
            <motion.div initial={{ height: 0, opacity: 0 }} animate={{ height: 'auto', opacity: 1 }} exit={{ height: 0, opacity: 0 }} className="md:hidden overflow-hidden bg-white border-b">
              <div className="px-4 py-4 space-y-4">
                {searchEnabled && (
                  <div className="flex items-center gap-2 bg-[#f5f5f7] rounded-full px-4 py-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#86868b" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
                    <input value={keyword} onChange={e => setKeyword(e.target.value)} onKeyDown={e => { if (e.key === 'Enter') { doSearch((e.target as HTMLInputElement).value); setOpen(false) } }} placeholder="搜索商品" className="flex-1 bg-transparent outline-none text-sm" />
                  </div>
                )}
                {navItems.map(item => (
                  <Link key={item.url} href={item.url} className="block text-lg" onClick={() => setOpen(false)}>{item.name}</Link>
                ))}
                <Link href="/account" className="block text-lg text-[var(--muted)]" onClick={() => setOpen(false)}>账户</Link>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </header>
    </>
  )
}
