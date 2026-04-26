import type { Metadata } from 'next'
import AntdProvider from '@/components/AntdProvider'

export const metadata: Metadata = { title: 'GoShop 管理后台' }

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN">
      <body style={{ margin: 0 }}><AntdProvider>{children}</AntdProvider></body>
    </html>
  )
}
