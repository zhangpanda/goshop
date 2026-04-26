'use client'
import { Drawer, Descriptions } from 'antd'
import type { ReactNode } from 'react'

interface Item { label: string; value: ReactNode }
interface Props { open: boolean; onClose: () => void; title: string; items: Item[]; width?: number; extra?: ReactNode }

export default function DetailDrawer({ open, onClose, title, items, width = 640, extra }: Props) {
  return (
    <Drawer title={title} open={open} onClose={onClose} width={width} extra={extra}>
      <Descriptions column={1} bordered size="small">
        {items.map((item, i) => (
          <Descriptions.Item key={i} label={item.label}>{item.value ?? '-'}</Descriptions.Item>
        ))}
      </Descriptions>
    </Drawer>
  )
}
