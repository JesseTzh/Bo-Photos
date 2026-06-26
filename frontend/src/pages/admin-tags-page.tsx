import { DeleteOutlined, EditOutlined, PlusOutlined } from "@ant-design/icons";
import { Button, Form, Input, Modal, Select, Space, Tree, Typography, message } from "antd";
import { useMemo, useState } from "react";
import { type TagNode, flattenTags, useCreateTag, useDeleteTag, useMoveTag, useTags, useUpdateTag } from "../features/tags/api";

export function AdminTagsPage() {
  const tags = useTags(true);
  const create = useCreateTag();
  const update = useUpdateTag();
  const move = useMoveTag();
  const remove = useDeleteTag();
  const [editing, setEditing] = useState<TagNode | null>();
  const [form] = Form.useForm();
  const [messageApi, contextHolder] = message.useMessage();
  const flat = useMemo(() => flattenTags(tags.data?.items ?? []), [tags.data]);
  const tree = (tags.data?.items ?? []).map(toTreeNode);

  function open(item?: TagNode) {
    setEditing(item ?? null);
    form.resetFields();
    form.setFieldsValue(item ?? { parent_id: "" });
  }
  async function submit(values: { name: string; category?: string; parent_id?: string; detail?: string }) {
    if (editing) {
      await update.mutateAsync({
        id: editing.id,
        input: {
          name: values.name,
          category: values.category ?? "",
          detail: values.detail ?? ""
        }
      });
      if ((editing.parent_id ?? "") !== (values.parent_id ?? "")) {
        await move.mutateAsync({ id: editing.id, parentId: values.parent_id ?? "" });
      }
    } else {
      await create.mutateAsync(values);
    }
    setEditing(undefined);
    messageApi.success("标签已保存");
  }
  return (
    <div className="admin-page-stack">
      {contextHolder}
      <div className="panel-actions">
        <div><Typography.Title>标签管理</Typography.Title><Typography.Paragraph type="secondary">移动时会阻止形成环；父标签存在子节点时删除会提示先处理子标签。</Typography.Paragraph></div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => open()}>新建标签</Button>
      </div>
      <section className="asset-panel">
        <Tree treeData={tree} titleRender={(node) => {
          const item = flat.find((tag) => tag.id === node.key);
          return <Space><span>{node.title as string}</span>{item ? <><Button size="small" icon={<EditOutlined />} onClick={() => open(item)} /><Button size="small" danger icon={<DeleteOutlined />} onClick={() => void remove.mutateAsync(item.id).catch((error) => messageApi.warning(error.message))} /></> : null}</Space>;
        }} />
      </section>
      <Modal open={editing !== undefined} title={editing ? "编辑标签" : "新建标签"} onCancel={() => setEditing(undefined)} onOk={() => form.submit()}>
        <Form form={form} layout="vertical" onFinish={submit}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="category" label="分类"><Input /></Form.Item>
          <Form.Item name="detail" label="描述"><Input.TextArea /></Form.Item>
          <Form.Item name="parent_id" label="父标签" extra="可选择任意标签；无效移动会提示，但不会隐藏选项。">
            <Select allowClear options={flat.map((item) => ({ value: item.id, label: `${"—".repeat(item.depth)} ${item.name}` }))} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

function toTreeNode(node: TagNode): { key: string; title: string; children?: ReturnType<typeof toTreeNode>[] } {
  return { key: node.id, title: node.name, children: node.children?.map(toTreeNode) };
}
