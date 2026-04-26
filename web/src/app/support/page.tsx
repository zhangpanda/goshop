'use client'
import { useSiteConfig } from '@/lib/site-config'

export default function SupportPage() {
  const { common_app_customer_service_tel, common_app_customer_service_email, common_app_customer_service_hours } = useSiteConfig()
  const email = common_app_customer_service_email || 'hi@zhangpanda.com'
  const hours = common_app_customer_service_hours || '周一至周五 9:00-18:00'

  return (
    <section className="min-h-screen py-24 px-4">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-4xl font-semibold text-center mb-12">支持与帮助</h1>

        <div className="space-y-6">
          {[
            { q: '如何查询订单？', a: '登录账户后，进入「我的订单」即可查看所有订单状态和物流信息。' },
            { q: '如何申请退换货？', a: '在订单详情中点击「申请售后」，选择退款或退货退款，提交后我们会在24小时内处理。' },
            { q: '支持哪些支付方式？', a: '目前支持微信支付。更多支付方式即将上线。' },
            { q: '如何联系客服？', a: `发送邮件至 ${email}，我们会在工作日24小时内回复。${common_app_customer_service_tel ? `也可拨打客服电话 ${common_app_customer_service_tel}。` : ''}` },
          ].map((item, i) => (
            <details key={i} className="group border border-gray-200 rounded-2xl overflow-hidden">
              <summary className="flex items-center justify-between p-5 cursor-pointer font-medium hover:bg-[var(--surface)] transition-colors">
                {item.q}
                <span className="text-[var(--muted)] group-open:rotate-45 transition-transform text-xl">+</span>
              </summary>
              <p className="px-5 pb-5 text-[var(--muted)] leading-relaxed">{item.a}</p>
            </details>
          ))}
        </div>

        <div className="mt-16 text-center">
          <h2 className="text-2xl font-semibold mb-4">联系我们</h2>
          {common_app_customer_service_tel && <p className="text-[var(--muted)]">客服电话：{common_app_customer_service_tel}</p>}
          <p className="text-[var(--muted)] mt-1">邮箱：{email}</p>
          <p className="text-[var(--muted)] mt-1">工作时间：{hours}</p>
        </div>
      </div>
    </section>
  )
}
