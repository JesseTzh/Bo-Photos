import { ArrowLeftOutlined, DeleteOutlined, SaveOutlined } from "@ant-design/icons";
import { Button, Form, Popconfirm, Result, Skeleton, Space, Typography, message } from "antd";
import { useEffect, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { AlbumForm } from "../features/albums/album-form";
import { type AlbumInput, useAdminAlbum, useDeleteAlbum, useSaveAlbum } from "../features/albums/api";
import { useAllAdminAssets } from "../features/assets/admin-api";
import { isVideoAsset } from "../features/assets/schema";

export function AdminAlbumEditPage() {
  const { id } = useParams();
  const creating = !id || id === "new";
  const album = useAdminAlbum(creating ? undefined : id);
  const assets = useAllAdminAssets({ status: "ready" });
  const saveAlbum = useSaveAlbum();
  const deleteAlbum = useDeleteAlbum();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [messageApi, contextHolder] = message.useMessage();
  const assetOptions = useMemo(
    () => (assets.data ?? []).filter((item) => !isVideoAsset(item)).map((item) => ({ value: item.id, label: item.title || item.original_name })),
    [assets.data]
  );

  useEffect(() => {
    if (creating) {
      form.setFieldsValue({ theme: "0", visible: true, random_show: false, image_sorting: 1, sort: 0 });
    } else if (album.data) {
      form.setFieldsValue(album.data);
    }
  }, [album.data, creating, form]);

  async function submit(values: Record<string, unknown>) {
    const saved = await saveAlbum.mutateAsync({ id: creating ? undefined : id, input: toAlbumInput(values) });
    messageApi.success("基本信息已保存");
    navigate(creating ? `/admin/albums/${saved.id}/images` : "/admin/albums");
  }

  async function remove() {
    if (!id) return;
    await deleteAlbum.mutateAsync(id);
    messageApi.success("相册已删除");
    navigate("/admin/albums");
  }

  if (!creating && album.isPending) return <Skeleton active />;
  if (!creating && !album.data) {
    return <Result status="404" title="相册不存在" extra={<Button onClick={() => navigate("/admin/albums")}>返回相册列表</Button>} />;
  }

  return (
    <div className="admin-page-stack">
      {contextHolder}
      <div className="panel-actions">
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/admin/albums")}>返回</Button>
          <div>
            <Typography.Title level={2}>{creating ? "新建相册" : `编辑 ${album.data?.name}`}</Typography.Title>
            <Typography.Text type="secondary">设置名称、路由、展示方式和封面等基本信息。</Typography.Text>
          </div>
        </Space>
      </div>
      <section className="asset-panel album-basic-form">
        <Form form={form} layout="vertical" onFinish={submit}>
          <AlbumForm assetOptions={assetOptions} />
          <div className="album-form-actions">
            <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saveAlbum.isPending}>保存基本信息</Button>
            {!creating ? (
              <Popconfirm
                title="删除相册？"
                description="只删除相册，图片本身不会被删除。"
                okText="删除"
                cancelText="取消"
                okButtonProps={{ danger: true }}
                onConfirm={remove}
              >
                <Button danger icon={<DeleteOutlined />} loading={deleteAlbum.isPending}>删除相册</Button>
              </Popconfirm>
            ) : null}
          </div>
        </Form>
      </section>
    </div>
  );
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
