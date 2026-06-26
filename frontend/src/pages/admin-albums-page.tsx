import { DeleteOutlined, EditOutlined, PlusOutlined, VerticalAlignTopOutlined } from "@ant-design/icons";
import { Button, Form, Image, Modal, Space, Switch, Table, Tag, Typography, message } from "antd";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAdminAssets } from "../features/assets/admin-api";
import { AlbumForm } from "../features/albums/album-form";
import { type Album, type AlbumInput, useAlbums, useDeleteAlbum, useReplaceAlbumAssets, useSaveAlbum, useSaveAlbumOrder, useSetAlbumCover } from "../features/albums/api";

export function AdminAlbumsPage() {
  const navigate = useNavigate();
  const albums = useAlbums(true);
  const assets = useAdminAssets({ page: 1, pageSize: 500, status: "ready" });
  const saveAlbum = useSaveAlbum();
  const remove = useDeleteAlbum();
  const replaceAssets = useReplaceAlbumAssets();
  const saveOrder = useSaveAlbumOrder();
  const setCover = useSetAlbumCover();
  const [editing, setEditing] = useState<Album | null>();
  const [form] = Form.useForm();
  const [messageApi, contextHolder] = message.useMessage();
  const albumItems = albums.data?.items ?? [];
  const albumIDs = albumItems.map((item) => item.id);
  const assetOptions = useMemo(
    () => (assets.data?.items ?? []).map((item) => ({ value: item.id, label: item.title || item.original_name })),
    [assets.data]
  );

  function open(item?: Album) {
    setEditing(item ?? null);
    form.resetFields();
    form.setFieldsValue(item ?? {
      theme: "0",
      visible: true,
      random_show: false,
      image_sorting: 1,
      sort: 0,
      asset_ids: []
    });
  }

  async function submit(values: Record<string, unknown>) {
    const assetIds = (values.asset_ids as string[] | undefined) ?? [];
    const saved = await saveAlbum.mutateAsync({
      id: editing?.id,
      input: toAlbumInput(values)
    });
    await replaceAssets.mutateAsync({ id: saved.id, assetIds });
    setEditing(undefined);
    messageApi.success("相册已保存");
  }

  async function quickSave(item: Album, patch: Partial<AlbumInput>) {
    await saveAlbum.mutateAsync({ id: item.id, input: toAlbumInput({ ...item, ...patch }) });
    messageApi.success("相册已更新");
  }

  async function move(id: string, direction: "top" | "up" | "down") {
    await saveOrder.mutateAsync(moveAlbum(albumIDs, id, direction));
  }

  async function clearCover(id: string) {
    await setCover.mutateAsync({ albumId: id, assetId: "" });
    messageApi.success("封面已清除");
  }

  return (
    <div className="admin-page-stack">
      {contextHolder}
      <div className="panel-actions">
        <div>
          <Typography.Title>相册管理</Typography.Title>
          <Typography.Paragraph type="secondary">管理封面、公开状态、图片关联与顺序。</Typography.Paragraph>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => open()}>新建相册</Button>
      </div>
      <Table
        rowKey="id"
        loading={albums.isPending}
        dataSource={albumItems}
        scroll={{ x: 980 }}
        columns={[
          {
            title: "封面",
            render: (_, item) => item.cover_url ? <Image className="admin-album-cover" src={item.cover_url} alt={item.name} /> : <div className="admin-album-cover" />
          },
          { title: "名称", dataIndex: "name" },
          { title: "路由", dataIndex: "album_value" },
          { title: "主题", render: (_, item) => item.theme === "1" ? "简洁" : "默认" },
          { title: "图片", dataIndex: "asset_count" },
          { title: "公开", render: (_, item) => <Switch checked={item.visible} onChange={(visible) => void quickSave(item, { visible })} /> },
          { title: "随机", render: (_, item) => <Switch checked={item.random_show} onChange={(random_show) => void quickSave(item, { random_show })} /> },
          { title: "排序", dataIndex: "sort" },
          { title: "图片规则", render: (_, item) => <Tag>{sortingLabel(item.image_sorting)}</Tag> },
          {
            title: "操作",
            render: (_, item) => (
              <Space wrap>
                <Button icon={<VerticalAlignTopOutlined />} onClick={() => void move(item.id, "top")} />
                <Button onClick={() => void move(item.id, "up")}>上移</Button>
                <Button onClick={() => void move(item.id, "down")}>下移</Button>
                <Button icon={<EditOutlined />} onClick={() => open(item)} />
                <Button onClick={() => navigate(`/admin/albums/${item.id}/sort`)}>图片排序</Button>
                <Button onClick={() => void clearCover(item.id)} disabled={!item.cover_asset_id}>清封面</Button>
                <Button danger icon={<DeleteOutlined />} onClick={() => void remove.mutateAsync(item.id)} />
              </Space>
            )
          }
        ]}
      />
      <Modal
        open={editing !== undefined}
        title={editing ? "编辑相册" : "新建相册"}
        onCancel={() => setEditing(undefined)}
        onOk={() => form.submit()}
        confirmLoading={saveAlbum.isPending || replaceAssets.isPending}
        width={760}
      >
        <Form form={form} layout="vertical" onFinish={submit}>
          <AlbumForm assetOptions={assetOptions} />
        </Form>
      </Modal>
    </div>
  );
}

function moveAlbum(ids: string[], id: string, direction: "top" | "up" | "down") {
  const next = [...ids];
  const index = next.indexOf(id);
  if (index < 0) return next;
  const [item] = next.splice(index, 1);
  if (direction === "top") next.unshift(item);
  if (direction === "up") next.splice(Math.max(0, index - 1), 0, item);
  if (direction === "down") next.splice(Math.min(next.length, index + 1), 0, item);
  return next;
}

function toAlbumInput(values: Record<string, unknown>): AlbumInput {
  return {
    name: String(values.name ?? ""),
    album_value: String(values.album_value ?? ""),
    detail: String(values.detail ?? ""),
    theme: String(values.theme ?? "0"),
    visible: Boolean(values.visible ?? true),
    sort: Number(values.sort ?? 0),
    random_show: Boolean(values.random_show ?? false),
    license: String(values.license ?? ""),
    cover_asset_id: typeof values.cover_asset_id === "string" ? values.cover_asset_id : "",
    image_sorting: Number(values.image_sorting ?? 1)
  };
}

function sortingLabel(value: number) {
  if (value === 2) return "拍摄时间新到旧";
  if (value === 3) return "上传时间旧到新";
  if (value === 4) return "拍摄时间旧到新";
  return "上传时间新到旧";
}
