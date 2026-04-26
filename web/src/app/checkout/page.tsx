'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { api, formatPrice } from '@/lib/api'

interface Address { id: number; name: string; phone: string; province: string; city: string; district: string; detail: string; is_default: boolean }
interface CartItem { id: number; quantity: number; sku: { price: number; name: string }; goods: { title: string } }
interface Payment { id: number; name: string; payment: string; logo: string }

export default function Checkout() {
  const router = useRouter()
  const [addresses, setAddresses] = useState<Address[]>([])
  const [items, setItems] = useState<CartItem[]>([])
  const [payments, setPayments] = useState<Payment[]>([])
  const [selectedAddr, setSelectedAddr] = useState<number>(0)
  const [selectedPay, setSelectedPay] = useState<number>(0)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    api.get<Address[]>('/address').then(d => { setAddresses(d); const def = d.find(a => a.is_default); if (def) setSelectedAddr(def.id) }).catch(() => {})
    api.get<CartItem[]>('/cart').then(setItems).catch(() => {})
    api.get<Payment[]>('/payments').then(d => { const list = d || []; setPayments(list); if (list[0]) setSelectedPay(list[0].id) }).catch(() => {})
  }, [])

  const total = items.reduce((s, i) => s + i.sku.price * i.quantity, 0)

  const submit = async () => {
    if (!selectedAddr) { alert('请选择收货地址'); return }
    if (!selectedPay) { alert('请选择支付方式'); return }
    setSubmitting(true)
    try {
      const res = await api.post<{ id: number }>('/orders', { address_id: selectedAddr, cart_ids: items.map(i => i.id), order_model: 0, payment_id: selectedPay })
      router.push(`/checkout/pay?order_id=${res.id}`)
    } catch (e: any) { alert(e.message) }
    setSubmitting(false)
  }

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-3xl font-semibold mb-10">结算</h1>

        {/* 地址 */}
        <div className="mb-8">
          <h2 className="text-lg font-medium mb-4">收货地址</h2>
          {addresses.length === 0 ? <p className="text-[var(--muted)]">暂无地址，请先添加</p> : (
            <div className="space-y-3">{addresses.map(addr => (
              <button key={addr.id} onClick={() => setSelectedAddr(addr.id)}
                className={`w-full text-left p-4 rounded-2xl border-2 transition-all ${selectedAddr === addr.id ? 'border-[var(--accent)] bg-blue-50' : 'border-gray-200'}`}>
                <p className="font-medium">{addr.name} <span className="text-[var(--muted)] font-normal ml-2">{addr.phone}</span></p>
                <p className="text-sm text-[var(--muted)] mt-1">{addr.province}{addr.city}{addr.district} {addr.detail}</p>
              </button>
            ))}</div>
          )}
        </div>

        {/* 支付方式 */}
        <div className="mb-8">
          <h2 className="text-lg font-medium mb-4">支付方式</h2>
          <div className="space-y-3">
            {payments.map(p => (
              <button key={p.id} onClick={() => setSelectedPay(p.id)}
                className={`w-full text-left p-4 rounded-2xl border-2 transition-all flex items-center gap-3 ${selectedPay === p.id ? 'border-[var(--accent)] bg-blue-50' : 'border-gray-200'}`}>
                <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${selectedPay === p.id ? 'border-[#0071e3]' : 'border-gray-300'}`}>
                  {selectedPay === p.id && <div className="w-2.5 h-2.5 rounded-full bg-[#0071e3]" />}
                </div>
                <span className="font-medium">{p.name}</span>
              </button>
            ))}
            {payments.length === 0 && <p className="text-[var(--muted)]">暂无可用支付方式</p>}
          </div>
        </div>

        {/* 商品 */}
        <div className="mb-8">
          <h2 className="text-lg font-medium mb-4">商品清单</h2>
          <div className="divide-y">{items.map(item => (
            <div key={item.id} className="flex justify-between py-3">
              <div><p className="font-medium">{item.goods.title}</p><p className="text-sm text-[var(--muted)]">{item.sku.name} × {item.quantity}</p></div>
              <p className="font-medium">¥{formatPrice(item.sku.price * item.quantity)}</p>
            </div>
          ))}</div>
        </div>

        {/* 提交 */}
        <div className="border-t pt-6">
          <div className="flex justify-between text-xl font-semibold mb-6"><span>合计</span><span>¥{formatPrice(total)}</span></div>
          <button onClick={submit} disabled={submitting}
            className="w-full py-4 bg-[var(--accent)] text-white rounded-full text-lg font-medium hover:bg-blue-600 transition-colors disabled:opacity-50">
            {submitting ? '提交中...' : '提交订单'}
          </button>
        </div>
      </div>
    </section>
  )
}
