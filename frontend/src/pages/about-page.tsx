import { motion, AnimatePresence } from "framer-motion";
import { Book, Camera, Code, MessagesSquare } from "lucide-react";
import { useEffect, useMemo, useState, type ReactNode } from "react";
import { Button, Empty } from "antd";
import { PublicNav } from "../features/site/public-nav";
import { usePublicSettings, useVisit } from "../features/site/api";

export function AboutPage() {
  useVisit("about");
  const settings = usePublicSettings().data;
  const [index, setIndex] = useState(0);
  const socialLinks = [
    { label: "INS", href: settings?.about_social_instagram, icon: <Camera className="h-4 w-4" /> },
    { label: "小红书", href: settings?.about_social_xiaohongshu, icon: <Book className="h-4 w-4" /> },
    { label: "微博", href: settings?.about_social_weibo, icon: <MessagesSquare className="h-4 w-4" /> },
    { label: "GitHub", href: settings?.about_social_github, icon: <Code className="h-4 w-4" /> }
  ].flatMap((item) => (item.href ? [{ label: item.label, href: item.href, icon: item.icon as ReactNode }] : []));
  const galleryIds = settings?.about_gallery_asset_ids ?? [];
  const images = useMemo(() => galleryIds.map((id) => `/media/assets/${id}/preview`), [galleryIds]);

  useEffect(() => {
    if (images.length < 2) return;
    const timer = window.setInterval(() => setIndex((value) => (value + 1) % images.length), 6000);
    return () => window.clearInterval(timer);
  }, [images.length]);

  return (
    <>
      <PublicNav />
      <main className="flex min-h-screen items-center justify-center bg-background text-foreground">
        <div className="mx-auto w-full max-w-6xl px-4 py-24 md:px-6 lg:px-8">
          <div className="flex flex-col items-center gap-10 lg:flex-row lg:gap-16">
            <section className="order-2 flex w-full flex-col justify-center space-y-6 lg:order-1 lg:w-[35%]">
              <header className="space-y-4">
                <div className="inline-block">
                  <h1 className="text-2xl font-bold tracking-[0.08em] text-foreground md:text-3xl lg:text-4xl">ABOUT ME</h1>
                  <div className="mt-2 h-0.5 w-12 bg-gradient-to-r from-foreground/80 to-transparent" />
                </div>
                <p className="max-w-md whitespace-pre-line text-sm leading-relaxed text-muted-foreground md:text-base lg:text-lg">
                  {settings?.about_intro || "尚未填写介绍。"}
                </p>
              </header>

              {socialLinks.length > 0 ? (
                <nav className="space-y-3">
                  <h2 className="text-xs font-medium uppercase tracking-[0.2em] text-foreground/80">Follow Me</h2>
                  <div className="flex flex-wrap items-center gap-3">
                    {socialLinks.map((link) => (
                      <Button key={link.label} href={link.href} target="_blank" rel="noreferrer" className="rounded-full">
                        <span className="opacity-80">{link.icon}</span>
                        <span>{link.label}</span>
                      </Button>
                    ))}
                  </div>
                </nav>
              ) : null}

              <div className="hidden items-center gap-3 pt-2 lg:flex">
                <div className="h-px flex-1 bg-gradient-to-r from-foreground/20 to-transparent" />
                <span className="text-[10px] uppercase tracking-widest text-foreground/60">Photography</span>
              </div>
            </section>

            <section className="order-1 w-full lg:order-2 lg:w-[60%] lg:pl-4">
              {images.length > 0 ? (
                <div className="relative aspect-video w-full overflow-hidden rounded-2xl border border-border bg-muted shadow-lg">
                  <AnimatePresence mode="wait">
                    <motion.img
                      key={images[index]}
                      src={images[index]}
                      alt=""
                      initial={{ opacity: 0, scale: 1.02 }}
                      animate={{ opacity: 1, scale: 1 }}
                      exit={{ opacity: 0, scale: 0.98 }}
                      transition={{ duration: 0.5 }}
                      className="absolute inset-0 h-full w-full object-cover"
                    />
                  </AnimatePresence>
                  {images.length > 1 ? (
                    <div className="absolute bottom-4 left-1/2 flex -translate-x-1/2 gap-2">
                      {images.map((image, imageIndex) => (
                        <button
                          key={image}
                          type="button"
                          onClick={() => setIndex(imageIndex)}
                          aria-label={`第 ${imageIndex + 1} 张`}
                          className={`h-1.5 rounded-full transition-all ${imageIndex === index ? "w-6 bg-white" : "w-1.5 bg-white/50"}`}
                        />
                      ))}
                    </div>
                  ) : null}
                </div>
              ) : (
                <div className="flex aspect-video w-full items-center justify-center rounded-2xl border border-border bg-muted">
                  <Empty description="尚未配置个人照片" />
                </div>
              )}
            </section>
          </div>
        </div>
      </main>
    </>
  );
}
