import { Check, LayoutGrid, Rows, Settings2, SlidersHorizontal, X, ArrowUpDown } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { Empty, Pagination, Spin } from "antd";
import type { Asset, AssetPage } from "../assets/schema";
import { AppLink } from "../../shared/adapters/link";
import { cn } from "../../shared/lib/utils";
import { VirtualWaterfallGallery } from "./virtual-waterfall-gallery";

interface FabItem {
  icon: React.ReactNode;
  label: string;
  active: boolean;
  badge?: number;
  onClick: () => void;
}

interface Option {
  value: string;
  label: string;
}

interface ThemeGalleryProps {
  page: number;
  pageSize: number;
  data?: AssetPage;
  loading?: boolean;
  error?: unknown;
  cameras: string[];
  lenses: string[];
  tags: Option[];
  selectedCameras: string[];
  selectedLenses: string[];
  selectedTags: string[];
  tagsOperator: "and" | "or";
  sort?: "asc" | "desc";
  onPageChange: (page: number) => void;
  onCamerasChange: (values: string[]) => void;
  onLensesChange: (values: string[]) => void;
  onTagsChange: (values: string[]) => void;
  onTagsOperatorChange: (value: "and" | "or") => void;
  onSortChange: (value?: "asc" | "desc") => void;
  onReset: () => void;
  album?: string;
  preferredStyle?: "waterfall" | "single";
  systemStyle?: "waterfall" | "single";
  previewSearch?: string;
}

const sortOptions: Array<{ value?: "asc" | "desc"; label: string }> = [
  { value: undefined, label: "默认" },
  { value: "desc", label: "最新" },
  { value: "asc", label: "最早" }
];

function toggleValue(values: string[], value: string) {
  return values.includes(value) ? values.filter((item) => item !== value) : [...values, value];
}

function ChipGroup({
  label,
  options,
  selected,
  onChange
}: {
  label: string;
  options: Option[];
  selected: string[];
  onChange: (values: string[]) => void;
}) {
  const [search, setSearch] = useState("");
  const filtered = useMemo(
    () => (search.trim() ? options.filter((item) => item.label.toLowerCase().includes(search.toLowerCase())) : options),
    [options, search]
  );

  if (!options.length) return null;

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <p className="text-xs font-medium text-muted-foreground">{label}</p>
        {selected.length ? (
          <button type="button" onClick={() => onChange([])} className="text-[11px] text-muted-foreground hover:text-foreground">
            清除
          </button>
        ) : null}
      </div>

      {options.length > 8 ? (
        <div className="relative mb-2">
          <input
            type="text"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder={`搜索${label}...`}
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary/50"
          />
          {search ? (
            <button type="button" onClick={() => setSearch("")} className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground">
              <X className="h-3.5 w-3.5" />
            </button>
          ) : null}
        </div>
      ) : null}

      <div className="flex flex-wrap gap-2">
        {filtered.map((option) => {
          const active = selected.includes(option.value);
          return (
            <button
              key={option.value}
              type="button"
              onClick={() => onChange(toggleValue(selected, option.value))}
              className={cn(
                "inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-medium transition-colors",
                active
                  ? "border-primary bg-primary/10 text-primary"
                  : "border-border bg-background text-muted-foreground hover:bg-accent hover:text-foreground"
              )}
            >
              {active ? <Check className="h-3 w-3 shrink-0" /> : null}
              {option.label}
            </button>
          );
        })}
        {!filtered.length ? <p className="text-xs text-muted-foreground">无匹配结果</p> : null}
      </div>
    </div>
  );
}

