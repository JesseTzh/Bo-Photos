import { useCallback, useEffect, useRef, useState } from "react";
import { fetchAdminAsset, uploadAsset } from "./admin-api";
import type { AssetStatus } from "./schema";

export type UploadQueueStatus = "queued" | "uploading" | AssetStatus;

export interface UploadMetadataDraft {
  album_ids: string[];
  tag_ids: string[];
  title?: string;
  description?: string;
  visible: boolean;
  private: boolean;
  show_on_homepage: boolean;
  sort: number;
  shoot_at?: string;
  camera?: string;
  lens?: string;
  exposure_time?: string;
  aperture?: string;
  iso?: string;
  focal_length?: string;
  latitude?: number;
  longitude?: number;
  exif_json?: string;
}

export interface UploadQueueItem {
  key: string;
  file: File;
  previewUrl: string;
  assetId?: string;
  status: UploadQueueStatus;
  duplicateAssetIds: string[];
  error?: string;
  metadata: UploadMetadataDraft;
  metadataApplied: boolean;
}

const sleep = (milliseconds: number) => new Promise((resolve) => setTimeout(resolve, milliseconds));

export const defaultUploadMetadata: UploadMetadataDraft = {
  album_ids: [],
  tag_ids: [],
  visible: true,
  private: false,
  show_on_homepage: true,
  sort: 0
};

function copyMetadata(metadata: UploadMetadataDraft): UploadMetadataDraft {
  return {
    ...metadata,
    album_ids: [...metadata.album_ids],
    tag_ids: [...metadata.tag_ids]
  };
}

export function useUploadQueue(
  concurrency = 2,
  defaults: UploadMetadataDraft = defaultUploadMetadata,
  onAssetReady?: (item: UploadQueueItem) => Promise<void> | void
) {
  const [items, setItems] = useState<UploadQueueItem[]>([]);
  const itemsRef = useRef<UploadQueueItem[]>([]);
  const running = useRef(0);
  const pending = useRef<UploadQueueItem[]>([]);
  const previewUrls = useRef(new Set<string>());
  const onAssetReadyRef = useRef(onAssetReady);
  const defaultsRef = useRef(defaults);

  useEffect(() => {
    onAssetReadyRef.current = onAssetReady;
  }, [onAssetReady]);

  useEffect(() => {
    defaultsRef.current = defaults;
  }, [defaults]);

  useEffect(() => {
    return () => {
      previewUrls.current.forEach((url) => URL.revokeObjectURL(url));
      previewUrls.current.clear();
    };
  }, []);

  const setQueueItems = useCallback((updater: (current: UploadQueueItem[]) => UploadQueueItem[]) => {
    setItems((current) => {
      const next = updater(current);
      itemsRef.current = next;
      return next;
    });
  }, []);

  const update = useCallback((key: string, patch: Partial<UploadQueueItem>) => {
    setQueueItems((current) => current.map((item) => item.key === key ? { ...item, ...patch } : item));
  }, [setQueueItems]);

  const updateMetadata = useCallback((key: string, patch: Partial<UploadMetadataDraft>) => {
    setQueueItems((current) => current.map((item) => item.key === key ? {
      ...item,
      metadata: {
        ...item.metadata,
        ...patch,
        album_ids: patch.album_ids ? [...patch.album_ids] : item.metadata.album_ids,
        tag_ids: patch.tag_ids ? [...patch.tag_ids] : item.metadata.tag_ids
      }
    } : item));
  }, [setQueueItems]);

  const poll = useCallback(async (item: UploadQueueItem, assetId: string) => {
    for (;;) {
      await sleep(1200);
      const asset = await fetchAdminAsset(assetId);
      update(item.key, { status: asset.status, error: asset.error_code });
      if (asset.status === "ready") {
        const latest = itemsRef.current.find((current) => current.key === item.key) ?? item;
        try {
          await onAssetReadyRef.current?.({ ...latest, assetId, status: asset.status, error: asset.error_code });
          update(item.key, { metadataApplied: true });
        } catch (error) {
          update(item.key, {
            status: "failed",
            error: error instanceof Error ? error.message : "上传后写入信息失败"
          });
        }
        return;
      }
      if (asset.status === "failed") return;
    }
  }, [update]);

  const drain = useCallback(() => {
    while (running.current < concurrency && pending.current.length > 0) {
      const item = pending.current.shift()!;
      running.current += 1;
      update(item.key, { status: "uploading" });
      void (async () => {
        try {
          const accepted = await uploadAsset(item.file);
          update(item.key, {
            assetId: accepted.id,
            status: accepted.status,
            duplicateAssetIds: accepted.duplicate_asset_ids ?? []
          });
          await poll(item, accepted.id);
        } catch (error) {
          update(item.key, {
            status: "failed",
            error: error instanceof Error ? error.message : "上传失败"
          });
        } finally {
          running.current -= 1;
          drain();
        }
      })();
    }
  }, [concurrency, poll, update]);

  const addFiles = useCallback((files: File[], metadataDefaults = defaultsRef.current) => {
    const queued = files.map((file, index): UploadQueueItem => {
      const previewUrl = URL.createObjectURL(file);
      previewUrls.current.add(previewUrl);
      return {
        key: `${file.name}-${file.size}-${file.lastModified}-${Date.now()}-${index}`,
        file,
        previewUrl,
        status: "queued",
        duplicateAssetIds: [],
        metadata: copyMetadata(metadataDefaults),
        metadataApplied: false
      };
    });
    pending.current.push(...queued);
    setQueueItems((current) => [...queued, ...current]);
    queueMicrotask(drain);
  }, [drain, setQueueItems]);

  const clearFinished = useCallback(() => {
    setQueueItems((current) => {
      const remaining = current.filter((item) => !["ready", "failed"].includes(item.status));
      current
        .filter((item) => ["ready", "failed"].includes(item.status))
        .forEach((item) => {
          URL.revokeObjectURL(item.previewUrl);
          previewUrls.current.delete(item.previewUrl);
        });
      return remaining;
    });
  }, [setQueueItems]);

  return { items, addFiles, clearFinished, update, updateMetadata };
}
