import { DownloadOutlined } from "@ant-design/icons";
import { Button, Descriptions, Drawer, Space, Tag } from "antd";
import { useAlbums } from "../albums/api";
import { flattenTags, useAssetTags, useTags } from "../tags/api";
import { useAssetAlbums } from "./admin-api";
import type { Asset } from "./schema";
import { AssetMedia } from "./asset-media";

interface AdminAssetViewDrawerProps {
  asset?: Asset;
  open: boolean;
  onClose: () => void;
}

export function AdminAssetViewDrawer({ asset, open, onClose }: AdminAssetViewDrawerProps) {
  const assetTags = useAssetTags(asset?.id);
  const assetAlbums = useAssetAlbums(asset?.id);
  const tags = useTags(true);
  const albums = useAlbums(true);
  const tagNames = flattenTags(tags.data?.items ?? [])
    .filter((tag) => assetTags.data?.tag_ids.includes(tag.id))
    .map((tag) => tag.name);
  const albumNames = (albums.data?.items ?? [])
    .filter((album) => assetAlbums.data?.album_ids.includes(album.id))
    .map((album) => album.name);

  return (
    <Drawer open={open} size={620} title="素材详情" onClose={onClose}>
      {asset ? (
        <Space orientation="vertical" size="large" className="drawer-stack">
          {asset.preview_url || asset.video_url ? <AssetMedia asset={asset} controls className="w-full" /> : null}
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="文件名">{asset.original_name}</Descriptions.Item>
            <Descriptions.Item label="类型">{asset.mime_type || "-"}</Descriptions.Item>
            <Descriptions.Item label="状态"><Tag className={`asset-status-tag asset-status-tag--${asset.status}`}>{asset.status}</Tag></Descriptions.Item>
            <Descriptions.Item label="尺寸">{asset.width} x {asset.height}</Descriptions.Item>
            <Descriptions.Item label="大小">{asset.byte_size ? `${(asset.byte_size / 1048576).toFixed(2)} MB` : "-"}</Descriptions.Item>
            <Descriptions.Item label="相机">{asset.camera || "-"}</Descriptions.Item>
            <Descriptions.Item label="镜头">{asset.lens || "-"}</Descriptions.Item>
            <Descriptions.Item label="曝光">{[asset.exposure_time, asset.aperture, asset.iso ? `ISO ${asset.iso}` : ""].filter(Boolean).join(" / ") || "-"}</Descriptions.Item>
            <Descriptions.Item label="焦距">{asset.focal_length || "-"}</Descriptions.Item>
            <Descriptions.Item label="拍摄时间">{asset.shoot_at ? new Date(asset.shoot_at).toLocaleString() : "-"}</Descriptions.Item>
            <Descriptions.Item label="坐标">{asset.latitude !== undefined && asset.longitude !== undefined ? `${asset.latitude}, ${asset.longitude}` : "-"}</Descriptions.Item>
            <Descriptions.Item label="描述">{asset.description || "-"}</Descriptions.Item>
            <Descriptions.Item label="标签">{tagNames.length ? tagNames.join(" / ") : "-"}</Descriptions.Item>
            <Descriptions.Item label="相册">{albumNames.length ? albumNames.join(" / ") : "-"}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{asset.created_at ? new Date(asset.created_at).toLocaleString() : "-"}</Descriptions.Item>
            <Descriptions.Item label="更新时间">{asset.updated_at ? new Date(asset.updated_at).toLocaleString() : "-"}</Descriptions.Item>
          </Descriptions>
          {asset.original_url ? <Button icon={<DownloadOutlined />} href={asset.original_url} target="_blank">打开原图</Button> : null}
        </Space>
      ) : null}
    </Drawer>
  );
}
