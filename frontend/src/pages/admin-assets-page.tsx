import { Typography } from "antd";
import { useState } from "react";
import { AdminAssets } from "../features/assets/admin-assets";
import { UploadPanel } from "../features/assets/upload-panel";

export function AdminAssetsPage() {
  const [refreshToken, setRefreshToken] = useState(0);

  return (
    <div className="admin-page-stack">
      <div>
        <Typography.Title>图片管理</Typography.Title>
        <Typography.Paragraph type="secondary">上传、处理、编辑、删除和恢复图库图片。</Typography.Paragraph>
      </div>
      <UploadPanel onAssetReady={() => setRefreshToken((value) => value + 1)} />
      <AdminAssets refreshToken={refreshToken} />
    </div>
  );
}
