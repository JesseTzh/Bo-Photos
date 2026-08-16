import { Alert, Spin, App as AntApp } from "antd";
import { ArrowLeft, Copy, Download, Expand, Link as LinkIcon } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { albumPublicHref } from "../features/albums/routes";
import { useAsset, useGallery } from "../features/assets/api";
import { AssetMedia } from "../features/assets/asset-media";
import { isVideoAsset } from "../features/assets/schema";
import { fetchAdminAsset, usePrivateAssets } from "../features/assets/admin-api";
import { useVisit } from "../features/site/api";
import { usePublicSettings } from "../features/site/api";
import { useQuery } from "@tanstack/react-query";
import { usePagedWheelNavigation } from "../shared/hooks/use-paged-wheel-navigation";

function csv(search: URLSearchParams, key: string) {
  return search.get(key)?.split(",").filter(Boolean) ?? [];
}

export function PreviewPage() {
  useVisit("gallery");
  const { id } = useParams();
  const [search] = useSearchParams();
  const navigate = useNavigate();
  const { message } = AntApp.useApp();
  const privateMode = search.get("private") === "1";
  const asset = useAsset(id, !privateMode);
  const privateAsset = useQuery({
    queryKey: ["assets", "private", id],
    queryFn: () => fetchAdminAsset(id!),
    enabled: Boolean(id && privateMode)
  });
  const album = search.get("album") ?? undefined;
  const cameras = csv(search, "cameras");
  const lenses = csv(search, "lenses");
  const selectedTags = csv(search, "tags");
  const tagsOperator = search.get("tags_operator") === "or" ? "or" : "and";
  const sort = search.get("sort") as "asc" | "desc" | null;
  const gallery = useGallery({
    page: 1,
    pageSize: 200,
    album,
    cameras,
    lenses,
    tags: selectedTags,
    tagsOperator,
    sortByShootTime: sort ?? undefined
  });
  const privateGallery = usePrivateAssets({ page: 1, pageSize: 200 });
  const settings = usePublicSettings();
  const [lightbox, setLightbox] = useState(false);
  const desktopMediaRef = useRef<HTMLDivElement>(null);
  const mobileMediaRef = useRef<HTMLDivElement>(null);

  const imageList = privateMode ? (privateGallery.data?.items ?? []) : (gallery.data?.items ?? []);
  const currentIndex = useMemo(() => imageList.findIndex((item) => item.id === id), [id, imageList]);
  const previous = currentIndex > 0 ? imageList[currentIndex - 1] : undefined;
  const next = currentIndex >= 0 && currentIndex < imageList.length - 1 ? imageList[currentIndex + 1] : undefined;
  const contextSearch = search.toString();
  const contextSuffix = contextSearch ? `?${contextSearch}` : "";
  const navigatePage = useCallback((direction: -1 | 1) => {
    const destination = direction < 0 ? previous : next;
    if (destination) navigate(`/preview/${destination.id}${contextSuffix}`, { replace: true });
  }, [contextSuffix, navigate, next, previous]);

  usePagedWheelNavigation(desktopMediaRef, navigatePage, !lightbox);
  usePagedWheelNavigation(mobileMediaRef, navigatePage, !lightbox);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "ArrowLeft" && previous) navigate(`/preview/${previous.id}${contextSuffix}`, { replace: true });
      if (event.key === "ArrowRight" && next) navigate(`/preview/${next.id}${contextSuffix}`, { replace: true });
      if (event.key === "Escape") setLightbox(false);
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [contextSuffix, navigate, next, previous]);

  const currentAsset = privateMode ? privateAsset : asset;
  if (currentAsset.isPending) {
    return <div className="page-loading"><Spin size="large" /></div>;
  }
  if (!currentAsset.data || currentAsset.error) {
    return <main className="min-h-screen bg-background p-6"><Alert type="error" message="图片不存在或暂不可见" showIcon /></main>;
  }
  const item = currentAsset.data;
  const imageUrl = item.preview_url || item.thumbnail_url || item.original_url || "";
  const video = isVideoAsset(item);
  const directUrl = item.video_url || imageUrl;
  const downloadEnabled = settings.data?.public_original_download !== false;
  const exifRows = [
    ["相机", item.camera],
    ["镜头", item.lens],
    ["日期", item.shoot_at ? new Date(item.shoot_at).toLocaleDateString("zh-CN") : undefined],
    ["光圈", item.aperture],
    ["快门", item.exposure_time],
    ["焦距", item.focal_length],
    ["ISO", item.iso],
    ["尺寸", item.width && item.height ? `${item.width} x ${item.height}` : undefined]
  ].filter((entry): entry is [string, string] => Boolean(entry[1]));

  function close() {
    if (window.history.length > 1) navigate(-1);
    else navigate(album ? albumPublicHref(album) : "/gallery");
  }

  async function copyText(value?: string, success = "已复制") {
    if (!value) return;
    await navigator.clipboard.writeText(value);
    message.success(success);
  }

  return (
    <main className="bg-background">
      <div className="hidden h-screen w-full flex-row overflow-hidden bg-background lg:flex">
        <div ref={desktopMediaRef} className="relative flex min-w-0 flex-1 items-center justify-center overflow-hidden">
          <div className="flex h-full w-full items-center justify-center">
            <AssetMedia asset={item} controls={video} autoPlay={video} className="max-h-full max-w-full object-contain" />
          </div>
        </div>

        <aside className="w-[300px] flex-shrink-0 overflow-y-auto border-l border-border bg-card xl:w-[340px]">
          <div className="sticky top-0 z-10 flex items-start justify-between gap-3 border-b border-border bg-card px-6 py-5">
            <h1 className="line-clamp-2 flex-1 text-lg font-bold leading-snug text-card-foreground">{item.title || item.original_name}</h1>
            <button
              onClick={close}
              className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
              aria-label="返回"
              type="button"
            >
              <ArrowLeft size={18} />
            </button>
          </div>
          <PreviewInfo
            description={item.description}
            exifRows={exifRows}
            imageUrl={video ? directUrl : item.original_url || directUrl}
            shareUrl={window.location.href}
            downloadUrl={downloadEnabled ? item.original_url : undefined}
            onCopy={copyText}
            onFullscreen={video ? undefined : () => setLightbox(true)}
          />
        </aside>
      </div>

      <div className="min-h-screen bg-background lg:hidden">
        <div className="sticky top-0 z-20 flex items-center justify-between border-b border-border bg-background/90 px-4 py-3 backdrop-blur-sm">
          <h1 className="line-clamp-1 flex-1 pr-3 text-sm font-bold leading-snug text-foreground">{item.title || item.original_name}</h1>
          <button onClick={close} className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted" aria-label="返回" type="button">
            <ArrowLeft size={18} />
          </button>
        </div>

        <div ref={mobileMediaRef} className="relative w-full overflow-hidden bg-muted/20">
          <AssetMedia asset={item} controls={video} autoPlay={video} className="h-auto w-full" />
        </div>

        <div className="border-t border-border bg-card">
          <PreviewInfo
            description={item.description}
            exifRows={exifRows}
            imageUrl={video ? directUrl : item.original_url || directUrl}
            shareUrl={window.location.href}
            downloadUrl={downloadEnabled ? item.original_url : undefined}
            onCopy={copyText}
            onFullscreen={video ? undefined : () => setLightbox(true)}
          />
        </div>
      </div>

      {lightbox && !video ? (
        <div className="fixed inset-0 z-[80] flex items-center justify-center bg-lightbox-surface/95 p-4" onClick={() => setLightbox(false)}>
          <img src={imageUrl} alt={item.title || item.original_name} className="max-h-full max-w-full object-contain" />
        </div>
      ) : null}
    </main>
  );
}

