'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { motion } from 'framer-motion'
import { api, formatPrice } from '@/lib/api'

const statusMap: Record<number, string> = { 0: '待付款', 1: '待发货', 2: '待收货', 3: '已完成', 4: '已取消', 5: '已退款' }
const statusColor: Record<number, string> = { 0: 'text-orange-500', 1: 'text-blue-500', 2: 'text-purple-500', 3: 'text-green-500', 4: 'text-gray-400', 5: 'text-red-400' }

interface OrderItem { id: number; title: string; sku_name: string; price: number; quantity: number; image: string }
interface Order { id: number; order_no: string; status: number; pay_amount: number; items: OrderItem[]; created_at: string }

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get<{ list: Order[] }>('/orders').then(d => setOrders(d.list || [])).catch(() => {}).finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="min-h-screen flex items-center justify-center text-[var(--muted)]">
        <Link href="/account" className="inline-flex items-center text-sm text-[#0071e3] hover:underline mb-6">← 返回账户</Link>加载中...</div>

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-3xl font-semibold mb-10">我的订单</h1>

        {orders.length === 0 ? (
          <div className="text-center py-20">
            <p className="text-[var(--muted)]">暂无订单</p>
            <Link href="/products" className="inline-block mt-4 text-[var(--accent)] hover:underline">去选购 →</Link>
          </div>
        ) : (
          <div className="space-y-6">
            {orders.map((order, i) => (
              <motion.div
                key={order.id}
                initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.05 }}
                className="border border-gray-200 rounded-2xl overflow-hidden"
              >
                <div className="flex items-center justify-between px-5 py-3 bg-[var(--surface)]">
                  <span className="text-sm text-[var(--muted)]">{order.order_no}</span>
                  <span className={`text-sm font-medium ${statusColor[order.status]}`}>{statusMap[order.status]}</span>
                </div>
                <div className="p-5">
                  {order.items?.map(item => (
                    <div key={item.id} className="flex gap-3 mb-3 last:mb-0">
                      <div className="w-14 h-14 bg-gray-100 rounded-xl flex-shrink-0 flex items-center justify-center overflow-hidden">
                        {item.image ? <img src={item.image} alt="" className="w-full h-full object-cover" /> : <span className="text-xs text-gray-400">图</span>}
                      </div>
                      <div className="flex-1">
                        <p className="text-sm font-medium">{item.title}</p>
                        <p className="text-xs text-[var(--muted)]">{item.sku_name} × {item.quantity}</p>
                      </div>
                      <p className="text-sm">¥{formatPrice(item.price * item.quantity)}</p>
                    </div>
                  ))}
                  <div className="flex justify-between items-center pt-3 border-t mt-3">
                    <span className="text-sm text-[var(--muted)]">{new Date(order.created_at).toLocaleDateString()}</span>
                    <div className="flex items-center gap-4">
                      <span className="font-semibold">¥{formatPrice(order.pay_amount)}</span>
                      <Link href={`/account/orders/${order.id}`} className="text-sm text-[#0071e3] hover:underline">查看详情</Link>
                    </div>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        )}
      </div>
    </section>
  )
}
