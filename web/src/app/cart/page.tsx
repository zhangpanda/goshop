'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { motion } from 'framer-motion'
import { api, formatPrice } from '@/lib/api'

interface CartItem { id: number; quantity: number; goods: { id: number; title: string; main_image: string }; sku: { id: number; name: string; price: number } }

export default function CartPage() {
  const [items, setItems] = useState<CartItem[]>([])
  const [loading, setLoading] = useState(true)

  const load = () => { api.get<CartItem[]>('/cart').then(setItems).catch(() => {}).finally(() => setLoading(false)) }
  useEffect(load, [])

  const remove = async (ids: number[]) => {
    await api.del('/cart', { ids })
    load()
  }

  const updateQty = async (id: number, quantity: number) => {
    if (quantity < 1) return
    await api.put(`/cart/${id}`, { quantity })
    setItems(prev => prev.map(i => i.id === id ? { ...i, quantity } : i))
  }

  const total = items.reduce((s, i) => s + i.sku.price * i.quantity, 0)

  if (loading) return <div className="min-h-screen flex items-center justify-center text-[var(--muted)]">加载中...</div>

  return (
    <section className="min-h-screen py-20 px-4">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-3xl md:text-4xl font-semibold mb-10">购物袋</h1>

        {items.length === 0 ? (
          <div className="text-center py-20">
            <p className="text-[var(--muted)] text-lg">你的购物袋是空的</p>
            <Link href="/products" className="inline-block mt-6 text-[var(--accent)] hover:underline">去选购 →</Link>
          </div>
        ) : (
          <>
            <div className="divide-y">
              {items.map((item, i) => (
                <motion.div
                  key={item.id}
                  initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: i * 0.05 }}
                  className="flex gap-4 py-6"
                >
                  <div className="w-20 h-20 bg-[var(--surface)] rounded-xl flex-shrink-0 flex items-center justify-center overflow-hidden">
                    {item.goods.main_image
                      ? <img src={item.goods.main_image} alt="" className="w-full h-full object-cover" />
                      : <span className="text-xs text-gray-400">图片</span>}
                  </div>
                  <div className="flex-1 min-w-0">
                    <Link href={`/products/${item.goods.id}`} className="font-medium hover:text-[var(--accent)]">{item.goods.title}</Link>
                    <p className="text-sm text-[var(--muted)] mt-1">{item.sku.name}</p>
                    <div className="flex items-center gap-3 mt-2">
                      <button onClick={() => updateQty(item.id, item.quantity - 1)}
                        className="w-7 h-7 rounded-full border border-gray-300 flex items-center justify-center text-sm hover:bg-gray-100"
                        disabled={item.quantity <= 1}>−</button>
                      <span className="text-sm w-6 text-center">{item.quantity}</span>
                      <button onClick={() => updateQty(item.id, item.quantity + 1)}
                        className="w-7 h-7 rounded-full border border-gray-300 flex items-center justify-center text-sm hover:bg-gray-100">+</button>
                      <button onClick={() => remove([item.id])}
                        className="ml-auto text-xs text-red-500 hover:text-red-700">移除</button>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-medium">¥{formatPrice(item.sku.price * item.quantity)}</p>
                  </div>
                </motion.div>
              ))}
            </div>

            <div className="border-t pt-6 mt-6">
              <div className="flex justify-between text-lg font-semibold">
                <span>合计</span>
                <span>¥{formatPrice(total)}</span>
              </div>
              <Link
                href="/checkout"
                className="block mt-6 w-full py-4 bg-[var(--accent)] text-white text-center rounded-full text-lg font-medium hover:bg-blue-600 transition-colors"
              >
                结算
              </Link>
            </div>
          </>
        )}
      </div>
    </section>
  )
}
