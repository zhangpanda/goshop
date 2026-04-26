'use client'
import { Typography, Button, Card, Space, message, Statistic, Row, Col } from 'antd'
import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
export default function CachePage() {
  const [stats, setStats] = useState<Record<string,unknown>>({})
  useEffect(() => { api.get<Record<string,unknown>>('/admin/cache/stats').then(setStats).catch(()=>{}) }, [])
  return (<><Typography.Title level={4}>缓存管理</Typography.Title>
    <Card title="缓存统计" style={{marginBottom:16}}>
      <Row gutter={16}>{Object.entries(stats).map(([k,v])=><Col span={6} key={k}><Statistic title={k} value={String(v)} /></Col>)}</Row>
    </Card>
    <Space>
      <Button type="primary" danger onClick={async () => { await api.post('/admin/cache/clear'); message.success('缓存已清除') }}>清除全部缓存</Button>
    </Space>
  </>)
}
