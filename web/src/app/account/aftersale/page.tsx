'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api, formatPrice } from '@/lib/api'

const statusMap: Record<number, string> = { 0: '待审核', 1: '已同意', 2: '已拒绝', 3: '已退货', 4: '已完成', 5: '已取消' }
const statusColor: Record<number, string> = { 0: 'text-orange-500', 1: 'text-blue-500', 2: 'text-red-400', 3: 'text-purple-500', 4: 'text-green-500', 5: 'text-gray-400' }

interface Aftersale { id: number; order_no: string; status: number; type: number; reason: string; price: number; created_at: string }

export default function AftersalePage() {
  const [list, setList] = useState<Aftersale[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get<{ list: Aftersale[] }>('/aftersale').then(d => setList(d.list || d as any || [])).catch(() => {}).finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">加载中...</div>

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-2xl mx-auto">
        <Link href="/account" className="text-sm text-[#0071e3] hover:underline mb-6 inline-block">← 返回账户</Link>
        <h1 className="text-2xl font-semibold mb-8" style={{ color: '#1d1d1f' }}>订单售后</h1>

        {list.length === 0 ? <p className="text-center py-20 text-[#86868b]">暂无售后记录</p> : (
          <div className="space-y-4">
            {list.map(item => (
              <div key={item.id} className="border border-gray-200 rounded-2xl p-5">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-[#86868b]">订单: {item.order_no}</span>
                  <span className={`text-sm font-medium ${statusColor[item.status] || ''}`}>{statusMap[item.status] || '未知'}</span>
                </div>
                <p className="text-sm" style={{ color: '#1d1d1f' }}>类型: {item.type === 0 ? '仅退款' : '退货退款'}</p>
                <p className="text-sm text-[#86868b] mt-1">原因: {item.reason}</p>
                {item.price > 0 && <p className="text-sm mt-1">退款金额: <span className="font-medium">¥{formatPrice(item.price)}</span></p>}
                <p className="text-xs text-[#86868b] mt-2">{new Date(item.created_at).toLocaleString()}</p>
              </div>
            ))}
          </div>
        )}
      </div>
    </section>
  )
}
