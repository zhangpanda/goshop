'use client'
import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { api } from './api'

interface Admin { id: number; username: string; role_id: number }
interface AuthCtx { admin: Admin | null; login: (u: string, p: string, captchaKey: string, captchaCode: string) => Promise<void>; logout: () => void }

const Ctx = createContext<AuthCtx>({ admin: null, login: async () => {}, logout: () => {} })
export const useAdmin = () => useContext(Ctx)

export function AdminAuthProvider({ children }: { children: ReactNode }) {
  const [admin, setAdmin] = useState<Admin | null>(null)
  const [ready, setReady] = useState(false)
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    const t = localStorage.getItem('admin_token')
    const a = localStorage.getItem('admin_info')
    if (t && a) { setAdmin(JSON.parse(a)) }
    setReady(true)
  }, [])

  useEffect(() => {
    if (ready && !admin && pathname !== '/login') router.replace('/login')
  }, [ready, admin, pathname, router])

  const login = async (username: string, password: string, captchaKey: string, captchaCode: string) => {
    const res = await api.post<{ token: string; admin: Admin }>('/admin/login', { username, password, captcha_key: captchaKey, captcha_code: captchaCode })
    localStorage.setItem('admin_token', res.token)
    localStorage.setItem('admin_info', JSON.stringify(res.admin))
    setAdmin(res.admin)
    router.replace('/')
  }

  const logout = () => {
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_info')
    setAdmin(null)
    router.replace('/login')
  }

  if (!ready) return null
  return <Ctx.Provider value={{ admin, login, logout }}>{children}</Ctx.Provider>
}
