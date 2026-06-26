import { Alert, Checkbox } from "antd";
import { motion, useScroll, useTransform } from "framer-motion";
import { ArrowLeft, Calendar, ImageIcon, MapPin } from "lucide-react";
import { useEffect, useRef, useState, type MutableRefObject } from "react";
import { useParams } from "react-router-dom";
import { albumPublicHref } from "../features/albums/routes";
import { useGuide, type GuideBlock, type GuideModule } from "../features/guides/api";
import { useVisit } from "../features/site/api";
import { AppLink } from "../shared/adapters/link";
import { cn } from "../shared/lib/utils";

const moduleIcons: Record<string, string> = {
  expense: "💸",
  checklist: "📋",
  itinerary: "🗓️",
  transport: "🚗",
  photo: "📷",
  tips: "💡",
  text: "📝",
  markdown: "📝"
};

const moduleToneClasses: Record<string, string> = {
  expense: "guide-tone-warm",
  checklist: "guide-tone-cool",
  itinerary: "guide-tone-warm",
  transport: "guide-tone-cool",
  photo: "guide-tone-warm",
  tips: "guide-tone-cool",
  text: "guide-tone-neutral",
  markdown: "guide-tone-neutral"
};

export function GuideDetailPage() {
  useVisit("guide");
  const { id } = useParams();
  const query = useGuide(id);
  const [activeModuleId, setActiveModuleId] = useState<string | null>(null);
  const [showMobileNav, setShowMobileNav] = useState(false);
  const moduleRefs = useRef<Record<string, HTMLElement | null>>({});
  const { scrollY } = useScroll();
  const coverY = useTransform(scrollY, [0, 500], [0, 150]);
  const coverOpacity = useTransform(scrollY, [0, 400], [1, 0.3]);

  useEffect(() => {
    if (!query.data?.modules.length) return;
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) setActiveModuleId(entry.target.id.replace("module-", ""));
        });
      },
      { rootMargin: "-100px 0px -70% 0px", threshold: 0 }
    );
    query.data.modules.forEach((module) => {
      const element = document.getElementById(`module-${module.id}`);
      if (element) observer.observe(element);
    });
    return () => observer.disconnect();
  }, [query.data?.modules]);

  if (query.error) return <Alert type="error" message="Guide 不存在" />;
  if (!query.data) return <main className="page-loading">加载中...</main>;

  const { guide, modules, albums } = query.data;

  const handleModuleClick = (moduleId: string) => {
    setActiveModuleId(moduleId);
    setShowMobileNav(false);
    const element = moduleRefs.current[moduleId];
    if (!element) return;
    const offsetPosition = element.getBoundingClientRect().top + window.pageYOffset - 100;
    window.scrollTo({ top: offsetPosition, behavior: "smooth" });
  };

  return (
    <main className="guide-detail-shell min-h-screen pb-24 lg:pb-0">
      <div className="relative h-[70vh] w-full overflow-hidden sm:h-[80vh]">
        <motion.div style={{ y: coverY, opacity: coverOpacity }} className="absolute inset-0">
          {guide.cover_url ? (
            <img src={guide.cover_url} alt={guide.title} className="h-full w-full object-cover" />
          ) : (
            <div className="guide-cover-fallback absolute inset-0" />
          )}
        </motion.div>
        <div className="guide-cover-bottom-fade absolute inset-0" />
        <div className="absolute inset-0 bg-gradient-to-b from-black/30 via-transparent to-transparent" />

        <AppLink
          href="/guides"
          className="absolute left-4 top-4 z-10 inline-flex items-center gap-2 rounded-full border border-border/70 bg-card/90 px-4 py-2 text-sm font-medium text-card-foreground shadow-lg backdrop-blur-xl transition-colors hover:bg-card sm:left-8 sm:top-8"
        >
          <ArrowLeft className="h-4 w-4" />
          <span className="hidden sm:inline">返回列表</span>
        </AppLink>

        <div className="absolute bottom-0 left-0 right-0 p-6 sm:p-12 lg:p-16">
          <div className="max-w-4xl">
            <motion.p initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="mb-3 text-xs font-semibold uppercase tracking-[0.2em] text-primary sm:text-sm">
              Travel Guide
            </motion.p>
            <motion.h1 initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} className="mb-4 text-2xl font-bold leading-tight tracking-tight text-foreground sm:text-4xl lg:text-6xl">
              {guide.title}
            </motion.h1>
            <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-muted-foreground">
              <span className="inline-flex items-center gap-1.5"><MapPin className="h-4 w-4 text-primary" />{guide.country} · {guide.city}</span>
              <span className="hidden text-border sm:inline">|</span>
              <span className="inline-flex items-center gap-1.5"><Calendar className="h-4 w-4 text-primary" />{guide.days} 天</span>
            </motion.div>
          </div>
        </div>
      </div>

      <div className="mx-auto w-full max-w-7xl px-3 py-8 sm:px-6 sm:py-12 lg:px-8 lg:py-16">
        <div className="flex justify-center gap-8">
          {modules.length > 0 ? (
            <aside className="hidden flex-shrink-0 lg:block">
              <Toc modules={modules} activeModuleId={activeModuleId} onModuleClick={handleModuleClick} />
            </aside>
          ) : null}

          <section className="min-w-0 flex-1 lg:max-w-[66.666667%]">
            {albums?.length > 0 ? (
              <motion.section initial={{ opacity: 0, y: 20 }} whileInView={{ opacity: 1, y: 0 }} viewport={{ once: true }} className="mb-8 sm:mb-14">
                <h2 className="mb-4 flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground sm:mb-6 sm:text-sm">
                  <ImageIcon className="h-4 w-4" />
                  关联相册
                </h2>
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4 lg:grid-cols-3">
                  {albums.map((album) => (
                    <AppLink key={album.id} href={albumPublicHref(album.value)} className="group block">
                      <div className="overflow-hidden rounded-xl border border-border/70 bg-card/70 backdrop-blur-xl transition-all duration-300 hover:-translate-y-0.5 hover:border-primary/60 hover:shadow-lg">
                        <div className="relative aspect-[3/2] overflow-hidden">
                          {album.coverURL ? <img src={album.coverURL} alt={album.name} className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105" /> : <div className="absolute inset-0 bg-muted" />}
                          <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent" />
                          <div className="absolute bottom-0 left-0 right-0 p-3 sm:p-4">
                            <p className="truncate text-sm font-semibold text-white sm:text-base">{album.name}</p>
                            <p className="mt-0.5 text-xs text-white/70">点击查看相册</p>
                          </div>
                        </div>
                      </div>
                    </AppLink>
                  ))}
                </div>
              </motion.section>
            ) : null}

            <div className="space-y-3 sm:space-y-8">
              {modules.map((module) => (
                <ModuleView key={module.id} module={module} moduleRefs={moduleRefs} />
              ))}
            </div>
          </section>
        </div>
      </div>

      {modules.length > 0 ? (
        <div className="fixed bottom-0 left-0 right-0 z-50 border-t border-border/70 bg-background/90 shadow-lg backdrop-blur-xl lg:hidden">
          <div className="flex items-center justify-between px-3 py-2.5">
            <button onClick={() => setShowMobileNav((value) => !value)} className="flex items-center gap-2 text-sm font-medium text-foreground transition-transform active:scale-95" aria-label="打开目录导航" type="button">
              <span className={cn("text-lg transition-transform", showMobileNav && "rotate-180")}>☰</span>
              <span>目录</span>
            </button>
            {activeModuleId ? <span className="max-w-[50%] truncate text-xs text-muted-foreground">{modules.find((module) => module.id === activeModuleId)?.name}</span> : null}
          </div>
          {showMobileNav ? (
            <div className="max-h-[60vh] overflow-y-auto overscroll-contain border-t border-border/70">
              <nav className="space-y-1 p-2">
                {modules.map((module) => <TocButton key={module.id} module={module} active={activeModuleId === module.id} onClick={() => handleModuleClick(module.id)} />)}
              </nav>
            </div>
          ) : null}
        </div>
      ) : null}
    </main>
  );
}

