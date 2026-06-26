import { memo, useEffect, useMemo, useRef, useState } from "react";
import type { Asset } from "../assets/schema";
import { AppLink } from "../../shared/adapters/link";
import { cn } from "../../shared/lib/utils";

interface LayoutItem {
  asset: Asset;
  index: number;
  x: number;
  y: number;
  w: number;
  h: number;
}

interface VirtualWaterfallGalleryProps {
  assets: Asset[];
  previewSearch?: string;
  overscanPx?: number;
}

export function VirtualWaterfallGallery({ assets, previewSearch = "", overscanPx = 800 }: VirtualWaterfallGalleryProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [containerWidth, setContainerWidth] = useState(0);
  const [viewport, setViewport] = useState(() => ({
    scrollY: typeof window === "undefined" ? 0 : window.scrollY,
    height: typeof window === "undefined" ? 800 : window.innerHeight
  }));
  const rafRef = useRef(0);

  useEffect(() => {
    const element = containerRef.current;
    if (!element) return;
    const resizeObserver = new ResizeObserver((entries) => {
      const width = Math.floor(entries[0]?.contentRect.width ?? 0);
      if (width > 0) setContainerWidth(width);
    });
    resizeObserver.observe(element);
    return () => resizeObserver.disconnect();
  }, []);

  useEffect(() => {
    const update = () => {
      window.cancelAnimationFrame(rafRef.current);
      rafRef.current = window.requestAnimationFrame(() => {
        setViewport({ scrollY: window.scrollY, height: window.innerHeight });
      });
    };
    update();
    window.addEventListener("scroll", update, { passive: true });
    window.addEventListener("resize", update, { passive: true });
    return () => {
      window.removeEventListener("scroll", update);
      window.removeEventListener("resize", update);
      window.cancelAnimationFrame(rafRef.current);
    };
  }, []);

  const layout = useMemo(() => {
    if (!containerWidth || !assets.length) return { items: [] as LayoutItem[], totalHeight: 220 };

    const cols = containerWidth < 640 ? 2 : containerWidth < 1024 ? 3 : 4;
    const gap = containerWidth < 640 ? 6 : 10;
    const colWidth = Math.floor((containerWidth - gap * (cols - 1)) / cols);
    const colHeights = new Array<number>(cols).fill(0);

    const items = assets.map((asset, index) => {
      const ratio = asset.width > 0 && asset.height > 0 ? asset.width / asset.height : 3 / 4;
      const h = Math.min(Math.round(colWidth / ratio), Math.round(colWidth * 2.5));
      let col = 0;
      for (let nextCol = 1; nextCol < cols; nextCol += 1) {
        if (colHeights[nextCol] < colHeights[col]) col = nextCol;
      }
      const item = {
        asset,
        index,
        x: col * (colWidth + gap),
        y: colHeights[col],
        w: colWidth,
        h
      };
      colHeights[col] += h + gap;
      return item;
    });

    return { items, totalHeight: Math.max(220, ...colHeights) };
  }, [assets, containerWidth]);

  const visibleItems = useMemo(() => {
    if (!layout.items.length) return layout.items;
    const containerTop = containerRef.current ? containerRef.current.getBoundingClientRect().top + window.scrollY : 0;
    const top = viewport.scrollY - containerTop - overscanPx;
    const bottom = viewport.scrollY - containerTop + viewport.height + overscanPx;
    return layout.items.filter((item) => item.y + item.h > top && item.y < bottom);
  }, [layout.items, overscanPx, viewport]);

  return (
    <div ref={containerRef} className="relative w-full" style={{ height: layout.totalHeight }}>
      {visibleItems.map((item) => (
        <WaterfallCard key={item.asset.id} item={item} previewSearch={previewSearch} />
      ))}
    </div>
  );
}

const WaterfallCard = memo(function WaterfallCard({ item, previewSearch }: { item: LayoutItem; previewSearch: string }) {
  const [loaded, setLoaded] = useState(false);
  const imageUrl = item.asset.preview_url || item.asset.thumbnail_url || "";
  const href = `/preview/${item.asset.id}${previewSearch}`;

  return (
    <AppLink
      href={href}
      className="group absolute cursor-pointer overflow-hidden rounded-lg bg-muted"
      style={{ left: item.x, top: item.y, width: item.w, height: item.h }}
    >
      {!loaded ? <div className="absolute inset-0 animate-pulse bg-muted" /> : null}
      {imageUrl ? (
        <img
          src={imageUrl}
          alt={item.asset.title || item.asset.original_name}
          width={item.w}
          height={item.h}
          loading={item.index < 8 ? "eager" : "lazy"}
          decoding="async"
          onLoad={() => setLoaded(true)}
          className={cn("h-full w-full object-cover transition-opacity duration-200", loaded ? "opacity-100" : "opacity-0")}
        />
      ) : null}
    </AppLink>
  );
});
