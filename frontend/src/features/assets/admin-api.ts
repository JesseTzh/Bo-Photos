import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "../../api/client";
import type { AdminAssetFilterOptions, Asset, AssetPage, AssetStatus } from "./schema";

export interface AdminAssetQuery {
  page: number;
  pageSize?: number;
  status?: AssetStatus;
  visible?: boolean;
  private?: boolean;
  featured?: boolean;
  camera?: string;
  lens?: string;
  exposure_time?: string;
  aperture?: string;
  iso?: string;
  album?: string;
  tags?: string[];
  tagsOperator?: "and" | "or";
  title?: string;
}

export interface AssetUpdate {
  title?: string;
  description?: string;
  width?: number;
  height?: number;
  longitude?: number | null;
  latitude?: number | null;
  shoot_at?: string;
  camera?: string;
  lens?: string;
  exposure_time?: string;
  aperture?: string;
  iso?: string;
  focal_length?: string;
  exif_json?: string;
  visible?: boolean;
  private?: boolean;
  show_on_homepage?: boolean;
  featured?: boolean;
  sort?: number;
}

function adminSearch(query: AdminAssetQuery) {
  const search = new URLSearchParams({
    page: String(query.page),
    page_size: String(query.pageSize ?? 20)
  });
  for (const [key, value] of Object.entries(query)) {
    if (key === "page" || key === "pageSize" || value === undefined || value === "") continue;
    if (Array.isArray(value)) {
      if (value.length) search.set(key, value.join(","));
      continue;
    }
    if (key === "tagsOperator") search.set("tags_operator", String(value));
    else search.set(key, String(value));
  }
  return search.toString();
}

export function useAdminAssets(query: AdminAssetQuery) {
  return useQuery({
    queryKey: ["assets", "admin", query],
    queryFn: () => apiRequest<AssetPage>(`/admin/assets?${adminSearch(query)}`)
  });
}

export function usePrivateAssets(query: { page: number; pageSize?: number }) {
  const search = new URLSearchParams({
    page: String(query.page),
    page_size: String(query.pageSize ?? 16)
  });
  return useQuery({
    queryKey: ["assets", "private", query],
    queryFn: () => apiRequest<AssetPage>(`/admin/assets/private?${search.toString()}`)
  });
}

export function fetchAdminAsset(id: string) {
  return apiRequest<Asset>(`/admin/assets/${id}`);
}

export function useAdminAssetFilterOptions() {
  return useQuery({
    queryKey: ["assets", "admin", "filters"],
    queryFn: () => apiRequest<AdminAssetFilterOptions>("/admin/assets/filters")
  });
}

export async function uploadAsset(file: File) {
  const form = new FormData();
  form.append("file", file);
  return apiRequest<{ id: string; status: AssetStatus; duplicate_asset_ids: string[] }>(
    "/admin/assets",
    { method: "POST", body: form }
  );
}

function useAssetMutation<T>(mutationFn: (input: T) => Promise<unknown>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["assets"] })
  });
}

export function useUpdateAsset() {
  return useAssetMutation(({ id, input }: { id: string; input: AssetUpdate }) =>
    apiRequest<void>(`/admin/assets/${id}`, { method: "PATCH", body: JSON.stringify(input) })
  );
}

export function useAssetAlbums(assetId?: string) {
  return useQuery({
    queryKey: ["asset-albums", assetId],
    queryFn: () => apiRequest<{ album_ids: string[] }>(`/admin/assets/${assetId}/albums`),
    enabled: Boolean(assetId)
  });
}

export function useSaveAssetAlbums() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ assetId, albumIds }: { assetId: string; albumIds: string[] }) =>
      apiRequest<void>(`/admin/assets/${assetId}/albums`, {
        method: "PUT",
        body: JSON.stringify({ album_ids: albumIds })
      }),
    onSuccess: (_, input) => {
      client.invalidateQueries({ queryKey: ["asset-albums", input.assetId] });
      client.invalidateQueries({ queryKey: ["albums"] });
      client.invalidateQueries({ queryKey: ["assets"] });
    }
  });
}

export function useDeleteAssets() {
  return useAssetMutation((ids: string[]) =>
    apiRequest<void>("/admin/assets", { method: "DELETE", body: JSON.stringify({ ids }) })
  );
}

export function useRestoreAsset() {
  return useAssetMutation((id: string) =>
    apiRequest<void>(`/admin/assets/${id}/restore`, { method: "POST" })
  );
}

export function usePurgeAsset() {
  return useAssetMutation((id: string) =>
    apiRequest<void>(`/admin/assets/${id}/purge`, { method: "POST" })
  );
}

export function useRetryAsset() {
  return useAssetMutation((id: string) =>
    apiRequest<void>(`/admin/assets/${id}/retry`, { method: "POST" })
  );
}
