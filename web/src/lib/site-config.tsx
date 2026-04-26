'use client'
import { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { api } from '@/lib/api'

interface SiteConfig {
  // base
  home_site_name: string
  home_site_logo: string
  home_site_logo_wap: string
  home_site_favicon: string
  home_footer_info: string
  home_statistics_code: string
  common_shop_notice: string
  // seo
  home_seo_site_title: string
  home_seo_site_keywords: string
  home_seo_site_description: string
  // site
  home_site_web_state: string
  home_site_close_reason: string
  home_user_login_type: string
  home_user_reg_type: string
  // app
  common_app_customer_service_tel: string
  common_app_customer_service_email: string
  common_app_customer_service_hours: string
  common_app_is_enable_search: string
}

const defaults: SiteConfig = {
  home_site_name: 'GoShop',
  home_site_logo: '',
  home_site_logo_wap: '',
  home_site_favicon: '',
  home_footer_info: '',
  home_statistics_code: '',
  common_shop_notice: '',
  home_seo_site_title: '',
  home_seo_site_keywords: '',
  home_seo_site_description: '',
  home_site_web_state: '1',
  home_site_close_reason: '',
  home_user_login_type: 'username',
  home_user_reg_type: 'username',
  common_app_customer_service_tel: '',
  common_app_customer_service_email: '',
  common_app_customer_service_hours: '',
  common_app_is_enable_search: '1',
}

const SiteContext = createContext<SiteConfig>(defaults)

export function SiteProvider({ children }: { children: ReactNode }) {
  const [config, setConfig] = useState<SiteConfig>(defaults)
  const [closed, setClosed] = useState(false)
  const [ready, setReady] = useState(false)

  useEffect(() => {
    api.get<Record<string, string>>('/site-config').then(d => {
      const c = { ...defaults, ...d }
      setConfig(c)

      // 页面标题
      if (c.home_seo_site_title) document.title = c.home_seo_site_title

      // favicon
      if (c.home_site_favicon) {
        let link = document.querySelector("link[rel~='icon']") as HTMLLinkElement
        if (!link) { link = document.createElement('link'); link.rel = 'icon'; document.head.appendChild(link) }
        link.href = c.home_site_favicon.startsWith('http') || c.home_site_favicon.startsWith('/') ? c.home_site_favicon : `/${c.home_site_favicon}`
      }

      // SEO meta
      const setMeta = (name: string, content: string) => {
        if (!content) return
        let el = document.querySelector(`meta[name="${name}"]`) as HTMLMetaElement
        if (!el) { el = document.createElement('meta'); el.name = name; document.head.appendChild(el) }
        el.content = content
      }
      setMeta('keywords', c.home_seo_site_keywords)
      setMeta('description', c.home_seo_site_description)

      // 统计代码
      if (c.home_statistics_code) {
        const div = document.createElement('div')
        div.innerHTML = c.home_statistics_code
        div.querySelectorAll('script').forEach(s => {
          const ns = document.createElement('script')
          if (s.src) ns.src = s.src; else ns.textContent = s.textContent
          document.body.appendChild(ns)
        })
      }

      // 网站关闭
      if (c.home_site_web_state === '0') setClosed(true)
      setReady(true)
    }).catch(() => setReady(true))
  }, [])

  if (!ready) return null

  // 网站关闭状态
  if (closed) {
    return (
      <div className="min-h-screen flex items-center justify-center px-4">
        <div className="text-center">
          <h1 className="text-2xl font-semibold mb-4" style={{ color: '#1d1d1f' }}>{config.home_site_name}</h1>
          <p className="text-[#86868b]">{config.home_site_close_reason || '网站维护中，请稍后访问'}</p>
        </div>
      </div>
    )
  }

  return <SiteContext.Provider value={config}>{children}</SiteContext.Provider>
}

export const useSiteConfig = () => useContext(SiteContext)
