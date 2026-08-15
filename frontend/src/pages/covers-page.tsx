import { Empty } from "antd";
import { ArrowLeft } from "lucide-react";
import type { Album } from "../features/albums/api";
import { useAlbums } from "../features/albums/api";
import { albumPublicHref } from "../features/albums/routes";
import { useVisit } from "../features/site/api";
import { PublicNav } from "../features/site/public-nav";
import { AppLink } from "../shared/adapters/link";

function DestinationCard({ album }: { album: Album }) {
  return (
    <div className="group h-full w-full">
      <AppLink
        href={albumPublicHref(album.album_value, "?style=1")}
        className="album-destination-card relative block h-full w-full overflow-hidden transition-all duration-500 ease-in-out group-hover:scale-105"
        aria-label={`Explore details for ${album.name}`}
      >
        {album.cover_url ? (
          <img
            src={album.cover_url}
            alt={album.name}
            loading="lazy"
            decoding="async"
            className="absolute inset-0 h-full w-full object-cover transition-transform duration-500 ease-in-out group-hover:scale-110"
          />
        ) : (
          <div className="absolute inset-0 bg-muted" />
        )}
        <div className="relative flex h-full flex-col items-center justify-center p-6 text-center text-on-media">
          <h2 className="media-title-shadow text-4xl font-bold uppercase tracking-[0.2em]">{album.name}</h2>
          <p className="mt-3 translate-y-4 text-sm font-medium uppercase tracking-widest text-on-media/90 opacity-0 transition-all duration-500 group-hover:translate-y-0 group-hover:opacity-100">
            {album.asset_count} PHOTOS
          </p>
        </div>
      </AppLink>
    </div>
  );
}

export function CoversPage() {
  useVisit("album");
  const albums = useAlbums();
  const items = albums.data?.items ?? [];

  return (
    <div className="min-h-screen bg-background">
      <PublicNav />
      <div className="container mx-auto mb-12 px-4 pt-20">
        <AppLink href="/gallery" className="mb-6 inline-flex items-center gap-2 rounded-full px-0 py-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
          <ArrowLeft className="h-4 w-4" />
          返回图库
        </AppLink>

        {!items.length && !albums.isPending ? <Empty description="暂无相册封面" /> : null}

        <div className="mx-auto grid max-w-6xl grid-cols-1 gap-8 md:grid-cols-2 md:gap-12">
          {items.filter((album) => album.cover_url).map((album) => (
            <div key={album.id} className="aspect-[4/3] w-full">
              <DestinationCard album={album} />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
