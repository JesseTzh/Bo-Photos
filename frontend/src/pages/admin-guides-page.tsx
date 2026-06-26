import { Button, Form, Input, InputNumber, Modal, Select, Space, Switch, Table, Typography } from "antd";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAdminAssets } from "../features/assets/admin-api";
import { type Guide,useDeleteGuide,useGuides,useSaveGuide,useSortGuides } from "../features/guides/api";

export function AdminGuidesPage(){
  const q=useGuides(true),save=useSaveGuide(),remove=useDeleteGuide(),sort=useSortGuides(),nav=useNavigate();
  const assets=useAdminAssets({page:1,pageSize:200,status:"ready"});
  const[editing,setEditing]=useState<Guide|null>(),[open,setOpen]=useState(false),[form]=Form.useForm();
  function edit(g?:Guide){setEditing(g??null);form.resetFields();form.setFieldsValue(g??{days:0,published:false,sort:0});setOpen(true)}
  return <div className="admin-page-stack"><div className="panel-actions"><div><Typography.Title>Guides</Typography.Title><Typography.Paragraph type="secondary">管理发布、封面、模块、目录和关联相册。</Typography.Paragraph></div><Button type="primary" onClick={()=>edit()}>新建</Button></div>
    <Table rowKey="id" dataSource={q.data?.items} columns={[{title:"标题",dataIndex:"title"},{title:"地点",render:(_,g)=>`${g.country} · ${g.city}`},{title:"发布",render:(_,g)=>g.published?"是":"否"},{title:"操作",render:(_,g,index)=><Space><Button onClick={()=>edit(g)}>编辑资料</Button><Button onClick={()=>nav(`/admin/guides/${g.id}`)}>内容</Button><Button onClick={()=>{const next=[...(q.data?.items??[])];if(index>0){[next[index-1],next[index]]=[next[index],next[index-1]];void sort.mutateAsync(next.map(x=>x.id))}}}>↑</Button><Button onClick={()=>{const next=[...(q.data?.items??[])];if(index<next.length-1){[next[index+1],next[index]]=[next[index],next[index+1]];void sort.mutateAsync(next.map(x=>x.id))}}}>↓</Button><Button danger onClick={()=>void remove.mutateAsync(g.id)}>删除</Button></Space>}]}/>
    <Modal open={open} onCancel={()=>setOpen(false)} onOk={()=>form.submit()}><Form form={form} layout="vertical" onFinish={async v=>{await save.mutateAsync({id:editing?.id,input:{...v,start_date:v.start_date?new Date(v.start_date).toISOString():undefined,end_date:v.end_date?new Date(v.end_date).toISOString():undefined}});setOpen(false)}}><Form.Item name="title" label="标题" rules={[{required:true}]}><Input/></Form.Item><Form.Item name="country" label="国家"><Input/></Form.Item><Form.Item name="city" label="城市"><Input/></Form.Item><Form.Item name="days" label="天数"><InputNumber min={0}/></Form.Item><Form.Item name="start_date" label="开始日期"><Input type="datetime-local"/></Form.Item><Form.Item name="end_date" label="结束日期"><Input type="datetime-local"/></Form.Item><Form.Item name="cover_asset_id" label="封面"><Select allowClear options={assets.data?.items.map(a=>({value:a.id,label:a.title||a.original_name}))}/></Form.Item><Form.Item name="published" label="发布" valuePropName="checked"><Switch/></Form.Item><Form.Item name="sort" label="排序"><InputNumber/></Form.Item></Form></Modal>
  </div>
}
