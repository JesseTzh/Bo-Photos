import { FilterOutlined, ReloadOutlined } from "@ant-design/icons";
import { Button, Input, Segmented, Select } from "antd";
import { useMemo } from "react";
import { useAlbums } from "../albums/api";
import { flattenTags, useTags } from "../tags/api";
import { useAdminAssetFilterOptions } from "./admin-api";
import type { AssetStatus } from "./schema";

export interface AdminAssetFilterState {
  title: string;
  album?: string;
  status?: AssetStatus;
  visible?: boolean;
  private?: boolean;
  featured?: boolean;
  camera?: string;
  lens?: string;
  exposure_time?: string;
  aperture?: string;
  iso?: string;
  tags: string[];
  tagsOperator: "and" | "or";
}

export const defaultAdminAssetFilters: AdminAssetFilterState = {
  title: "",
  tags: [],
  tagsOperator: "and"
};

interface AdminAssetFiltersProps {
  value: AdminAssetFilterState;
  onChange: (value: AdminAssetFilterState) => void;
  onApply: () => void;
  onReset: () => void;
}

const statuses: AssetStatus[] = ["processing", "ready", "failed", "deleted", "purged"];

export function AdminAssetFilters({ value, onChange, onApply, onReset }: AdminAssetFiltersProps) {
  const options = useAdminAssetFilterOptions();
  const albums = useAlbums(true);
  const tags = useTags(true);
  const tagOptions = useMemo(
    () => flattenTags(tags.data?.items ?? []).map((item) => ({ value: item.id, label: `${"—".repeat(item.depth)} ${item.name}` })),
    [tags.data]
  );

  const patch = (patchValue: Partial<AdminAssetFilterState>) => onChange({ ...value, ...patchValue });

  return (
    <div className="admin-asset-filter-bar">
      <Input.Search placeholder="标题" allowClear value={value.title} onChange={(event) => patch({ title: event.target.value })} onSearch={onApply} />
      <Select allowClear placeholder="相册" value={value.album} options={(albums.data?.items ?? []).map((item) => ({ value: item.album_value, label: item.name }))} onChange={(album) => patch({ album })} />
      <Select allowClear placeholder="状态" value={value.status} options={statuses.map((status) => ({ value: status, label: status }))} onChange={(status) => patch({ status })} />
      <Select allowClear placeholder="公开" value={value.visible} options={[{ value: true, label: "公开" }, { value: false, label: "隐藏" }]} onChange={(visible) => patch({ visible })} />
      <Select allowClear placeholder="隐私" value={value.private} options={[{ value: true, label: "隐私" }, { value: false, label: "非隐私" }]} onChange={(privateValue) => patch({ private: privateValue })} />
      <Select allowClear placeholder="精选" value={value.featured} options={[{ value: true, label: "精选" }, { value: false, label: "非精选" }]} onChange={(featured) => patch({ featured })} />
      <Select allowClear showSearch placeholder="相机" value={value.camera} options={(options.data?.cameras ?? []).map((item) => ({ value: item, label: item }))} onChange={(camera) => patch({ camera })} />
      <Select allowClear showSearch placeholder="镜头" value={value.lens} options={(options.data?.lenses ?? []).map((item) => ({ value: item, label: item }))} onChange={(lens) => patch({ lens })} />
      <Select allowClear showSearch placeholder="快门" value={value.exposure_time} options={(options.data?.exposure_times ?? []).map((item) => ({ value: item, label: item }))} onChange={(exposure_time) => patch({ exposure_time })} />
      <Select allowClear showSearch placeholder="光圈" value={value.aperture} options={(options.data?.apertures ?? []).map((item) => ({ value: item, label: item }))} onChange={(aperture) => patch({ aperture })} />
      <Select allowClear showSearch placeholder="ISO" value={value.iso} options={(options.data?.isos ?? []).map((item) => ({ value: item, label: item }))} onChange={(iso) => patch({ iso })} />
      <Select mode="multiple" allowClear placeholder="标签" value={value.tags} options={tagOptions} onChange={(tagsValue) => patch({ tags: tagsValue })} />
      <Segmented value={value.tagsOperator} options={[{ value: "and", label: "AND" }, { value: "or", label: "OR" }]} onChange={(tagsOperator) => patch({ tagsOperator: tagsOperator as "and" | "or" })} />
      <Button type="primary" icon={<FilterOutlined />} onClick={onApply}>应用</Button>
      <Button icon={<ReloadOutlined />} onClick={onReset}>重置</Button>
    </div>
  );
}
