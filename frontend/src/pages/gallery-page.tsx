import { useParams, useSearchParams } from "react-router-dom";
import { useMemo } from "react";
import { useAlbums } from "../features/albums/api";
import { useAssetFilters, useGallery } from "../features/assets/api";
import { flattenTags, useTags } from "../features/tags/api";
import { usePublicSettings, useVisit } from "../features/site/api";
import { PublicNav } from "../features/site/public-nav";
import { ThemeGallery } from "../features/gallery/theme-gallery";

function csv(search: URLSearchParams, key: string) {
  return search.get(key)?.split(",").filter(Boolean) ?? [];
}

export function GalleryPage() {
  useVisit("gallery");
  const { album } = useParams();
  const [search, setSearch] = useSearchParams();
  const page = Math.max(1, Number(search.get("page") || 1));
  const cameras = csv(search, "cameras");
  const lenses = csv(search, "lenses");
  const selectedTags = csv(search, "tags");
  const tagsOperator = search.get("tags_operator") === "or" ? "or" : "and";
  const sort = search.get("sort") as "asc" | "desc" | null;
  const preferredStyle = search.get("style") === "1" ? "single" : search.get("style") === "2" ? "waterfall" : undefined;
  const filters = useAssetFilters();
  const tags = useTags();
  const albums = useAlbums();
  const settings = usePublicSettings();
  const pageSize = 96;
  const resolvedAlbum = useMemo(() => {
    if (!album) return undefined;
    const withSlash = album.startsWith("/") ? album : `/${album}`;
    return (
      albums.data?.items.find((item) => item.album_value === album)?.album_value ??
      albums.data?.items.find((item) => item.album_value === withSlash)?.album_value ??
      album
    );
  }, [album, albums.data?.items]);
  const gallery = useGallery({
    page,
    pageSize,
    cameras,
    lenses,
    tags: selectedTags,
    tagsOperator,
    album: resolvedAlbum,
    sortByShootTime: sort ?? undefined
  });

  const previewSearch = useMemo(() => {
    const next = new URLSearchParams();
    if (resolvedAlbum) next.set("album", resolvedAlbum);
    if (cameras.length) next.set("cameras", cameras.join(","));
    if (lenses.length) next.set("lenses", lenses.join(","));
    if (selectedTags.length) {
      next.set("tags", selectedTags.join(","));
      next.set("tags_operator", tagsOperator);
    }
    if (sort) next.set("sort", sort);
    const value = next.toString();
    return value ? `?${value}` : "";
  }, [cameras, lenses, resolvedAlbum, selectedTags, sort, tagsOperator]);

  function update(key: string, value?: string) {
    const next = new URLSearchParams(search);
    if (value) next.set(key, value);
    else next.delete(key);
    next.set("page", "1");
    setSearch(next);
  }

  function setValues(key: string, values: string[]) {
    update(key, values.length ? values.join(",") : undefined);
  }

  function resetFilters() {
    const next = new URLSearchParams(search);
    next.delete("cameras");
    next.delete("lenses");
    next.delete("tags");
    next.delete("tags_operator");
    next.delete("sort");
    next.set("page", "1");
    setSearch(next);
  }

  return (
    <>
      <PublicNav />
      <ThemeGallery
        page={page}
        pageSize={pageSize}
        data={gallery.data}
        loading={gallery.isPending}
        error={gallery.error}
        album={resolvedAlbum}
        preferredStyle={preferredStyle}
        systemStyle={settings.data?.gallery_layout === "single" ? "single" : "waterfall"}
        previewSearch={previewSearch}
        cameras={filters.data?.cameras ?? []}
        lenses={filters.data?.lenses ?? []}
        tags={flattenTags(tags.data?.items ?? []).map((item) => ({
          value: item.id,
          label: `${"—".repeat(item.depth)} ${item.name}`
        }))}
        selectedCameras={cameras}
        selectedLenses={lenses}
        selectedTags={selectedTags}
        tagsOperator={tagsOperator}
        sort={sort ?? undefined}
        onPageChange={(value) => {
          const next = new URLSearchParams(search);
          next.set("page", String(value));
          setSearch(next);
        }}
        onCamerasChange={(values) => setValues("cameras", values)}
        onLensesChange={(values) => setValues("lenses", values)}
        onTagsChange={(values) => setValues("tags", values)}
        onTagsOperatorChange={(value) => update("tags_operator", value)}
        onSortChange={(value) => update("sort", value)}
        onReset={resetFilters}
      />
    </>
  );
}
