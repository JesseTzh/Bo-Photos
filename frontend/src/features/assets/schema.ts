export type AssetStatus = "processing" | "ready" | "failed" | "deleted" | "purged";

export interface Asset {
  id: string;
  status: AssetStatus;
  original_name: string;
  mime_type?: string;
  byte_size?: number;
  width: number;
  height: number;
  title?: string;
  description?: string;
  longitude?: number;
  latitude?: number;
  blurhash?: string;
  exif_json?: string;
  shoot_at?: string;
  camera?: string;
  lens?: string;
  exposure_time?: string;
  aperture?: string;
  iso?: string;
  focal_length?: string;
  visible: boolean;
  private: boolean;
  show_on_homepage: boolean;
  featured: boolean;
  sort: number;
  thumbnail_url?: string;
  preview_url?: string;
  original_url?: string;
  video_url?: string;
  error_code?: string;
  created_at?: string;
  updated_at?: string;
}

export function isVideoAsset(asset: Pick<Asset, "mime_type">) {
  return asset.mime_type?.startsWith("video/") ?? false;
}

export interface AssetPage {
  page: number;
  page_size: number;
  total: number;
  items: Asset[];
}

export interface AssetFilters {
  cameras: string[];
  lenses: string[];
}

export interface AdminAssetFilterOptions {
  cameras: string[];
  lenses: string[];
  exposure_times: string[];
  apertures: string[];
  isos: string[];
}

export interface GalleryQuery {
  page: number;
  pageSize?: number;
  cameras?: string[];
  lenses?: string[];
  tags?: string[];
  tagsOperator?: "and" | "or";
  album?: string;
  featured?: boolean;
  homepage?: boolean;
  private?: boolean;
  sortByShootTime?: "asc" | "desc";
}
