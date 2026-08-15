import { ArrowLeftOutlined, DeleteOutlined, PlusOutlined, SortAscendingOutlined } from "@ant-design/icons";
import { Button, Empty, Image, Modal, Popconfirm, Result, Skeleton, Space, Table, Typography, message } from "antd";
import { useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { type AlbumAsset, useAdminAlbum, useAlbumAssets, useReplaceAlbumAssets } from "../features/albums/api";
import { useAllAdminAssets } from "../features/assets/admin-api";
import type { Asset } from "../features/assets/schema";

export function AdminAlbumImagesPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const album = useAdminAlbum(id);
  const albumAssets = useAlbumAssets(id);
  const allAssets = useAllAdminAssets({ status: "ready" });
  const replaceAssets = useReplaceAlbumAssets();
  const [pickerOpen, setPickerOpen] = useState(false);
  const [selectedToAdd, setSelectedToAdd] = useState<React.Key[]>([]);
  const [selectedToRemove, setSelectedToRemove] = useState<React.Key[]>([]);
  const [messageApi, contextHolder] = message.useMessage();
  const currentItems = albumAssets.data?.items ?? [];
  const currentIDs = useMemo(() => new Set(currentItems.map((item) => item.id)), [currentItems]);
  const availableItems = useMemo(
    () => (allAssets.data ?? []).filter((item) => !currentIDs.has(item.id)),
    [allAssets.data, currentIDs]
  );

  async function replace(nextIDs: string[], successMessage: string) {
    if (!id) return;
    await replaceAssets.mutateAsync({ id, assetIds: nextIDs });
    setSelectedToRemove([]);
    messageApi.success(successMessage);
  }

  async function addSelected() {
    await replace([...currentItems.map((item) => item.id), ...selectedToAdd.map(String)], "图片已加入相册");
    setSelectedToAdd([]);
    setPickerOpen(false);
  }

  async function removeSelected(ids: string[]) {
    const removed = new Set(ids);
    await replace(currentItems.filter((item) => !removed.has(item.id)).map((item) => item.id), "图片已从相册移除");
  }

  if (album.isPending) return <Skeleton active />;
  if (!album.data) {
    return <Result status="404" title="相册不存在" extra={<Button onClick={() => navigate("/admin/albums")}>返回相册列表</Button>} />;
  }

  return (
    <div className="admin-page-stack">
      {contextHolder}
      <div className="panel-actions">
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/admin/albums")}>返回</Button>
          <div>
            <Typography.Title level={2}>{album.data.name} · 图片管理</Typography.Title>
            <Typography.Text type="secondary">当前共 {currentItems.length} 张图片</Typography.Text>
          </div>
        </Space>
        <Space wrap>
          <Button icon={<SortAscendingOutlined />} onClick={() => navigate(`/admin/albums/${id}/sort`)}>图片排序</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setPickerOpen(true)}>新增图片</Button>
        </Space>
      </div>
      <section className="asset-panel">
        <div className="album-image-toolbar">
          <Typography.Text strong>相册图片</Typography.Text>
          <Popconfirm
            title={`移除选中的 ${selectedToRemove.length} 张图片？`}
            description="只会解除相册关联，不会删除图片。"
            okText="移除"
            cancelText="取消"
            onConfirm={() => removeSelected(selectedToRemove.map(String))}
          >
            <Button danger icon={<DeleteOutlined />} disabled={!selectedToRemove.length} loading={replaceAssets.isPending}>批量移除</Button>
          </Popconfirm>
        </div>
        <Table<AlbumAsset>
          rowKey="id"
          loading={albumAssets.isPending}
          dataSource={currentItems}
          locale={{ emptyText: <Empty description="这个相册还没有图片" /> }}
          rowSelection={{ selectedRowKeys: selectedToRemove, onChange: setSelectedToRemove }}
          columns={albumImageColumns((assetId) => removeSelected([assetId]), replaceAssets.isPending)}
        />
      </section>
      <Modal
        open={pickerOpen}
        title="新增图片"
        width={900}
        okText={`加入相册${selectedToAdd.length ? `（${selectedToAdd.length}）` : ""}`}
        cancelText="取消"
        okButtonProps={{ disabled: !selectedToAdd.length, loading: replaceAssets.isPending }}
        onOk={() => addSelected()}
        onCancel={() => { setPickerOpen(false); setSelectedToAdd([]); }}
      >
        <Typography.Paragraph type="secondary">这里只显示尚未加入当前相册的图片。</Typography.Paragraph>
        <Table<Asset>
          rowKey="id"
          size="small"
          loading={allAssets.isPending}
          dataSource={availableItems}
          pagination={{ pageSize: 8, showSizeChanger: false }}
          locale={{ emptyText: <Empty description="没有可新增的图片" /> }}
          rowSelection={{ selectedRowKeys: selectedToAdd, onChange: setSelectedToAdd }}
          columns={pickerColumns}
        />
      </Modal>
    </div>
  );
}

function albumImageColumns(onRemove: (id: string) => void, loading: boolean) {
  return [
    {
      title: "图片",
      width: 120,
      render: (_: unknown, item: AlbumAsset) => item.thumbnail_url || item.preview_url ? (
        <Image className="album-management-thumb" width={96} height={64} src={item.thumbnail_url || item.preview_url} alt={item.title || item.original_name} />
      ) : <div className="album-management-thumb" />
    },
    { title: "名称", render: (_: unknown, item: AlbumAsset) => item.title || item.original_name },
    { title: "尺寸", width: 140, render: (_: unknown, item: AlbumAsset) => `${item.width} x ${item.height}` },
    {
      title: "操作",
      width: 100,
      render: (_: unknown, item: AlbumAsset) => (
        <Popconfirm title="从相册移除这张图片？" description="图片本身不会被删除。" okText="移除" cancelText="取消" onConfirm={() => onRemove(item.id)}>
          <Button danger type="text" icon={<DeleteOutlined />} loading={loading}>移除</Button>
        </Popconfirm>
      )
    }
  ];
}

const pickerColumns = [
  {
    title: "图片",
    width: 100,
    render: (_: unknown, item: Asset) => item.thumbnail_url || item.preview_url ? (
      <Image className="album-picker-thumb" width={72} height={48} preview={false} src={item.thumbnail_url || item.preview_url} alt={item.title || item.original_name} />
    ) : <div className="album-picker-thumb" />
  },
  { title: "名称", render: (_: unknown, item: Asset) => item.title || item.original_name },
  { title: "尺寸", width: 130, render: (_: unknown, item: Asset) => `${item.width} x ${item.height}` }
];
