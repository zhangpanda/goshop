'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '@/lib/api'

interface PointLog { id: number; type: number; value: number; msg: string; created_at: string }
interface Profile { integral: number }

export default function PointsPage() {
  const [logs, setLogs] = useState<PointLog[]>([])
  const [integral, setIntegral] = useState(0)
  const [signing, setSigning] = useState(false)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get<Profile>('/user/profile').then(d => setIntegral(d.integral || 0)).catch(() => {})
    api.get<{ list: PointLog[] }>('/points/log').then(d => setLogs(d.list || d as any || [])).catch(() => {}).finally(() => setLoading(false))
  }, [])

  const sign = async () => {
    setSigning(true)
    try {
      await api.post('/points/sign')
      alert('签到成功！')
      // 刷新
      api.get<Profile>('/user/profile').then(d => setIntegral(d.integral || 0)).catch(() => {})
      api.get<{ list: PointLog[] }>('/points/log').then(d => setLogs(d.list || d as any || [])).catch(() => {})
    } catch (e: any) { alert(e.message || '今日已签到') }
    setSigning(false)
  }

  if (loading) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">加载中...</div>

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-2xl mx-auto">
        <Link href="/account" className="text-sm text-[#0071e3] hover:underline mb-6 inline-block">← 返回账户</Link>
        <h1 className="text-2xl font-semibold mb-6" style={{ color: '#1d1d1f' }}>我的积分</h1>

        <div className="bg-[#f5f5f7] rounded-2xl p-6 mb-8 flex items-center justify-between">
          <div>
            <p className="text-sm text-[#86868b]">当前积分</p>
            <p className="text-3xl font-semibold mt-1" style={{ color: '#1d1d1f' }}>{integral}</p>
          </div>
          <button onClick={sign} disabled={signing}
            className="px-6 py-2.5 bg-[#0071e3] text-white rounded-full text-sm font-medium hover:bg-blue-600 transition-colors disabled:opacity-50">
            {signing ? '签到中...' : '每日签到'}
          </button>
        </div>

        <h2 className="text-lg font-medium mb-4" style={{ color: '#1d1d1f' }}>积分记录</h2>
        {logs.length === 0 ? <p className="text-center py-10 text-[#86868b]">暂无记录</p> : (
          <div className="space-y-3">
            {logs.map(log => (
              <div key={log.id} className="flex items-center justify-between py-3 border-b border-gray-100">
                <div>
                  <p className="text-sm" style={{ color: '#1d1d1f' }}>{log.msg || '积分变动'}</p>
                  <p className="text-xs text-[#86868b] mt-0.5">{new Date(log.created_at).toLocaleString()}</p>
                </div>
                <span className={`text-sm font-medium ${log.value > 0 ? 'text-green-500' : 'text-red-400'}`}>
                  {log.value > 0 ? '+' : ''}{log.value}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </section>
  )
}
