import { DeleteOutlined, EditOutlined, EyeOutlined, ReloadOutlined, UndoOutlined } from "@ant-design/icons";
import { Button, Card, Checkbox, Image, Popconfirm, Space, Switch, Tag, Tooltip } from "antd";
import type { Asset } from "./schema";

interface AdminAssetCardProps {
  asset: Asset;
  selected: boolean;
  onSelect: (checked: boolean) => void;
  onView: () => void;
  onEdit: () => void;
  onRetry: () => void;
  onRestore: () => void;
  onDelete: () => void;
  onPurge: () => void;
  onToggleVisible: (value: boolean) => void;
  onToggleFeatured: (value: boolean) => void;
}

export function AdminAssetCard({
  asset,
  selected,
  onSelect,
  onView,
  onEdit,
  onRetry,
  onRestore,
  onDelete,
  onPurge,
  onToggleVisible,
  onToggleFeatured
}: AdminAssetCardProps) {
  const imageUrl = asset.thumbnail_url || asset.preview_url;

  return (
    <Card
      className="admin-asset-card"
      cover={imageUrl ? (
        <Image src={imageUrl} alt={asset.title || asset.original_name} />
      ) : (
        <div className="admin-asset-card-empty">
          <span>{asset.status === "failed" ? "处理失败" : "暂无预览"}</span>
        </div>
      )}
      actions={[
        <Tooltip title="查看" key="view"><Button type="text" icon={<EyeOutlined />} onClick={onView} /></Tooltip>,
        <Tooltip title="编辑" key="edit"><Button type="text" icon={<EditOutlined />} onClick={onEdit} /></Tooltip>,
        asset.status === "failed" || asset.status === "ready"
          ? <Tooltip title="重试" key="retry"><Button type="text" icon={<ReloadOutlined />} onClick={onRetry} /></Tooltip>
          : <span key="retry" />,
        asset.status === "deleted"
          ? <Space key="deleted-actions" size={0}>
              <Tooltip title="恢复"><Button type="text" icon={<UndoOutlined />} onClick={onRestore} /></Tooltip>
              <Popconfirm
                title="彻底删除图片？"
                description="将立即删除原图和衍生图，无法恢复。"
                okText="彻底删除"
                cancelText="取消"
                okButtonProps={{ danger: true }}
                onConfirm={onPurge}
              >
                <Tooltip title="彻底删除"><Button type="text" danger icon={<DeleteOutlined />} /></Tooltip>
              </Popconfirm>
            </Space>
          : <Tooltip title="删除" key="delete"><Button type="text" danger icon={<DeleteOutlined />} onClick={onDelete} /></Tooltip>
      ]}
    >
      <Card.Meta
        title={<Space><Checkbox checked={selected} onChange={(event) => onSelect(event.target.checked)} />{asset.title || asset.original_name}</Space>}
        description={(
          <Space orientation="vertical" size={8}>
            <Space wrap><Tag>{asset.status}</Tag>{asset.camera ? <span>{asset.camera}</span> : null}</Space>
            <Space wrap>
              <Switch size="small" checked={asset.visible} onChange={onToggleVisible} />
              <span>公开</span>
              <Switch size="small" checked={asset.featured} onChange={onToggleFeatured} />
              <span>精选</span>
            </Space>
          </Space>
        )}
      />
    </Card>
  );
}
