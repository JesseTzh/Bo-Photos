import { motion, AnimatePresence, useReducedMotion } from "framer-motion";
import { ArrowLeft, ArrowRight } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useAlbums } from "../features/albums/api";
import { albumPublicHref } from "../features/albums/routes";
import { useGallery } from "../features/assets/api";
import { PublicNav } from "../features/site/public-nav";
import { useVisit } from "../features/site/api";
import { AppLink } from "../shared/adapters/link";
import { useAppRouter } from "../shared/adapters/navigation";

export function HomePage() {
  useVisit("home");
  const router = useAppRouter();
  const reduceMotion = useReducedMotion();
  const gallery = useGallery({ page: 1, pageSize: 8, featured: true });
  const albums = useAlbums();
  const heroImages = useMemo(
    () => gallery.data?.items.filter((item) => item.preview_url || item.thumbnail_url).slice(0, 5) ?? [],
    [gallery.data?.items]
  );
  const [currentIndex, setCurrentIndex] = useState(0);
  const currentImage = heroImages[currentIndex];

  useEffect(() => {
    if (heroImages.length < 2) return;
    const timer = window.setInterval(() => {
      setCurrentIndex((value) => (value + 1) % heroImages.length);
    }, 5000);
    return () => window.clearInterval(timer);
  }, [heroImages.length]);

  function nextHero() {
    if (!heroImages.length) return;
    setCurrentIndex((value) => (value + 1) % heroImages.length);
  }

  function previousHero() {
    if (!heroImages.length) return;
    setCurrentIndex((value) => (value - 1 + heroImages.length) % heroImages.length);
  }

  return (
    <main className="min-h-screen bg-background">
      <PublicNav />

      <section className="relative h-[100dvh] min-h-[480px] w-full overflow-hidden">
        <div className="absolute inset-0 z-0">
          <AnimatePresence mode="sync">
            {currentImage ? (
              <motion.div
                key={currentImage.id}
                initial={reduceMotion ? false : { opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: reduceMotion ? 0 : 1.8, ease: [0.25, 0.1, 0.25, 1] }}
                className="absolute inset-0"
              >
                <div
                  className="absolute inset-0 bg-cover bg-center"
                  style={{ backgroundImage: `url(${currentImage.preview_url || currentImage.thumbnail_url})` }}
                />
                <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/30 to-black/20" />
                <div className="home-hero-vignette absolute inset-0" />
              </motion.div>
            ) : (
              <div className="home-hero-fallback absolute inset-0" />
            )}
          </AnimatePresence>
        </div>

        <div
          className="relative z-10 flex h-full flex-col px-5 sm:px-8 md:px-12 lg:px-16"
          style={{
            paddingTop: "max(calc(56px + env(safe-area-inset-top)), 68px)",
            paddingBottom: "max(env(safe-area-inset-bottom), 20px)"
          }}
        >
          <div className="flex shrink-0 items-center py-1">
            <motion.span
              initial={reduceMotion ? false : { opacity: 0, x: -16 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: reduceMotion ? 0 : 0.2, duration: reduceMotion ? 0 : 0.8 }}
              className="text-xs font-light uppercase tracking-[0.25em] text-white/80"
            >
              Photography
            </motion.span>
          </div>

          <div className="flex min-h-0 flex-1 flex-col items-center justify-center text-center">
            <motion.div
              initial={reduceMotion ? false : { opacity: 0, y: 28 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: reduceMotion ? 0 : 0.5, duration: reduceMotion ? 0 : 1, ease: "easeOut" }}
            >
              <p className="mb-2 text-[10px] uppercase tracking-[0.35em] text-white/50 sm:mb-3 sm:text-xs">
                Visual Storytelling
              </p>
              <h1 className="mb-3 text-4xl font-light leading-[1.1] text-white sm:mb-4 sm:text-5xl md:text-6xl lg:text-7xl">
                <span className="block">Every Moment</span>
                <span className="home-hero-accent-text block bg-clip-text text-transparent">
                  Tells a Story
                </span>
              </h1>
              <p className="mx-auto max-w-[260px] text-xs font-light leading-relaxed text-white/60 sm:max-w-sm sm:text-sm">
                捕捉光影，定格永恒 - 用镜头记录生活的美好瞬间
              </p>
            </motion.div>
          </div>

          <div className="mb-6 grid shrink-0 grid-cols-3 items-center gap-2 py-1 sm:mb-10 md:mb-14">
            <motion.div
              initial={reduceMotion ? false : { opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: reduceMotion ? 0 : 0.8, duration: reduceMotion ? 0 : 0.7 }}
              className="flex items-center gap-2 sm:gap-3"
            >
              {heroImages.length > 0 ? (
                <>
                  <span className="select-none font-mono text-[10px] text-white/50 tabular-nums sm:text-xs">
                    {String(currentIndex + 1).padStart(2, "0")}
                    <span className="mx-1 text-white/25">/</span>
                    {String(heroImages.length).padStart(2, "0")}
                  </span>
                  <div className="relative hidden h-px max-w-[72px] flex-1 overflow-hidden rounded-full bg-white/15 sm:block">
                    <motion.div
                      className="home-hero-progress absolute inset-y-0 left-0"
                      animate={{ width: `${((currentIndex + 1) / heroImages.length) * 100}%` }}
                      transition={{ duration: 0.6, ease: "easeInOut" }}
                    />
                  </div>
                </>
              ) : null}
            </motion.div>

            <div className="flex justify-center">
              <motion.button
                type="button"
                onClick={() => router.push("/covers?from=hero")}
                initial={reduceMotion ? false : { opacity: 0, y: 16 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: reduceMotion ? 0 : 0.9, duration: reduceMotion ? 0 : 0.8 }}
                whileHover={reduceMotion ? {} : { scale: 1.04 }}
                whileTap={{ scale: 0.97 }}
                className="btn-press group flex items-center gap-1.5 whitespace-nowrap rounded-full border border-white/20 bg-white/10 px-5 py-2.5 text-xs font-medium tracking-wide !text-white backdrop-blur-md transition-all duration-300 hover:bg-white/20 sm:gap-2 sm:px-7 sm:py-3 sm:text-sm"
              >
                探索作品集
                <ArrowRight className="h-3.5 w-3.5 text-white transition-transform duration-200 group-hover:translate-x-0.5 sm:h-4 sm:w-4" />
              </motion.button>
            </div>

            <motion.div
              initial={reduceMotion ? false : { opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: reduceMotion ? 0 : 0.8, duration: reduceMotion ? 0 : 0.7 }}
              className="flex items-center justify-end gap-1 sm:gap-2"
            >
              {heroImages.length > 0 ? (
                <>
                  <button
                    type="button"
                    onClick={previousHero}
                    aria-label="Previous"
                    className="btn-press flex h-8 w-8 items-center justify-center rounded-full border border-white/15 bg-white/5 text-white/60 transition-all duration-200 hover:border-white/40 hover:bg-white/15 hover:text-white sm:h-9 sm:w-9"
                  >
                    <ArrowLeft className="h-3.5 w-3.5" />
                  </button>
                  <button
                    type="button"
                    onClick={nextHero}
                    aria-label="Next"
                    className="btn-press flex h-8 w-8 items-center justify-center rounded-full border border-white/15 bg-white/5 text-white/60 transition-all duration-200 hover:border-white/40 hover:bg-white/15 hover:text-white sm:h-9 sm:w-9"
                  >
                    <ArrowRight className="h-3.5 w-3.5" />
                  </button>
                </>
              ) : null}
            </motion.div>
          </div>
        </div>
      </section>

      <section className="bg-background py-24">
        <div className="container mx-auto max-w-7xl px-4">
          <motion.div
            initial={reduceMotion ? false : { opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: reduceMotion ? 0 : 0.6, ease: "easeOut" }}
            className="mb-16 text-center"
          >
            <h2 className="mb-4 text-3xl font-light sm:text-4xl md:text-5xl">
              <span className="text-foreground">作品集</span>
              <span className="ml-2 text-primary">精选</span>
            </h2>
            <p className="mx-auto max-w-2xl text-lg text-muted-foreground">
              探索不同主题的摄影作品，每一个相册都记录着独特的故事
            </p>
          </motion.div>

          <div className="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-3">
            {albums.data?.items.slice(0, 6).map((album, index) => (
              <motion.div
                key={album.id}
                initial={reduceMotion ? false : { opacity: 0, y: 30 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: reduceMotion ? 0 : 0.6, ease: "easeOut", delay: reduceMotion ? 0 : index * 0.1 }}
                whileHover={reduceMotion ? {} : { y: -8 }}
              >
                <AppLink href={albumPublicHref(album.album_value, "?style=1")} className="group block">
                  <div className="relative mb-4 aspect-[4/3] overflow-hidden rounded-2xl bg-muted">
                    {album.cover_url ? (
                      <img
                        src={album.cover_url}
                        alt={album.name}
                        className="h-full w-full object-cover transition-transform duration-700 ease-out group-hover:scale-110"
                        loading="lazy"
                      />
                    ) : (
                      <div className="album-cover-fallback h-full w-full" />
                    )}
                    <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-black/10 to-transparent opacity-60 transition-opacity duration-300 group-hover:opacity-80" />
                    <div className="absolute inset-x-0 bottom-0 translate-y-4 p-6 transition-transform duration-500 group-hover:translate-y-0">
                      <p className="mb-2 text-xs uppercase tracking-widest text-white/80">相册</p>
                      <h3 className="mb-1 text-xl font-medium text-white">{album.name}</h3>
                      <p className="flex items-center gap-2 text-sm text-white/70">
                        <span className="inline-block h-2 w-2 rounded-full bg-primary" />
                        {album.random_show ? "随机展示" : "顺序排列"}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center justify-between">
                    <div className="min-w-0">
                      <h3 className="font-medium text-foreground transition-colors group-hover:text-primary">{album.name}</h3>
                      <p className="mt-1 truncate text-sm text-muted-foreground">{album.detail || "探索精彩瞬间"}</p>
                    </div>
                    <div className="flex items-center gap-1 text-primary opacity-0 transition-opacity group-hover:opacity-100">
                      <span className="text-sm">浏览</span>
                      <ArrowRight className="h-4 w-4" />
                    </div>
                  </div>
                </AppLink>
              </motion.div>
            ))}
          </div>

          {(albums.data?.items.length ?? 0) > 6 ? (
            <motion.div
              initial={reduceMotion ? false : { opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: reduceMotion ? 0 : 0.6, delay: reduceMotion ? 0 : 0.6 }}
              className="mt-16 text-center"
            >
              <AppLink
                href="/covers"
                className="btn-press group inline-flex items-center gap-2 rounded-full bg-foreground px-8 py-4 font-medium text-background transition-all duration-300 hover:bg-foreground/90"
              >
                查看全部相册
                <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
              </AppLink>
            </motion.div>
          ) : null}
        </div>
      </section>

    </main>
  );
}
