'use client'
import { useState } from 'react'
import { Upload, Modal, message, Button, Space } from 'antd'
import { PlusOutlined, FolderOpenOutlined } from '@ant-design/icons'
import type { UploadFile } from 'antd'
import AttachmentPicker from './AttachmentPicker'

interface Props { value?: string | string[]; onChange?: (v: string | string[]) => void; max?: number }

export default function ImageUpload({ value, onChange, max = 1 }: Props) {
  const [preview, setPreview] = useState('')
  const [pickerOpen, setPickerOpen] = useState(false)
  const urls = Array.isArray(value) ? value : (value ? [value] : [])
  const fileList: UploadFile[] = urls.filter(Boolean).map((u, i) => ({ uid: String(i), name: `img-${i}`, status: 'done', url: u }))

  const addUrl = (url: string) => {
    const newUrls = [...urls, url]
    onChange?.(max === 1 ? url : newUrls)
  }

  return (
    <>
      <Space direction="vertical">
        <Upload listType="picture-card" fileList={fileList} accept="image/*"
          customRequest={async ({ file, onSuccess, onError }) => {
            const fd = new FormData(); fd.append('file', file as File)
            try {
              const token = localStorage.getItem('admin_token')
              const res = await fetch('/api/admin/upload', { method: 'POST', headers: token ? { Authorization: `Bearer ${token}` } : {}, body: fd })
              const data = await res.json()
              if (data.code !== 0) throw new Error(data.msg)
              addUrl(data.data.url); onSuccess?.(data)
            } catch (e) { message.error((e as Error).message); onError?.(e as Error) }
          }}
          onRemove={file => {
            const newUrls = urls.filter(u => u !== file.url)
            onChange?.(max === 1 ? (newUrls[0] || '') : newUrls)
          }}
          onPreview={file => setPreview(file.url || '')}
        >
          {fileList.length < max && (
            <div><PlusOutlined /><div style={{ marginTop: 4, fontSize: 12 }}>上传</div></div>
          )}
        </Upload>
        {fileList.length < max && (
          <Button size="small" icon={<FolderOpenOutlined />} onClick={() => setPickerOpen(true)}>从附件库选择</Button>
        )}
      </Space>
      <Modal open={!!preview} footer={null} onCancel={() => setPreview('')}>
        <img src={preview} alt="" style={{ width: '100%' }} />
      </Modal>
      <AttachmentPicker open={pickerOpen} onClose={() => setPickerOpen(false)}
        multiple={max > 1} onSelect={addUrl}
        onSelectMultiple={urls => { urls.forEach(addUrl) }} />
    </>
  )
}
