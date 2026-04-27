'use client'
import CrudPage from '@/components/CrudPage'
import { Form, Input, Button, Space, Tag } from 'antd'
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons'
export default function GoodsSpecTplPage() {
  return <CrudPage title="商品规格模板" listUrl="/admin/spec-templates" createUrl="/admin/spec-templates"
    searchClient searchPlaceholder="模板名称"
    deleteUrl={r => `/admin/spec-templates/${r.id}`} updateUrl={r => `/admin/spec-templates/${r.id}`} batchDelete
    columns={[
      {title:'ID',dataIndex:'id',width:60},
      {title:'名称',dataIndex:'name'},
      {title:'规格项',dataIndex:'specs',render:(v:string)=>{
        try{const arr=JSON.parse(v||'[]');return arr.map((s:{name:string},i:number)=><Tag key={i}>{s.name}</Tag>)}catch{return '-'}
      }},
    ]}
    formItems={() => (<>
      <Form.Item name="name" label="模板名称" rules={[{required:true}]}><Input /></Form.Item>
      <Form.List name="spec_items">
        {(fields, {add, remove}) => (<>
          <div style={{marginBottom:8,fontWeight:500}}>规格项</div>
          {fields.map(f => (
            <Space key={f.key} align="baseline" style={{display:'flex',marginBottom:8}}>
              <Form.Item {...f} name={[f.name,'name']} rules={[{required:true,message:'请输入规格名'}]} noStyle>
                <Input placeholder="规格名(如颜色)" style={{width:150}} />
              </Form.Item>
              <Form.Item {...f} name={[f.name,'values']} noStyle>
                <Input placeholder="规格值(逗号分隔,如红,蓝,绿)" style={{width:250}} />
              </Form.Item>
              <MinusCircleOutlined onClick={()=>remove(f.name)} />
            </Space>
          ))}
          <Button type="dashed" onClick={()=>add()} icon={<PlusOutlined />} block>添加规格项</Button>
        </>)}
      </Form.List>
    </>)} modalWidth={600} />
}
