import { useEffect, useState, useRef } from 'react'
import { api } from './api'

interface UserInfo { id: number; username: string; nickname: string; phone: string }

const cache: Record<number, UserInfo> = {}
const pending = new Set<number>()
let timer: ReturnType<typeof setTimeout> | null = null
let listeners: (() => void)[] = []

function flush() {
  const ids = [...pending]
  pending.clear()
  if (!ids.length) return
  api.get<{ list: UserInfo[] }>(`/admin/users?ids=${ids.join(',')}&page_size=100`).then(r => {
    const list = r?.list || (Array.isArray(r) ? r as unknown as UserInfo[] : [])
    list.forEach(u => { cache[u.id] = u })
    listeners.forEach(fn => fn())
  })
}

function request(ids: number[]) {
  ids.forEach(id => { if (id > 0 && !cache[id]) pending.add(id) })
  if (pending.size > 0) {
    if (timer) clearTimeout(timer)
    timer = setTimeout(flush, 50) // batch within 50ms
  }
}

/** 批量解析 user_id → 用户名。传入 ID 数组，返回 id→name 映射 */
export function useUserMap(ids: number[]): Record<number, string> {
  const [, setTick] = useState(0)
  const ref = useRef<(() => void) | undefined>(undefined)
  ref.current = () => setTick(t => t + 1)

  useEffect(() => {
    const fn = () => ref.current?.()
    listeners.push(fn)
    return () => { listeners = listeners.filter(f => f !== fn) }
  }, [])

  useEffect(() => { request(ids) }, [ids.join(',')])  // eslint-disable-line

  const map: Record<number, string> = {}
  ids.forEach(id => {
    if (!id) return
    const u = cache[id]
    map[id] = u ? (u.nickname || u.username) : `#${id}`
  })
  return map
}

/** 同上，用于商品ID解析 */
const goodsCache: Record<number, string> = {}
export function useGoodsMap(ids: number[]): Record<number, string> {
  const [, setTick] = useState(0)
  useEffect(() => {
    const missing = ids.filter(id => id > 0 && !goodsCache[id])
    if (!missing.length) return
    api.get<{ list: { id: number; title: string }[] }>(`/admin/goods?ids=${missing.join(',')}&page_size=100`).then(r => {
      const list = r?.list || (Array.isArray(r) ? r as unknown as { id: number; title: string }[] : [])
      list.forEach(g => { goodsCache[g.id] = g.title })
      setTick(t => t + 1)
    })
  }, [ids.join(',')])  // eslint-disable-line

  const map: Record<number, string> = {}
  ids.forEach(id => { map[id] = id > 0 ? (goodsCache[id] || `#${id}`) : '-' })
  return map
}
