'use client'
import { motion } from 'framer-motion'

const fadeUp = (delay: number) => ({
  initial: false as const,
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true },
  style: { opacity: 0, transform: 'translateY(40px)' },
  transition: { delay, duration: 0.8 },
})

export default function StoryPage() {
  return (
    <section className="min-h-screen py-24 px-4">
      <div className="max-w-3xl mx-auto">
        <motion.p {...fadeUp(0)} className="text-sm text-center uppercase tracking-widest mb-4" style={{ color: '#86868b' }}>Our Story</motion.p>
        <motion.h1 {...fadeUp(0.15)} className="text-4xl md:text-6xl font-semibold text-center leading-tight" style={{ color: '#1d1d1f' }}>
          每一个细节，<br />都值得被认真对待。
        </motion.h1>
        <motion.div {...fadeUp(0.3)} className="mt-12">
          <img src="/story.svg" alt="品牌故事" className="w-full rounded-3xl" />
        </motion.div>
        <motion.p {...fadeUp(0.45)} className="mt-10 text-lg leading-relaxed" style={{ color: '#86868b' }}>
          我们相信好的产品不需要过多解释。从材料到工艺，从设计到体验，我们追求的是那种拿在手里就能感受到的品质。
        </motion.p>
        <motion.p {...fadeUp(0.6)} className="mt-6 text-lg leading-relaxed" style={{ color: '#86868b' }}>
          GoShop 诞生于对品质的执着追求。我们不做最便宜的，只做最值得的。每一件产品都经过反复打磨，每一次交付都力求完美。
        </motion.p>
        <motion.p {...fadeUp(0.75)} className="mt-6 text-lg leading-relaxed" style={{ color: '#86868b' }}>
          这不仅是一家商店，更是一种生活态度的表达。
        </motion.p>
      </div>
    </section>
  )
}
