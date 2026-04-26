'use client'
import { useEffect, useState, ReactNode } from 'react'
import { Form, Input, InputNumber, Switch, Select, Button, Typography, Card, message } from 'antd'
import { api } from '@/lib/api'
import ImageUpload from '@/components/ImageUpload'

const IMG = new Set(['home_site_logo','home_site_logo_wap','home_site_logo_square','home_site_favicon','admin_logo'])
const SW = new Set(['home_site_web_state','common_register_is_enable_audit','common_img_verify_state','common_data_is_use_cache','admin_login_img_verify_state','common_app_is_enable_search','common_app_is_enable_answer','common_app_is_online_service','common_app_is_header_nav_fixed','common_app_is_use_mobile_detail','common_app_mini_weixin_upload_shipping_status'])
const NUM = new Set(['common_verify_interval_time','common_verify_expire_time','common_order_close_limit_time','common_order_success_limit_time','common_goods_give_integral_limit_time','home_order_aftersale_return_launch_day','home_max_limit_image','home_max_limit_file','home_max_limit_video','common_cache_data_redis_port','common_cache_data_redis_expire','common_email_smtp_port','common_page_size','home_content_max_width'])
const TA = new Set(['common_shop_notice','admin_notice','home_footer_info','home_statistics_code','home_site_close_reason','home_order_aftersale_return_only_money_reason','home_order_aftersale_return_money_goods_reason','home_seo_site_description','home_email_user_reg_template','home_email_user_forget_pwd_template','home_email_user_email_bind_template','home_sms_user_reg_template','home_sms_user_forget_pwd_template','home_sms_user_mobile_bind_template'])
const PW = new Set(['common_email_smtp_pwd','common_sms_secret','common_app_mini_weixin_appsecret','common_app_mini_alipay_rsa_private','common_cache_data_redis_password'])

function field(k: string): ReactNode {
  if (IMG.has(k)) return <ImageUpload />
  if (SW.has(k)) return <Switch checkedChildren="开" unCheckedChildren="关" />
  if (NUM.has(k)) return <InputNumber min={0} style={{ width: '100%' }} />
  if (TA.has(k)) return <Input.TextArea rows={4} />
  if (PW.has(k)) return <Input.Password />
  return <Input />
}

interface CI { key: string; value: string; desc: string }
export default function Page() {
  const [items, setItems] = useState<CI[]>([])
  const [form] = Form.useForm()
  useEffect(() => { api.get<CI[]>('/admin/config?group=seo').then(r => setItems(Array.isArray(r)?r:[])).catch(()=>setItems([])) }, [])
  useEffect(() => {
    const v: Record<string,unknown> = {}
    items.forEach(i => { if(SW.has(i.key)) v[i.key]=i.value==='1'; else if(NUM.has(i.key)) v[i.key]=i.value?Number(i.value):undefined; else v[i.key]=i.value })
    form.setFieldsValue(v)
  }, [items, form])
  return (<><Typography.Title level={4}>SEO设置</Typography.Title><Card><Form form={form} layout="vertical">
    {items.map(i => <Form.Item key={i.key} name={i.key} label={<>{i.desc||i.key} <span style={{color:'#aaa',fontSize:11}}>({i.key})</span></>} valuePropName={SW.has(i.key)?'checked':'value'}>{field(i.key)}</Form.Item>)}
    {items.length>0 && <Form.Item><Button type="primary" size="large" onClick={async()=>{
      const raw=form.getFieldsValue(); const configs:Record<string,string>={}
      for(const[k,v] of Object.entries(raw)){if(SW.has(k))configs[k]=v?'1':'0';else configs[k]=v==null?'':String(v)}
      await api.post('/admin/config',{group:'seo',configs}); message.success('保存成功')
    }}>保存配置</Button></Form.Item>}
  </Form></Card></>)
}
