'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, Button, Space, Tag } from 'antd'
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons'
export default function GoodsParamsTplPage() {
  return <CrudPage title="商品参数模板" listUrl="/admin/params-templates" createUrl="/admin/params-templates"
    deleteUrl={r => `/admin/params-templates/${r.id}`} updateUrl={r => `/admin/params-templates/${r.id}`} batchDelete
    columns={[
      {title:'ID',dataIndex:'id',width:60},
      {title:'名称',dataIndex:'name'},
      {title:'参数项',dataIndex:'params',render:(v:string)=>{
        try{const arr=JSON.parse(v||'[]');return arr.map((p:{name:string},i:number)=><Tag key={i}>{p.name}</Tag>)}catch{return '-'}
      }},
    ]}
    formItems={() => (<>
      <Form.Item name="name" label="模板名称" rules={[{required:true}]}><Input /></Form.Item>
      <Form.List name="param_items">
        {(fields, {add, remove}) => (<>
          <div style={{marginBottom:8,fontWeight:500}}>参数项</div>
          {fields.map(f => (
            <Space key={f.key} align="baseline" style={{display:'flex',marginBottom:8}}>
              <Form.Item {...f} name={[f.name,'name']} rules={[{required:true,message:'请输入参数名'}]} noStyle>
                <Input placeholder="参数名(如品牌)" style={{width:150}} />
              </Form.Item>
              <Form.Item {...f} name={[f.name,'value']} noStyle>
                <Input placeholder="默认值(可选)" style={{width:200}} />
              </Form.Item>
              <MinusCircleOutlined onClick={()=>remove(f.name)} />
            </Space>
          ))}
          <Button type="dashed" onClick={()=>add()} icon={<PlusOutlined />} block>添加参数项</Button>
        </>)}
      </Form.List>
    </>)} modalWidth={600} />
}
