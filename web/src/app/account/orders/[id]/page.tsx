'use client'
import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { api, formatPrice } from '@/lib/api'

const statusMap: Record<number, string> = { 0: '待付款', 1: '待发货', 2: '待收货', 3: '已完成', 4: '已取消', 5: '已退款' }
const statusColor: Record<number, string> = { 0: 'text-orange-500', 1: 'text-blue-500', 2: 'text-purple-500', 3: 'text-green-500', 4: 'text-gray-400', 5: 'text-red-400' }

interface OrderItem { id: number; title: string; sku_name: string; price: number; quantity: number; image: string }
interface Order { id: number; order_no: string; status: number; pay_amount: number; total_price: number; items: OrderItem[]; created_at: string; address_name: string; address_phone: string; address_detail: string }

export default function OrderDetailPage() {
  const { id } = useParams()
  const router = useRouter()
  const [order, setOrder] = useState<Order | null>(null)
  const [loading, setLoading] = useState(true)

  const load = () => {
    api.get<Order>(`/orders/${id}`).then(setOrder).catch(() => {}).finally(() => setLoading(false))
  }
  useEffect(load, [id])

  const cancel = async () => {
    if (!confirm('确定取消订单？')) return
    try { await api.put(`/orders/${id}/cancel`); load() } catch (e: any) { alert(e.message) }
  }

  const receive = async () => {
    if (!confirm('确认已收到商品？')) return
    try { await api.put(`/orders/${id}/receive`); load() } catch (e: any) { alert(e.message) }
  }

  if (loading) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">加载中...</div>
  if (!order) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">订单不存在</div>

  const img = (src: string) => src && !src.startsWith('http') && !src.startsWith('/') ? `/${src}` : src

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-2xl mx-auto">
        <Link href="/account/orders" className="text-sm text-[#0071e3] hover:underline mb-6 inline-block">← 返回订单列表</Link>

        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-semibold" style={{ color: '#1d1d1f' }}>订单详情</h1>
          <span className={`text-lg font-medium ${statusColor[order.status]}`}>{statusMap[order.status]}</span>
        </div>

        {/* 订单信息 */}
        <div className="bg-[#f5f5f7] rounded-2xl p-5 mb-6">
          <div className="flex justify-between text-sm mb-2">
            <span className="text-[#86868b]">订单号</span>
            <span style={{ color: '#1d1d1f' }}>{order.order_no}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-[#86868b]">下单时间</span>
            <span style={{ color: '#1d1d1f' }}>{new Date(order.created_at).toLocaleString()}</span>
          </div>
        </div>

        {/* 收货地址 */}
        {order.address_name && (
          <div className="border border-gray-200 rounded-2xl p-5 mb-6">
            <h3 className="text-sm font-medium mb-2" style={{ color: '#1d1d1f' }}>收货地址</h3>
            <p className="text-sm">{order.address_name} {order.address_phone}</p>
            <p className="text-sm text-[#86868b]">{order.address_detail}</p>
          </div>
        )}

        {/* 商品列表 */}
        <div className="border border-gray-200 rounded-2xl overflow-hidden mb-6">
          <div className="p-5 space-y-4">
            {order.items?.map(item => (
              <div key={item.id} className="flex gap-4">
                <div className="w-16 h-16 bg-[#f5f5f7] rounded-xl flex-shrink-0 overflow-hidden">
                  {item.image ? <img src={img(item.image)} alt="" className="w-full h-full object-cover" /> : <div className="w-full h-full flex items-center justify-center text-xs text-gray-400">图</div>}
                </div>
                <div className="flex-1">
                  <p className="text-sm font-medium" style={{ color: '#1d1d1f' }}>{item.title}</p>
                  <p className="text-xs text-[#86868b] mt-0.5">{item.sku_name} × {item.quantity}</p>
                </div>
                <p className="text-sm font-medium" style={{ color: '#1d1d1f' }}>¥{formatPrice(item.price * item.quantity)}</p>
              </div>
            ))}
          </div>
          <div className="border-t px-5 py-4 flex justify-between items-center">
            <span className="text-sm text-[#86868b]">合计</span>
            <span className="text-xl font-semibold" style={{ color: '#1d1d1f' }}>¥{formatPrice(order.pay_amount)}</span>
          </div>
        </div>

        {/* 操作按钮 */}
        <div className="flex gap-3">
          {order.status === 0 && (
            <>
              <button onClick={() => router.push(`/checkout/pay?order_id=${order.id}`)}
                className="flex-1 py-3 bg-[#0071e3] text-white rounded-full font-medium hover:bg-blue-600 transition-colors">去支付</button>
              <button onClick={cancel}
                className="flex-1 py-3 border-2 border-gray-200 rounded-full font-medium hover:border-gray-400 transition-colors">取消订单</button>
            </>
          )}
          {order.status === 2 && (
            <button onClick={receive}
              className="flex-1 py-3 bg-[#0071e3] text-white rounded-full font-medium hover:bg-blue-600 transition-colors">确认收货</button>
          )}
        </div>
      </div>
    </section>
  )
}
