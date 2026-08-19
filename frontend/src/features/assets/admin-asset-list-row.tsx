import { DeleteOutlined, EditOutlined, EyeOutlined, ReloadOutlined, UndoOutlined } from "@ant-design/icons";
import { Button, Checkbox, Popconfirm, Space, Switch, Tag, Tooltip } from "antd";
import { isVideoAsset, type Asset } from "./schema";
import { AssetMedia } from "./asset-media";

interface AdminAssetListRowProps {
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
  onTogglePrivate: (value: boolean) => void;
}

export function AdminAssetListRow({
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
  onTogglePrivate
}: AdminAssetListRowProps) {
  const hasPreview = Boolean(asset.thumbnail_url || asset.preview_url || asset.video_url);

  return (
    <div className="admin-asset-row">
      <Checkbox checked={selected} onChange={(event) => onSelect(event.target.checked)} />
      {hasPreview ? (
        <AssetMedia asset={asset} className="admin-asset-row-thumb" muted />
      ) : (
        <div className="admin-asset-row-thumb admin-asset-row-thumb-empty">
          {asset.status === "failed" ? "失败" : "-"}
        </div>
      )}
      <div className="admin-asset-row-main">
        <strong>{asset.title || asset.original_name}</strong>
        <span>{asset.width} x {asset.height}</span>
        <span>{[asset.camera, asset.lens].filter(Boolean).join(" / ") || "Unknown"}</span>
        <span>{asset.shoot_at ? new Date(asset.shoot_at).toLocaleString() : "-"}</span>
      </div>
      <div className="admin-asset-row-actions">
        <Space wrap>
          <Tag className={`asset-status-tag asset-status-tag--${asset.status}`}>{asset.status}</Tag>
          {asset.private ? <Tag className="privacy-status-tag">隐私</Tag> : null}
          <Switch size="small" checked={asset.visible} onChange={onToggleVisible} />
          <Switch size="small" checked={asset.private} onChange={onTogglePrivate} />
          <Tooltip title="查看"><Button icon={<EyeOutlined />} onClick={onView} /></Tooltip>
          <Tooltip title="编辑"><Button icon={<EditOutlined />} onClick={onEdit} /></Tooltip>
          {(asset.status === "failed" || asset.status === "ready") && !isVideoAsset(asset) ? <Tooltip title="重试"><Button icon={<ReloadOutlined />} onClick={onRetry} /></Tooltip> : null}
          {asset.status === "deleted" ? (
            <>
              <Tooltip title="恢复"><Button icon={<UndoOutlined />} onClick={onRestore} /></Tooltip>
              <Popconfirm
                title="彻底删除图片？"
                description="将立即删除原图和衍生图，无法恢复。"
                okText="彻底删除"
                cancelText="取消"
                okButtonProps={{ danger: true }}
                onConfirm={onPurge}
              >
                <Tooltip title="彻底删除"><Button danger icon={<DeleteOutlined />} /></Tooltip>
              </Popconfirm>
            </>
          ) : asset.status !== "purged" ? <Tooltip title="删除"><Button danger icon={<DeleteOutlined />} onClick={onDelete} /></Tooltip> : null}
        </Space>
      </div>
    </div>
  );
}
