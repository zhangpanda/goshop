'use client'
import { useEffect, useState, Suspense } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { api, formatPrice } from '@/lib/api'

interface Order { id: number; order_no: string; status: number; pay_amount: number; payment_id: number }

function PayInner() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const orderId = searchParams.get('order_id')
  const [order, setOrder] = useState<Order | null>(null)
  const [paying, setPaying] = useState(false)
  const [paid, setPaid] = useState(false)

  useEffect(() => {
    if (!orderId) return
    api.get<Order>(`/orders/${orderId}`).then(d => {
      setOrder(d)
      if (d.status !== 0) setPaid(true)
    }).catch(() => {})
  }, [orderId])

  const doPay = async () => {
    if (!order) return
    setPaying(true)
    try {
      await api.post('/pay', { order_id: order.id, payment_id: order.payment_id })
      setPaid(true)
    } catch (e: any) {
      // 即使支付接口报错（如线下支付），也标记为已提交
      setPaid(true)
    }
    setPaying(false)
  }

  if (!orderId) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">缺少订单参数</div>
  if (!order) return <div className="min-h-screen flex items-center justify-center text-[#86868b]">加载中...</div>

  if (paid || order.status !== 0) {
    return (
      <section className="min-h-screen flex flex-col items-center justify-center px-4">
        <div className="w-16 h-16 rounded-full bg-green-100 flex items-center justify-center mb-6">
          <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#22c55e" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M20 6 9 17l-5-5"/></svg>
        </div>
        <h1 className="text-2xl font-semibold mb-2" style={{ color: '#1d1d1f' }}>{order.status !== 0 ? '订单已处理' : '支付请求已提交'}</h1>
        <p className="text-[#86868b] mb-8">订单号: {order.order_no}</p>
        <Link href={`/account/orders/${order.id}`} className="px-8 py-3 bg-[#0071e3] text-white rounded-full font-medium hover:bg-blue-600 transition-colors">查看订单</Link>
      </section>
    )
  }

  return (
    <section className="min-h-screen flex flex-col items-center justify-center px-4">
      <div className="max-w-md w-full text-center">
        <h1 className="text-2xl font-semibold mb-2" style={{ color: '#1d1d1f' }}>确认支付</h1>
        <p className="text-[#86868b] mb-8">订单号: {order.order_no}</p>
        <div className="bg-[#f5f5f7] rounded-2xl p-8 mb-8">
          <p className="text-sm text-[#86868b] mb-2">支付金额</p>
          <p className="text-4xl font-semibold" style={{ color: '#1d1d1f' }}>¥{formatPrice(order.pay_amount)}</p>
        </div>
        <button onClick={doPay} disabled={paying}
          className="w-full py-4 bg-[#0071e3] text-white rounded-full text-lg font-medium hover:bg-blue-600 transition-colors disabled:opacity-50">
          {paying ? '处理中...' : '确认支付'}
        </button>
        <button onClick={() => router.push(`/account/orders/${order.id}`)} className="mt-3 text-sm text-[#86868b] hover:text-[#1d1d1f]">稍后支付</button>
      </div>
    </section>
  )
}

export default function PayPage() {
  return <Suspense><PayInner /></Suspense>
}
