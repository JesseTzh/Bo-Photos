import type { CSSProperties } from "react";
import { isVideoAsset, type Asset } from "./schema";

interface AssetMediaProps {
  asset: Asset;
  className?: string;
  style?: CSSProperties;
  controls?: boolean;
  autoPlay?: boolean;
  loop?: boolean;
  muted?: boolean;
  loading?: "eager" | "lazy";
  onLoad?: () => void;
}

export function AssetMedia({
  asset,
  className,
  style,
  controls = false,
  autoPlay = false,
  loop = false,
  muted = true,
  loading = "lazy",
  onLoad
}: AssetMediaProps) {
  const label = asset.title || asset.original_name;
  if (isVideoAsset(asset) && asset.video_url) {
    return (
      <video
        src={asset.video_url}
        aria-label={label}
        className={className}
        style={style}
        controls={controls}
        autoPlay={autoPlay}
        loop={loop}
        muted={muted}
        playsInline
        preload={autoPlay ? "auto" : "metadata"}
        onLoadedData={onLoad}
      />
    );
  }
  const source = asset.preview_url || asset.thumbnail_url || asset.original_url;
  return source ? (
    <img
      src={source}
      alt={label}
      className={className}
      style={style}
      loading={loading}
      decoding="async"
      onLoad={onLoad}
    />
  ) : null;
}