function Toc({ modules, activeModuleId, onModuleClick }: { modules: GuideModule[]; activeModuleId: string | null; onModuleClick: (id: string) => void }) {
  return (
    <div className="sticky top-20 h-fit w-72 overflow-hidden rounded-2xl border border-border bg-card/80 shadow-sm backdrop-blur-xl" style={{ maxHeight: "calc(100vh - 120px)" }}>
      <div className="border-b border-border px-5 py-4">
        <span className="text-sm font-medium text-foreground">目录导航</span>
      </div>
      <nav className="max-h-[calc(100vh-200px)] overflow-y-auto p-3">
        <ul className="space-y-1.5">
          {modules.map((module) => (
            <li key={module.id}>
              <TocButton module={module} active={activeModuleId === module.id} onClick={() => onModuleClick(module.id)} />
            </li>
          ))}
        </ul>
      </nav>
    </div>
  );
}

function TocButton({ module, active, onClick }: { module: GuideModule; active: boolean; onClick: () => void }) {
  const toneClass = moduleToneClasses[module.template || ""] || moduleToneClasses.tips;
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "relative flex w-full items-center gap-3 overflow-hidden rounded-xl px-4 py-3 text-left transition-all duration-300",
        active ? `${toneClass} font-medium shadow-sm` : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
      )}
    >
      {active ? <span className="absolute left-0 top-1/2 h-8 w-1 -translate-y-1/2 rounded-r-full bg-current opacity-60" /> : null}
      <span className="flex-shrink-0 text-lg">{moduleIcons[module.template || ""] || "📄"}</span>
      <span className="flex-1 truncate text-sm">{module.name}</span>
    </button>
  );
}

