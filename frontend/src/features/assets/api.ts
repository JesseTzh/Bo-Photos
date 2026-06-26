import { useQuery } from "@tanstack/react-query";
import { apiRequest } from "../../api/client";
import type { Asset, AssetFilters, AssetPage, GalleryQuery } from "./schema";

function gallerySearch(query: GalleryQuery) {
  const search = new URLSearchParams({
    page: String(query.page),
    page_size: String(query.pageSize ?? 16)
  });
  if (query.cameras?.length) search.set("cameras", query.cameras.join(","));
  if (query.lenses?.length) search.set("lenses", query.lenses.join(","));
  if (query.tags?.length) search.set("tags", query.tags.join(","));
  if (query.tagsOperator) search.set("tags_operator", query.tagsOperator);
  if (query.album) search.set("album", query.album);
  if (query.featured !== undefined) search.set("featured", String(query.featured));
  if (query.homepage) search.set("homepage", "true");
  if (query.sortByShootTime) search.set("sort_by_shoot_time", query.sortByShootTime);
  return search.toString();
}

export function useGallery(query: GalleryQuery) {
  return useQuery({
    queryKey: ["assets", "public", query],
    queryFn: () => apiRequest<AssetPage>(`/public/assets?${gallerySearch(query)}`)
  });
}

export function useAsset(id?: string) {
  return useQuery({
    queryKey: ["assets", "public", id],
    queryFn: () => apiRequest<Asset>(`/public/assets/${id}`),
    enabled: Boolean(id)
  });
}

export function useAssetFilters() {
  return useQuery({
    queryKey: ["assets", "filters"],
    queryFn: () => apiRequest<AssetFilters>("/public/assets/filters"),
    staleTime: 60_000
  });
}
