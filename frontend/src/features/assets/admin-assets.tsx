import { AppstoreOutlined, ClearOutlined, DeleteOutlined, ReloadOutlined, UnorderedListOutlined } from "@ant-design/icons";
import { Button, Empty, Pagination, Segmented, Space, Spin, message } from "antd";
import { useEffect, useMemo, useState } from "react";
import { useAdminSettings } from "../site/api";
import { AdminAssetCard } from "./admin-asset-card";
import { AdminAssetEditDrawer } from "./admin-asset-edit-drawer";
import { AdminAssetFilters, defaultAdminAssetFilters, type AdminAssetFilterState } from "./admin-asset-filters";
import { AdminAssetListRow } from "./admin-asset-list-row";
import { AdminAssetViewDrawer } from "./admin-asset-view-drawer";
import { type AdminAssetQuery, useAdminAssets, useDeleteAssets, usePurgeAsset, useRestoreAsset, useRetryAsset, useUpdateAsset } from "./admin-api";
import type { Asset } from "./schema";

interface AdminAssetsProps {
  refreshToken?: number;
}

export function AdminAssets({ refreshToken = 0 }: AdminAssetsProps) {
  const settings = useAdminSettings();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [filters, setFilters] = useState<AdminAssetFilterState>(defaultAdminAssetFilters);
  const [appliedFilters, setAppliedFilters] = useState<AdminAssetFilterState>(defaultAdminAssetFilters);
  const [viewMode, setViewMode] = useState<"card" | "list">("card");
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [viewing, setViewing] = useState<Asset>();
  const [editing, setEditing] = useState<Asset>();
  const [messageApi, contextHolder] = message.useMessage();
  const remove = useDeleteAssets();
  const purge = usePurgeAsset();
  const restore = useRestoreAsset();
  const retry = useRetryAsset();
  const update = useUpdateAsset();

  useEffect(() => {
    if (settings.data?.admin_images_per_page) setPageSize(settings.data.admin_images_per_page);
  }, [settings.data?.admin_images_per_page]);

  const query: AdminAssetQuery = useMemo(() => ({
    page,
    pageSize,
    title: appliedFilters.title,
    album: appliedFilters.album,
    status: appliedFilters.status,
    visible: appliedFilters.visible,
    private: appliedFilters.private,
    camera: appliedFilters.camera,
    lens: appliedFilters.lens,
    exposure_time: appliedFilters.exposure_time,
    aperture: appliedFilters.aperture,
    iso: appliedFilters.iso,
    tags: appliedFilters.tags,
    tagsOperator: appliedFilters.tagsOperator
  }), [appliedFilters, page, pageSize]);
  const assets = useAdminAssets(query);
  const items = assets.data?.items ?? [];

  useEffect(() => {
    if (refreshToken > 0) void assets.refetch();
  }, [refreshToken]);

  function setItemSelected(id: string, checked: boolean) {
    setSelected((current) => {
      const next = new Set(current);
      if (checked) next.add(id);
      else next.delete(id);
      return next;
    });
  }

  async function deleteItems(ids: string[]) {
    if (!ids.length) return;
    await remove.mutateAsync(ids);
    setSelected(new Set());
    messageApi.success("图片已删除");
  }

  async function purgeItem(id: string) {
    await purge.mutateAsync(id);
    setSelected((current) => {
      const next = new Set(current);
      next.delete(id);
      return next;
    });
    messageApi.success("图片已彻底删除");
  }

  async function toggleAsset(asset: Asset, input: Pick<Asset, "visible"> | Pick<Asset, "private">) {
    await update.mutateAsync({ id: asset.id, input });
  }

  return (
    <section className="asset-panel">
      {contextHolder}
      <AdminAssetFilters
        value={filters}
        onChange={setFilters}
        onApply={() => {
          setAppliedFilters(filters);
          setPage(1);
        }}
        onReset={() => {
          setFilters(defaultAdminAssetFilters);
          setAppliedFilters(defaultAdminAssetFilters);
          setPage(1);
        }}
      />
      <div className="asset-toolbar">
        <Space wrap>
          <Button icon={<ReloadOutlined />} onClick={() => void assets.refetch()}>刷新</Button>
          <Button onClick={() => setSelected(new Set(items.map((item) => item.id)))}>选择当前页</Button>
          <Button icon={<ClearOutlined />} onClick={() => setSelected(new Set())}>清空选择</Button>
          <Button danger icon={<DeleteOutlined />} disabled={!selected.size} loading={remove.isPending} onClick={() => void deleteItems([...selected])}>批量删除</Button>
        </Space>
        <Segmented
          value={viewMode}
          options={[
            { value: "card", icon: <AppstoreOutlined /> },
            { value: "list", icon: <UnorderedListOutlined /> }
          ]}
          onChange={(value) => setViewMode(value as "card" | "list")}
        />
      </div>
      {assets.isPending ? <Spin /> : null}
      {!items.length && !assets.isPending ? <Empty description="暂无图片" /> : null}
      {viewMode === "card" ? (
        <div className="admin-asset-grid">
          {items.map((item) => (
            <AdminAssetCard
              key={item.id}
              asset={item}
              selected={selected.has(item.id)}
              onSelect={(checked) => setItemSelected(item.id, checked)}
              onView={() => setViewing(item)}
              onEdit={() => setEditing(item)}
              onRetry={() => void retry.mutateAsync(item.id)}
              onRestore={() => void restore.mutateAsync(item.id)}
              onDelete={() => void deleteItems([item.id])}
              onPurge={() => void purgeItem(item.id)}
              onToggleVisible={(visible) => void toggleAsset(item, { visible })}
              onTogglePrivate={(privateValue) => void toggleAsset(item, { private: privateValue })}
            />
          ))}
        </div>
      ) : (
        <div className="admin-asset-list">
          {items.map((item) => (
            <AdminAssetListRow
              key={item.id}
              asset={item}
              selected={selected.has(item.id)}
              onSelect={(checked) => setItemSelected(item.id, checked)}
              onView={() => setViewing(item)}
              onEdit={() => setEditing(item)}
              onRetry={() => void retry.mutateAsync(item.id)}
              onRestore={() => void restore.mutateAsync(item.id)}
              onDelete={() => void deleteItems([item.id])}
              onPurge={() => void purgeItem(item.id)}
              onToggleVisible={(visible) => void toggleAsset(item, { visible })}
              onTogglePrivate={(privateValue) => void toggleAsset(item, { private: privateValue })}
            />
          ))}
        </div>
      )}
      {assets.data && assets.data.total > 0 ? (
        <Pagination
          current={page}
          pageSize={pageSize}
          total={assets.data.total}
          showSizeChanger
          onChange={(nextPage, nextPageSize) => {
            setPage(nextPage);
            setPageSize(nextPageSize);
          }}
        />
      ) : null}
      <AdminAssetViewDrawer asset={viewing} open={Boolean(viewing)} onClose={() => setViewing(undefined)} />
      <AdminAssetEditDrawer
        asset={editing}
        open={Boolean(editing)}
        onClose={() => setEditing(undefined)}
        onSaved={() => void assets.refetch()}
      />
    </section>
  );
}
