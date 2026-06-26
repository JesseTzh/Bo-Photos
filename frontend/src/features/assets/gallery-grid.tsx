import { Empty, Skeleton } from "antd";
import type { Asset } from "./schema";
import { AppLink } from "../../shared/adapters/link";

interface GalleryGridProps {
  assets?: Asset[];
  loading?: boolean;
}

export function GalleryGrid({ assets, loading }: GalleryGridProps) {
  if (loading) {
    return (
      <div className="columns-[260px] gap-3">
        {Array.from({ length: 8 }, (_, index) => (
          <Skeleton.Image key={index} active className="mb-3 !h-[260px] !w-full !rounded-lg" />
        ))}
      </div>
    );
  }
  if (!assets?.length) {
    return <Empty description="这里还没有可展示的照片" />;
  }
  return (
    <div className="columns-[260px] gap-3">
      {assets.map((item) => (
        <AppLink
          className="group relative mb-3 block w-full break-inside-avoid overflow-hidden rounded-lg bg-muted"
          key={item.id}
          href={`/preview/${item.id}`}
        >
          <img
            src={item.thumbnail_url || item.preview_url}
            alt={item.title || item.original_name}
            loading="lazy"
            decoding="async"
            className="block w-full object-cover transition-transform duration-500 ease-out group-hover:scale-[1.03]"
            style={{ aspectRatio: item.width > 0 && item.height > 0 ? `${item.width}/${item.height}` : "4/3" }}
          />
          <div className="absolute inset-x-0 bottom-0 translate-y-2 bg-gradient-to-t from-black/80 to-transparent px-4 pb-4 pt-12 opacity-0 transition-all duration-300 group-hover:translate-y-0 group-hover:opacity-100">
            <strong className="block truncate text-sm font-medium text-white">{item.title || item.original_name}</strong>
            <span className="mt-1 block truncate text-xs text-white/70">{[item.camera, item.lens].filter(Boolean).join(" · ")}</span>
          </div>
        </AppLink>
      ))}
    </div>
  );
}
