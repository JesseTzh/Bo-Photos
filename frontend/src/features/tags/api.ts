import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiRequest } from "../../api/client";

export interface TagNode {
  id: string;
  name: string;
  category?: string;
  parent_id?: string;
  detail?: string;
  children?: TagNode[];
}

export function useTags(admin = false) {
  return useQuery({
    queryKey: ["tags", admin],
    queryFn: () => apiRequest<{ items: TagNode[] }>(`${admin ? "/admin" : "/public"}/tags`)
  });
}

export function flattenTags(nodes: TagNode[], depth = 0): Array<TagNode & { depth: number }> {
  return nodes.flatMap((node) => [{ ...node, depth }, ...flattenTags(node.children ?? [], depth + 1)]);
}

export function useCreateTag() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (input: { name: string; category?: string; parent_id?: string; detail?: string }) =>
      apiRequest<TagNode>("/admin/tags", { method: "POST", body: JSON.stringify(input) }),
    onSuccess: () => client.invalidateQueries({ queryKey: ["tags"] })
  });
}

export function useUpdateTag() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: Record<string, string> }) =>
      apiRequest<void>(`/admin/tags/${id}`, { method: "PUT", body: JSON.stringify(input) }),
    onSuccess: () => client.invalidateQueries({ queryKey: ["tags"] })
  });
}

export function useMoveTag() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, parentId }: { id: string; parentId: string }) =>
      apiRequest<void>(`/admin/tags/${id}/parent`, {
        method: "PUT",
        body: JSON.stringify({ parent_id: parentId })
      }),
    onSuccess: () => client.invalidateQueries({ queryKey: ["tags"] })
  });
}

export function useDeleteTag() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiRequest<void>(`/admin/tags/${id}`, { method: "DELETE" }),
    onSuccess: () => client.invalidateQueries({ queryKey: ["tags"] })
  });
}

export function useAssetTags(assetId?: string) {
  return useQuery({
    queryKey: ["asset-tags", assetId],
    queryFn: () => apiRequest<{ tag_ids: string[] }>(`/admin/assets/${assetId}/tags`),
    enabled: Boolean(assetId)
  });
}

export function useSaveAssetTags() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ assetId, tagIds }: { assetId: string; tagIds: string[] }) =>
      apiRequest<void>(`/admin/assets/${assetId}/tags`, {
        method: "PUT",
        body: JSON.stringify({ tag_ids: tagIds })
      }),
    onSuccess: (_, input) => client.invalidateQueries({ queryKey: ["asset-tags", input.assetId] })
  });
}
