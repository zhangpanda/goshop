'use client'
import { useRef, useEffect } from 'react'
import { Button, Space } from 'antd'
import { BoldOutlined, ItalicOutlined, UnorderedListOutlined, OrderedListOutlined, PictureOutlined } from '@ant-design/icons'

interface Props { value?: string; onChange?: (v: string) => void; height?: number }

export default function RichEditor({ value, onChange, height = 300 }: Props) {
  const ref = useRef<HTMLDivElement>(null)
  const init = useRef(false)

  useEffect(() => {
    if (ref.current && !init.current && value) { ref.current.innerHTML = value; init.current = true }
  }, [value])

  const exec = (cmd: string, val?: string) => { document.execCommand(cmd, false, val); ref.current?.focus() }
  const insertImage = () => {
    const url = prompt('输入图片URL')
    if (url) exec('insertImage', url)
  }

  return (
    <div style={{ border: '1px solid #d9d9d9', borderRadius: 6 }}>
      <Space style={{ padding: '4px 8px', borderBottom: '1px solid #d9d9d9', background: '#fafafa' }} size={4}>
        <Button type="text" size="small" icon={<BoldOutlined />} onClick={() => exec('bold')} />
        <Button type="text" size="small" icon={<ItalicOutlined />} onClick={() => exec('italic')} />
        <Button type="text" size="small" icon={<UnorderedListOutlined />} onClick={() => exec('insertUnorderedList')} />
        <Button type="text" size="small" icon={<OrderedListOutlined />} onClick={() => exec('insertOrderedList')} />
        <Button type="text" size="small" icon={<PictureOutlined />} onClick={insertImage} />
      </Space>
      <div ref={ref} contentEditable suppressContentEditableWarning
        style={{ minHeight: height, padding: 12, outline: 'none', overflow: 'auto' }}
        onInput={() => onChange?.(ref.current?.innerHTML || '')}
      />
    </div>
  )
}