function FilterPanel(props: ThemeGalleryProps) {
  const hasAny = props.selectedCameras.length || props.selectedLenses.length || props.selectedTags.length || props.sort;
  return (
    <div className="space-y-6">
      <div>
        <p className="mb-2 text-xs font-medium text-muted-foreground">拍摄时间排序</p>
        <div className="flex gap-1.5">
          {sortOptions.map((option) => (
            <button
              key={String(option.value)}
              type="button"
              onClick={() => props.onSortChange(option.value)}
              className={cn(
                "flex-1 rounded-lg border py-2 text-sm font-medium transition-colors",
                option.value === props.sort
                  ? "border-primary bg-primary text-primary-foreground"
                  : "border-border bg-background text-muted-foreground hover:bg-accent hover:text-foreground"
              )}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      <ChipGroup
        label="相机"
        options={props.cameras.map((value) => ({ value, label: value }))}
        selected={props.selectedCameras}
        onChange={props.onCamerasChange}
      />
      <ChipGroup
        label="镜头"
        options={props.lenses.map((value) => ({ value, label: value }))}
        selected={props.selectedLenses}
        onChange={props.onLensesChange}
      />
      <ChipGroup label="标签" options={props.tags} selected={props.selectedTags} onChange={props.onTagsChange} />

      {props.selectedTags.length > 1 ? (
        <div className="flex items-center gap-2">
          <span className="text-xs text-muted-foreground">标签匹配：</span>
          {(["and", "or"] as const).map((operator) => (
            <button
              key={operator}
              type="button"
              onClick={() => props.onTagsOperatorChange(operator)}
              className={cn(
                "rounded-md border px-3 py-1 text-xs font-medium transition-colors",
                props.tagsOperator === operator
                  ? "border-primary bg-primary text-primary-foreground"
                  : "border-border bg-background text-muted-foreground hover:bg-accent"
              )}
            >
              {operator === "and" ? "全部匹配" : "任一匹配"}
            </button>
          ))}
        </div>
      ) : null}

      {hasAny ? (
        <button
          type="button"
          onClick={props.onReset}
          className="w-full rounded-xl border border-destructive/30 py-2.5 text-sm text-destructive transition-colors hover:bg-destructive/10"
        >
          清除全部筛选
        </button>
      ) : null}
    </div>
  );
}

function SingleGallery({ assets, previewSearch = "" }: { assets: Asset[]; previewSearch?: string }) {
  if (!assets.length) return <Empty description="暂无匹配的图片" />;
  return (
    <div className="mx-auto w-full max-w-[900px] space-y-4 px-3 pb-16 sm:px-4 md:px-6">
      {assets.map((asset) => (
        <AppLink key={asset.id} href={`/preview/${asset.id}${previewSearch}`} className="group block overflow-hidden rounded-xl bg-muted">
          <img
            src={asset.preview_url || asset.thumbnail_url}
            alt={asset.title || asset.original_name}
            loading="lazy"
            decoding="async"
            className="h-auto w-full object-cover transition-opacity duration-300 group-hover:opacity-95"
          />
          <div className="border-x border-b border-border bg-card px-4 py-3">
            <h3 className="truncate text-sm font-medium text-card-foreground">{asset.title || asset.original_name}</h3>
            <p className="mt-1 truncate text-xs text-muted-foreground">{[asset.camera, asset.lens].filter(Boolean).join(" · ")}</p>
          </div>
        </AppLink>
      ))}
    </div>
  );
}

export function ThemeGallery(props: ThemeGalleryProps) {
  const [sheetOpen, setSheetOpen] = useState(false);
  const [fabOpen, setFabOpen] = useState(false);
  const fabRef = useRef<HTMLDivElement>(null);
  const assets = props.data?.items ?? [];
  const total = props.data?.total ?? 0;
  const isSingleAlbum = Boolean(props.album && props.album !== "/" && props.album !== "all");
  const baseStyle = useMemo<"waterfall" | "single">(() => {
    if (props.preferredStyle) return props.preferredStyle;
    if (isSingleAlbum && props.data) return total > 10 ? "waterfall" : "single";
    return props.systemStyle ?? "waterfall";
  }, [isSingleAlbum, props.data, props.preferredStyle, props.systemStyle, total]);
  const [currentStyle, setCurrentStyle] = useState<"waterfall" | "single">(baseStyle);
  const [userOverridden, setUserOverridden] = useState(false);
  const activeCount = props.selectedCameras.length + props.selectedLenses.length + props.selectedTags.length;
  const hasActivity = activeCount > 0 || props.sort !== undefined;
  const enableFilters = !isSingleAlbum;

  useEffect(() => {
    if (!userOverridden) setCurrentStyle(baseStyle);
  }, [baseStyle, userOverridden]);

  useEffect(() => {
    if (!fabOpen) return;
    const handleClick = (event: MouseEvent) => {
      if (fabRef.current && !fabRef.current.contains(event.target as Node)) setFabOpen(false);
    };
    document.addEventListener("pointerdown", handleClick);
    return () => document.removeEventListener("pointerdown", handleClick);
  }, [fabOpen]);

  function cycleSort() {
    props.onSortChange(props.sort === undefined ? "desc" : props.sort === "desc" ? "asc" : undefined);
  }

  const sortLabel = props.sort === "desc" ? "最新" : props.sort === "asc" ? "最早" : "默认";
  const toggleStyle = () => {
    setUserOverridden(true);
    setCurrentStyle((value) => (value === "waterfall" ? "single" : "waterfall"));
    setFabOpen(false);
  };
  const fabItems: FabItem[] = enableFilters
    ? [
        {
          icon: currentStyle === "waterfall" ? <Rows className="h-[15px] w-[15px]" /> : <LayoutGrid className="h-[15px] w-[15px]" />,
          label: currentStyle === "waterfall" ? "单列" : "瀑布流",
          active: false,
          onClick: toggleStyle
        },
        {
          icon: <ArrowUpDown className="h-[15px] w-[15px]" />,
          label: sortLabel,
          active: props.sort !== undefined,
          onClick: cycleSort
        },
        {
          icon: <SlidersHorizontal className="h-[15px] w-[15px]" />,
          label: "筛选",
          active: activeCount > 0,
          badge: activeCount,
          onClick: () => {
            setSheetOpen(true);
            setFabOpen(false);
          }
        }
      ]
    : [
        {
          icon: currentStyle === "waterfall" ? <Rows className="h-[15px] w-[15px]" /> : <LayoutGrid className="h-[15px] w-[15px]" />,
          label: currentStyle === "waterfall" ? "单列" : "瀑布流",
          active: false,
          onClick: toggleStyle
        }
      ];

  return (
    <div className="min-h-screen bg-background px-3 pt-16 pb-16">
      {props.loading ? (
        <div className="flex min-h-[60vh] items-center justify-center">
          <Spin />
        </div>
      ) : currentStyle === "waterfall" ? (
        <div className="mx-auto w-full max-w-[1400px]">
          {assets.length ? <VirtualWaterfallGallery assets={assets} previewSearch={props.previewSearch} /> : <Empty description="暂无匹配的图片" />}
        </div>
      ) : (
        <SingleGallery assets={assets} previewSearch={props.previewSearch} />
      )}

      {props.error ? <div className="py-8 text-center text-sm text-destructive">图库加载失败</div> : null}

      {props.data && props.data.total > props.data.page_size ? (
        <div className="flex justify-center py-8">
          <Pagination
            current={props.data.page}
            pageSize={props.data.page_size}
            total={props.data.total}
            showSizeChanger={false}
            onChange={props.onPageChange}
          />
        </div>
      ) : null}

      <div ref={fabRef} className="fixed bottom-6 right-5 z-40 flex flex-col items-end gap-3">
        {fabItems.map((item, index) => (
          <div
            key={item.label}
            className="flex items-center gap-2.5"
            style={{
              opacity: fabOpen ? 1 : 0,
              transform: fabOpen ? "translateY(0)" : "translateY(10px)",
              pointerEvents: fabOpen ? "auto" : "none",
              transition: "opacity 0.18s ease, transform 0.18s ease",
              transitionDelay: fabOpen ? `${index * 35}ms` : "0ms"
            }}
          >
            <span className="whitespace-nowrap rounded-md border border-border/60 bg-background/90 px-2.5 py-1 text-[11px] font-medium tracking-wide text-foreground shadow-sm backdrop-blur-md">
              {item.label}
            </span>
            <button
              type="button"
              onClick={item.onClick}
              className={cn(
                "relative flex h-10 w-10 shrink-0 items-center justify-center rounded-full border shadow-sm transition-opacity duration-150 active:opacity-60",
                item.active ? "border-primary bg-primary text-primary-foreground" : "border-border/60 bg-background/90 text-foreground backdrop-blur-md"
              )}
            >
              {item.icon}
              {item.badge ? (
                <span className="absolute -right-0.5 -top-0.5 flex h-3.5 min-w-3.5 items-center justify-center rounded-full bg-primary text-[8px] font-bold text-primary-foreground">
                  {item.badge > 9 ? "9+" : item.badge}
                </span>
              ) : null}
            </button>
          </div>
        ))}

        <button
          type="button"
          aria-label={fabOpen ? "关闭菜单" : "打开菜单"}
          aria-expanded={fabOpen}
          onClick={() => setFabOpen((value) => !value)}
          className={cn(
            "gallery-floating-control relative flex h-12 w-12 items-center justify-center rounded-full border border-border/80 bg-background/95 text-foreground backdrop-blur-xl transition-transform duration-200 active:scale-95",
            fabOpen && "rotate-45"
          )}
        >
          {fabOpen ? <X className="h-4 w-4" /> : <Settings2 className="h-4 w-4" />}
          {!fabOpen && hasActivity ? <span className="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-primary" /> : null}
        </button>
      </div>

      {enableFilters ? (
        <div
          className={cn(
            "fixed inset-0 z-50 bg-media-scrim/30 transition-opacity duration-300",
            sheetOpen ? "pointer-events-auto opacity-100" : "pointer-events-none opacity-0"
          )}
          onClick={() => setSheetOpen(false)}
        >
          <aside
            className={cn(
              "absolute bottom-0 right-0 flex max-h-[90dvh] w-full flex-col rounded-t-2xl bg-background shadow-2xl transition-transform duration-300 md:bottom-auto md:top-0 md:h-full md:max-h-none md:w-96 md:rounded-none",
              sheetOpen ? "translate-y-0 md:translate-x-0" : "translate-y-full md:translate-x-full md:translate-y-0"
            )}
            onClick={(event) => event.stopPropagation()}
          >
            <div className="absolute left-1/2 top-2.5 h-1 w-10 -translate-x-1/2 rounded-full bg-muted-foreground/25 md:hidden" />
            <header className="flex shrink-0 items-center justify-between border-b px-5 py-4">
              <h2 className="text-base font-semibold">筛选 & 排序</h2>
              {hasActivity ? (
                <button type="button" onClick={props.onReset} className="text-xs text-destructive hover:underline">
                  清除全部
                </button>
              ) : null}
            </header>
            <div className="flex-1 overflow-y-auto overscroll-contain px-5 py-5">
              <FilterPanel {...props} />
            </div>
            <div className="shrink-0 border-t bg-background px-5 py-4 md:hidden">
              <button
                type="button"
                onClick={() => setSheetOpen(false)}
                className="w-full rounded-xl bg-primary py-3 text-sm font-medium text-primary-foreground transition-opacity active:opacity-70"
              >
                完成
              </button>
            </div>
          </aside>
        </div>
      ) : null}
    </div>
  );
}
