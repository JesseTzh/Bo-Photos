import { EyeOutlined, SearchOutlined, VideoCameraOutlined } from "@ant-design/icons";
import { Button, Empty, Input, Modal, Pagination, Tag, Tooltip } from "antd";
import { useEffect, useMemo, useState } from "react";
import { AssetMedia } from "../assets/asset-media";
import { isVideoAsset, type Asset } from "../assets/schema";

interface HomeAssetPickerProps {
  open: boolean;
  assets: Asset[];
  currentId?: string;
  excludedIds?: string[];
  onCancel: () => void;
  onSelect: (asset: Asset) => void;
}

export function HomeAssetPicker({ open, assets, currentId, excludedIds = [], onCancel, onSelect }: HomeAssetPickerProps) {
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [previewing, setPreviewing] = useState<Asset>();
  const pageSize = 12;

  useEffect(() => {
    if (open) {
      setSearch("");
      setPage(1);
    }
  }, [open]);

  const filtered = useMemo(() => {
    const query = search.trim().toLowerCase();
    const excluded = new Set(excludedIds.filter((id) => id !== currentId));
    return assets.filter((asset) => {
      if (excluded.has(asset.id)) return false;
      if (!query) return true;
      return `${asset.title ?? ""} ${asset.original_name}`.toLowerCase().includes(query);
    });
  }, [assets, currentId, excludedIds, search]);
  const pageItems = filtered.slice((page - 1) * pageSize, page * pageSize);

  return (
    <>
      <Modal open={open} title="选择首页素材" footer={null} width={1040} onCancel={onCancel} destroyOnHidden>
        <Input
          allowClear
          prefix={<SearchOutlined />}
          placeholder="搜索素材标题或文件名"
          value={search}
          onChange={(event) => {
            setSearch(event.target.value);
            setPage(1);
          }}
        />
        {pageItems.length ? (
          <div className="home-asset-picker-grid">
            {pageItems.map((asset) => (
              <div key={asset.id} className={`home-asset-picker-item${asset.id === currentId ? " is-selected" : ""}`}>
                <div className="home-asset-picker-media">
                  <AssetMedia asset={asset} muted />
                  {isVideoAsset(asset) ? <Tag icon={<VideoCameraOutlined />}>视频</Tag> : null}
                </div>
                <div className="home-asset-picker-info">
                  <span title={asset.title || asset.original_name}>{asset.title || asset.original_name}</span>
                  <div>
                    <Tooltip title="查看素材">
                      <Button type="text" icon={<EyeOutlined />} onClick={() => setPreviewing(asset)} />
                    </Tooltip>
                    <Button type={asset.id === currentId ? "default" : "primary"} onClick={() => onSelect(asset)}>
                      {asset.id === currentId ? "已选择" : "选择"}
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : <Empty className="home-asset-picker-empty" description="没有可选素材" />}
        {filtered.length > pageSize ? (
          <Pagination current={page} pageSize={pageSize} total={filtered.length} hideOnSinglePage onChange={setPage} />
        ) : null}
      </Modal>
      <HomeAssetPreview asset={previewing} onClose={() => setPreviewing(undefined)} />
    </>
  );
}

export function HomeAssetPreview({ asset, onClose }: { asset?: Asset; onClose: () => void }) {
  return (
    <Modal open={Boolean(asset)} title={asset?.title || asset?.original_name || "查看素材"} footer={null} width={960} onCancel={onClose}>
      {asset ? <AssetMedia asset={asset} controls className="home-asset-preview" /> : null}
    </Modal>
  );
}
