import { useSearchParams } from "react-router-dom";
import { usePrivateAssets } from "../features/assets/admin-api";
import { PublicNav } from "../features/site/public-nav";
import { ThemeGallery } from "../features/gallery/theme-gallery";

export function PrivateAlbumPage() {
  const [search, setSearch] = useSearchParams();
  const page = Math.max(1, Number(search.get("page") || 1));
  const pageSize = 96;
  const assets = usePrivateAssets({ page, pageSize });

  return (
    <>
      <PublicNav />
      <ThemeGallery
        page={page}
        pageSize={pageSize}
        data={assets.data}
        loading={assets.isPending}
        error={assets.error}
        album="private"
        preferredStyle="waterfall"
        systemStyle="waterfall"
        previewSearch="?private=1"
        cameras={[]}
        lenses={[]}
        tags={[]}
        selectedCameras={[]}
        selectedLenses={[]}
        selectedTags={[]}
        tagsOperator="and"
        onPageChange={(value) => {
          const next = new URLSearchParams(search);
          next.set("page", String(value));
          setSearch(next);
        }}
        onCamerasChange={() => undefined}
        onLensesChange={() => undefined}
        onTagsChange={() => undefined}
        onTagsOperatorChange={() => undefined}
        onSortChange={() => undefined}
        onReset={() => undefined}
      />
    </>
  );
}
