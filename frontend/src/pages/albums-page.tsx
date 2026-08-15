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

export function AlbumsPage({ coversOnly = false }: { coversOnly?: boolean }) {
  useVisit("album");
  const albums = useAlbums();
  const items = albums.data?.items ?? [];

  if (coversOnly) {
    return (
      <div className="min-h-screen bg-background">
        <PublicNav />
        <div className="container mx-auto mb-12 px-4 pt-20">
          <AppLink href="/albums" className="mb-6 inline-flex items-center gap-2 rounded-full px-0 py-2 text-sm text-muted-foreground transition-colors hover:text-foreground">
              <ArrowLeft className="h-4 w-4" />
              返回作品合集
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

  return (
    <div className="min-h-screen bg-background">
      <PublicNav />
      <main className="pt-14">
        <section className="px-4 py-16 sm:px-6 lg:px-8">
          <div className="mx-auto flex max-w-6xl flex-wrap items-end justify-between gap-8 sm:flex-nowrap">
            <div>
              <h1 className="text-3xl font-light leading-none tracking-tight text-foreground sm:text-4xl">景行集</h1>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground/80">跨相册浏览照片，使用筛选和排序查找你想看的作品。</p>
            </div>
            <AppLink
              href="/gallery"
              className="btn-press inline-flex items-center rounded-full border border-border bg-card px-5 py-2 text-sm text-foreground transition-colors hover:bg-muted"
            >
              进入图库
            </AppLink>
          </div>
        </section>

        <div className="h-px w-full bg-border/50" />

        <section className="px-4 py-12 sm:px-6 sm:py-16 lg:px-8">
          {!items.length && !albums.isPending ? <Empty description="暂无公开相册" /> : null}
          <div className="mx-auto grid max-w-6xl grid-cols-1 gap-10 sm:gap-12 md:grid-cols-2 lg:gap-14">
            {items.map((album) => (
              <AppLink key={album.id} href={albumPublicHref(album.album_value)} className="group block">
                <article className="flex flex-col">
                  <div className="relative aspect-[4/3] overflow-hidden rounded-xl border border-border/60 bg-muted">
                    {album.cover_url ? (
                      <img
                        src={album.cover_url}
                        alt={album.name}
                        loading="lazy"
                        decoding="async"
                        className="h-full w-full object-cover transition-transform duration-[1100ms] ease-out group-hover:scale-[1.03]"
                      />
                    ) : (
                      <div className="absolute inset-0 bg-gradient-to-br from-muted to-muted/60" />
                    )}
                    <div className="absolute inset-0 bg-gradient-to-t from-media-scrim/50 via-media-scrim/0 to-media-scrim/0" />
                    <div className="absolute left-3 top-3 rounded-full border border-media-control/10 bg-media-scrim/50 px-2.5 py-1 text-[11px] font-medium text-on-media/90 backdrop-blur-md">
                      {album.asset_count} 张
                    </div>
                  </div>
                  <header className="mt-4">
                    <h2 className="text-xl font-light leading-snug tracking-tight text-foreground transition-colors duration-300 group-hover:text-foreground/60 sm:text-2xl">
                      {album.name}
                    </h2>
                    <p className="mt-1.5 text-xs tracking-wide text-muted-foreground">{album.detail || "探索精彩瞬间"}</p>
                  </header>
                </article>
              </AppLink>
            ))}
          </div>
        </section>
      </main>
    </div>
  );
}
