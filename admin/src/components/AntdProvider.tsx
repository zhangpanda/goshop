'use client'
import '@ant-design/v5-patch-for-react-19'
import { ReactNode } from 'react'

export default function AntdProvider({ children }: { children: ReactNode }) {
  return <>{children}</>
}