function PreviewInfo({
  description,
  exifRows,
  imageUrl,
  shareUrl,
  downloadUrl,
  onCopy,
  onFullscreen
}: {
  description?: string;
  exifRows: Array<[string, string]>;
  imageUrl: string;
  shareUrl: string;
  downloadUrl?: string;
  onCopy: (value?: string, success?: string) => void;
  onFullscreen?: () => void;
}) {
  return (
    <div className="space-y-5 px-6 py-5">
      {description ? <p className="text-sm leading-relaxed text-muted-foreground">{description}</p> : null}
      {exifRows.length > 0 ? (
        <section>
          <div className="mb-3 flex items-center gap-2">
            <span className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">EXIF</span>
            <div className="h-px flex-1 bg-border" />
          </div>
          <div className="space-y-2">
            {exifRows.map(([label, value]) => (
              <div key={label} className="flex items-center gap-3 rounded-lg bg-muted/30 px-3 py-2">
                <span className="w-16 flex-shrink-0 text-xs text-muted-foreground">{label}</span>
                <span className="truncate text-xs font-medium text-foreground">{value}</span>
              </div>
            ))}
          </div>
        </section>
      ) : null}
      <section>
        <div className="border-t border-border/60 pt-4">
          <div className="grid grid-cols-2 gap-2">
            <ActionButton icon={<Copy size={14} />} label="复制直链" onClick={() => onCopy(imageUrl, "图片链接已复制")} />
            <ActionButton icon={<LinkIcon size={14} />} label="分享链接" onClick={() => onCopy(shareUrl, "分享链接已复制")} />
            {downloadUrl ? <ActionButton icon={<Download size={14} />} label="下载" onClick={() => window.open(downloadUrl, "_blank")} /> : null}
            {onFullscreen ? <ActionButton icon={<Expand size={14} />} label="全屏" onClick={onFullscreen} /> : null}
          </div>
        </div>
      </section>
    </div>
  );
}

function ActionButton({ icon, label, onClick }: { icon: React.ReactNode; label: string; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="flex items-center gap-2 rounded-lg border border-transparent bg-muted/60 px-3 py-2.5 text-xs font-medium text-muted-foreground transition-all duration-150 hover:border-accent hover:bg-accent hover:text-accent-foreground active:scale-[0.98]"
      type="button"
    >
      <span className="flex-shrink-0">{icon}</span>
      {label}
    </button>
  );
}
