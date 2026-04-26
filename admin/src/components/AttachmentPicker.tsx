'use client'
import { useState, useEffect } from 'react'
import { Modal, Upload, Tabs, Card, Row, Col, Image, Button, message, Pagination, Empty } from 'antd'
import { PlusOutlined, CheckCircleFilled } from '@ant-design/icons'
import { api } from '@/lib/api'

interface Attachment { id: number; name: string; path: string; ext: string }
interface Props { open: boolean; onClose: () => void; onSelect: (url: string) => void; multiple?: boolean; onSelectMultiple?: (urls: string[]) => void }

export default function AttachmentPicker({ open, onClose, onSelect, multiple, onSelectMultiple }: Props) {
  const [list, setList] = useState<Attachment[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState<string[]>([])

  const load = async (p = 1) => {
    const r = await api.get<{ total: number; list: Attachment[] }>(`/admin/attachments?page=${p}&page_size=24`)
    setList(r.list || []); setTotal(r.total); setPage(p)
  }

  useEffect(() => { if (open) { load(); setSelected([]) } }, [open])

  const toggle = (url: string) => {
    if (multiple) {
      setSelected(prev => prev.includes(url) ? prev.filter(u => u !== url) : [...prev, url])
    } else {
      onSelect(url); onClose()
    }
  }

  const confirm = () => {
    if (multiple && onSelectMultiple) { onSelectMultiple(selected); onClose() }
    else if (selected.length) { onSelect(selected[0]); onClose() }
  }

  const uploadProps = {
    accept: 'image/*',
    showUploadList: false,
    customRequest: async ({ file }: { file: unknown }) => {
      const fd = new FormData(); fd.append('file', file as File)
      const token = localStorage.getItem('admin_token')
      const res = await fetch('/api/upload', { method: 'POST', headers: token ? { Authorization: `Bearer ${token}` } : {}, body: fd })
      const data = await res.json()
      if (data.code === 0) { message.success('上传成功'); load(1) }
      else message.error(data.msg)
    },
  }

  return (
    <Modal title="选择图片" open={open} onCancel={onClose} width={800} onOk={confirm}
      footer={multiple ? undefined : null}>
      <Tabs items={[
        { key: 'lib', label: '附件库', children: (
          <>
            <Row gutter={[8, 8]}>
              {list.filter(a => /\.(jpg|jpeg|png|gif|webp|svg)$/i.test(a.path)).map(a => (
                <Col span={4} key={a.id}>
                  <div onClick={() => toggle(a.path)} style={{ cursor: 'pointer', border: selected.includes(a.path) ? '2px solid #1890ff' : '2px solid transparent', borderRadius: 4, padding: 2, position: 'relative' }}>
                    <Image src={a.path} alt={a.name} preview={false} style={{ width: '100%', height: 80, objectFit: 'cover', borderRadius: 4 }} />
                    {selected.includes(a.path) && <CheckCircleFilled style={{ position: 'absolute', top: 4, right: 4, color: '#1890ff', fontSize: 18 }} />}
                  </div>
                </Col>
              ))}
            </Row>
            {!list.length && <Empty description="暂无附件" />}
            <div style={{ textAlign: 'center', marginTop: 12 }}>
              <Pagination current={page} total={total} pageSize={24} onChange={p => load(p)} size="small" />
            </div>
          </>
        )},
        { key: 'upload', label: '上传图片', children: (
          <Upload.Dragger {...uploadProps}>
            <p><PlusOutlined style={{ fontSize: 32, color: '#999' }} /></p>
            <p>点击或拖拽上传图片</p>
          </Upload.Dragger>
        )},
      ]} />
    </Modal>
  )
}
