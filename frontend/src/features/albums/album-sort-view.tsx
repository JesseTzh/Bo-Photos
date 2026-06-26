import { ArrowDownOutlined, ArrowLeftOutlined, ArrowUpOutlined, SaveOutlined, VerticalAlignBottomOutlined, VerticalAlignTopOutlined } from "@ant-design/icons";
import { Button, Checkbox, Empty, Image, Space, Spin, Typography, message } from "antd";
import { useEffect, useMemo, useState } from "react";
import { type Album, type AlbumAsset, useAlbumAssets, useResetAlbumAssetSort, useSaveAlbumAssetSort } from "./api";

interface AlbumSortViewProps {
  album: Album;
  onBack: () => void;
}

export function AlbumSortView({ album, onBack }: AlbumSortViewProps) {
  const query = useAlbumAssets(album.id);
  const saveSort = useSaveAlbumAssetSort();
  const resetSort = useResetAlbumAssetSort();
  const [items, setItems] = useState<AlbumAsset[]>([]);
  const [originalIDs, setOriginalIDs] = useState<string[]>([]);
  const [selectedIDs, setSelectedIDs] = useState<Set<string>>(new Set());
  const [messageApi, contextHolder] = message.useMessage();
  const currentIDs = useMemo(() => items.map((item) => item.id), [items]);
  const dirty = currentIDs.join("\n") !== originalIDs.join("\n");

  useEffect(() => {
    const next = query.data?.items ?? [];
    setItems(next);
    setOriginalIDs(next.map((item) => item.id));
    setSelectedIDs(new Set());
  }, [query.data]);

  useEffect(() => {
    const listener = (event: BeforeUnloadEvent) => {
      if (!dirty) return;
      event.preventDefault();
    };
    window.addEventListener("beforeunload", listener);
    return () => window.removeEventListener("beforeunload", listener);
  }, [dirty]);

  function leave() {
    if (!dirty || window.confirm("当前排序尚未保存，确定离开？")) onBack();
  }

  function move(id: string, direction: "top" | "up" | "down" | "bottom") {
    setItems((current) => moveItem(current, id, direction));
  }

  function moveSelected(direction: "top" | "bottom") {
    setItems((current) => {
      const selected = current.filter((item) => selectedIDs.has(item.id));
      const rest = current.filter((item) => !selectedIDs.has(item.id));
      return direction === "top" ? [...selected, ...rest] : [...rest, ...selected];
    });
  }

  async function save() {
    await saveSort.mutateAsync({
      albumId: album.id,
      orders: items.map((item, index) => ({ asset_id: item.id, sort: index }))
    });
    setOriginalIDs(items.map((item) => item.id));
    messageApi.success("排序已保存");
  }

  async function reset() {
    if (!window.confirm("重置会按上传时间重新排序，确定继续？")) return;
    await resetSort.mutateAsync(album.id);
    await query.refetch();
    messageApi.success("排序已重置");
  }

  return (
    <div className="admin-page-stack">
      {contextHolder}
      <div className="panel-actions">
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={leave}>返回</Button>
          <Typography.Title level={2}>{album.name}</Typography.Title>
        </Space>
        <Space wrap>
          <Button onClick={() => setSelectedIDs(new Set(items.map((item) => item.id)))}>全选</Button>
          <Button onClick={() => setSelectedIDs(new Set())}>清空</Button>
          <Button icon={<VerticalAlignTopOutlined />} disabled={!selectedIDs.size} onClick={() => moveSelected("top")}>批量置顶</Button>
          <Button icon={<VerticalAlignBottomOutlined />} disabled={!selectedIDs.size} onClick={() => moveSelected("bottom")}>批量置底</Button>
          <Button onClick={() => void reset()} loading={resetSort.isPending}>重置</Button>
          <Button type="primary" icon={<SaveOutlined />} disabled={!dirty} loading={saveSort.isPending} onClick={() => void save()}>保存</Button>
        </Space>
      </div>
      {query.isPending ? <Spin /> : null}
      {!items.length && !query.isPending ? <Empty description="暂无图片" /> : null}
      <div className="album-sort-list">
        {items.map((item) => (
          <div className="album-sort-row" key={item.id}>
            <Checkbox checked={selectedIDs.has(item.id)} onChange={(event) => {
              setSelectedIDs((current) => {
                const next = new Set(current);
                if (event.target.checked) next.add(item.id);
                else next.delete(item.id);
                return next;
              });
            }} />
            <strong>{items.indexOf(item) + 1}</strong>
            {item.thumbnail_url ? <Image className="album-sort-thumb" src={item.thumbnail_url} alt={item.title || item.original_name} /> : <div className="album-sort-thumb" />}
            <div className="album-sort-main">
              <strong>{item.title || item.original_name}</strong>
              <span>{item.width} x {item.height}</span>
            </div>
            <Space className="album-sort-actions">
              <Button icon={<VerticalAlignTopOutlined />} onClick={() => move(item.id, "top")} />
              <Button icon={<ArrowUpOutlined />} onClick={() => move(item.id, "up")} />
              <Button icon={<ArrowDownOutlined />} onClick={() => move(item.id, "down")} />
              <Button icon={<VerticalAlignBottomOutlined />} onClick={() => move(item.id, "bottom")} />
            </Space>
          </div>
        ))}
      </div>
    </div>
  );
}

function moveItem(items: AlbumAsset[], id: string, direction: "top" | "up" | "down" | "bottom") {
  const next = [...items];
  const index = next.findIndex((item) => item.id === id);
  if (index < 0) return next;
  const [item] = next.splice(index, 1);
  if (direction === "top") next.unshift(item);
  if (direction === "up") next.splice(Math.max(0, index - 1), 0, item);
  if (direction === "down") next.splice(Math.min(next.length, index + 1), 0, item);
  if (direction === "bottom") next.push(item);
  return next;
}