function ModuleView({ module, moduleRefs }: { module: GuideModule; moduleRefs: MutableRefObject<Record<string, HTMLElement | null>> }) {
  const toneClass = moduleToneClasses[module.template || ""] || moduleToneClasses.tips;
  const icon = moduleIcons[module.template || ""] || "📄";
  return (
    <motion.section
      id={`module-${module.id}`}
      ref={(element) => {
        moduleRefs.current[module.id] = element;
      }}
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-80px" }}
      transition={{ duration: 0.5 }}
      className={cn("mb-3 overflow-hidden rounded-xl border-l-4 backdrop-blur-xl last:mb-0 sm:mb-8 sm:rounded-2xl", toneClass)}
    >
      <div className="p-3 sm:p-6 lg:p-8">
        <div className="mb-3 flex items-center gap-2 sm:mb-6 sm:gap-3">
          <span className="text-lg sm:text-2xl">{icon}</span>
          <h2 className="text-base font-bold tracking-tight text-foreground sm:text-xl lg:text-2xl">{module.name}</h2>
        </div>
        <div className="rounded-lg border border-border/50 bg-card/70 p-2.5 backdrop-blur-sm sm:rounded-xl sm:p-4 lg:p-6">
          {module.kind === "structured" ? <StructuredView template={module.template} data={module.structured_data} /> : module.blocks?.map((block) => <BlockView key={block.id} block={block} />)}
        </div>
      </div>
    </motion.section>
  );
}

function StructuredView({ template, data }: { template?: string; data: unknown }) {
  const items = Array.isArray(data) ? data : [];
  if (!items.length) return <p className="text-sm text-muted-foreground">暂无内容</p>;
  return (
    <div className="space-y-3">
      {items.map((item: any, index) => (
        <div key={item.id ?? index} className="rounded-xl border border-border/70 bg-card/80 p-4 backdrop-blur transition-all hover:shadow-md">
          <div className="mb-2 flex items-center justify-between gap-3">
            <h3 className="font-semibold text-foreground">{item.title ?? item.name ?? item.route ?? `项目 ${index + 1}`}</h3>
            {template === "expense" ? <b className="text-primary">¥{item.subtotal ?? item.amount ?? 0}</b> : null}
          </div>
          {template === "checklist" && Array.isArray(item.items) ? (
            <div className="space-y-1">
              {item.items.map((value: any) => <Checkbox key={value.id ?? value.text} checked={value.completed}>{value.text}</Checkbox>)}
            </div>
          ) : null}
          <p className="whitespace-pre-line text-sm leading-relaxed text-muted-foreground">{item.description ?? item.text ?? item.notes ?? item.location}</p>
          {item.date ? <small className="mt-2 block text-xs text-muted-foreground">{item.date} {item.time}</small> : null}
        </div>
      ))}
    </div>
  );
}

function BlockView({ block }: { block: GuideBlock }) {
  const data = block.data as any;
  switch (block.type) {
    case "markdown":
      return <p className="mb-3 whitespace-pre-line text-sm leading-relaxed text-foreground/90">{data.text}</p>;
    case "image":
      return (
        <figure className="mb-4 sm:mb-8">
          <img src={`/media/assets/${data.asset_id}/preview`} alt={data.caption || ""} className="h-auto w-full rounded-xl" />
          {data.caption ? <figcaption className="mt-1.5 text-center text-xs text-muted-foreground sm:mt-3">{data.caption}</figcaption> : null}
        </figure>
      );
    case "video":
      return <video controls src={data.url} className="w-full rounded-xl" />;
    case "link":
      return (
        <a href={data.url} target="_blank" rel="noreferrer" className="mb-3 block rounded-xl border border-border/70 bg-card/70 p-4 text-sm transition-colors hover:border-primary/60">
          <b>{data.title || "链接"}</b>
          <p className="mt-1 text-muted-foreground">{data.description}</p>
        </a>
      );
    case "tasks":
      return <div className="space-y-1">{data.items?.map((item: any) => <Checkbox key={item.id ?? item.text} checked={item.completed}>{item.text}</Checkbox>)}</div>;
    case "warning":
      return <Alert type={data.level === "warning" ? "warning" : "info"} message={data.title} description={data.text} showIcon className="mb-3" />;
    case "divider":
      return <hr className="my-4 border-border/70" />;
    default:
      return null;
  }
}
