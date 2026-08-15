import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "../../api/client";

export interface Album {
  id: string;
  name: string;
  album_value: string;
  detail: string;
  theme: string;
  visible: boolean;
  sort: number;
  random_show: boolean;
  license: string;
  cover_asset_id: string;
  cover_url: string;
  image_sorting: number;
  asset_ids: string[];
  asset_count: number;
}

export type AlbumInput = Omit<Album, "id" | "cover_url" | "asset_ids" | "asset_count"> & {
  cover_asset_id?: string;
};

export interface AlbumAsset {
  id: string;
  original_name: string;
  title?: string;
  width: number;
  height: number;
  visible: boolean;
  featured: boolean;
  sort: number;
  preview_url?: string;
  thumbnail_url?: string;
}

export function useAlbums(admin = false) {
  return useQuery({
    queryKey: ["albums", admin],
    queryFn: () => apiRequest<{ items: Album[] }>(`${admin ? "/admin" : "/public"}/albums`)
  });
}

export function useAlbum(value?: string) {
  return useQuery({
    queryKey: ["albums", "public", value],
    queryFn: () => apiRequest<Album>(`/public/albums/${value}`),
    enabled: Boolean(value)
  });
}

export function useAdminAlbum(id?: string) {
  return useQuery({
    queryKey: ["albums", "admin", id],
    queryFn: () => apiRequest<Album>(`/admin/albums/${id}`),
    enabled: Boolean(id)
  });
}

export function useSaveAlbum() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id?: string; input: AlbumInput }) =>
      apiRequest<Album>(id ? `/admin/albums/${id}` : "/admin/albums", {
        method: id ? "PUT" : "POST",
        body: JSON.stringify(input)
      }),
    onSuccess: () => client.invalidateQueries({ queryKey: ["albums"] })
  });
}

export function useDeleteAlbum() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiRequest<void>(`/admin/albums/${id}`, { method: "DELETE" }),
    onSuccess: () => client.invalidateQueries({ queryKey: ["albums"] })
  });
}

export function useSaveAlbumOrder() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (ids: string[]) => apiRequest<void>("/admin/albums/sort", {
      method: "PUT",
      body: JSON.stringify({ ids })
    }),
    onSuccess: () => client.invalidateQueries({ queryKey: ["albums"] })
  });
}

export function useSetAlbumCover() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ albumId, assetId }: { albumId: string; assetId: string }) =>
      apiRequest<void>(`/admin/albums/${albumId}/cover`, {
        method: "PUT",
        body: JSON.stringify({ asset_id: assetId })
      }),
    onSuccess: () => client.invalidateQueries({ queryKey: ["albums"] })
  });
}

export function useReplaceAlbumAssets() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, assetIds }: { id: string; assetIds: string[] }) =>
      apiRequest<void>(`/admin/albums/${id}/assets`, {
        method: "PUT",
        body: JSON.stringify({ asset_ids: assetIds })
      }),
    onSuccess: (_, input) => {
      client.invalidateQueries({ queryKey: ["albums"] });
      client.invalidateQueries({ queryKey: ["albums", input.id, "assets"] });
      client.invalidateQueries({ queryKey: ["assets"] });
    }
  });
}

export function useAlbumAssets(albumId?: string) {
  return useQuery({
    queryKey: ["albums", albumId, "assets"],
    queryFn: () => apiRequest<{ items: AlbumAsset[] }>(`/admin/albums/${albumId}/assets`),
    enabled: Boolean(albumId)
  });
}

export function useSaveAlbumAssetSort() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ albumId, orders }: { albumId: string; orders: Array<{ asset_id: string; sort: number }> }) =>
      apiRequest<void>(`/admin/albums/${albumId}/assets/sort`, {
        method: "PUT",
        body: JSON.stringify({ orders })
      }),
    onSuccess: (_, input) => {
      client.invalidateQueries({ queryKey: ["albums", input.albumId, "assets"] });
      client.invalidateQueries({ queryKey: ["assets"] });
    }
  });
}

export function useResetAlbumAssetSort() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (albumId: string) =>
      apiRequest<void>(`/admin/albums/${albumId}/assets/sort/reset`, { method: "POST" }),
    onSuccess: (_, albumId) => client.invalidateQueries({ queryKey: ["albums", albumId, "assets"] })
  });
}
