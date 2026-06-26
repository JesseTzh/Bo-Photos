import { CalendarOutlined, EnvironmentOutlined } from "@ant-design/icons";
import { Alert, Empty } from "antd";
import { motion } from "framer-motion";
import { useGuides } from "../features/guides/api";
import { PublicNav } from "../features/site/public-nav";
import { useVisit } from "../features/site/api";
import { AppLink } from "../shared/adapters/link";

function formatRange(start?: string, end?: string): string {
  const startDate = start ? new Date(start) : null;
  const endDate = end ? new Date(end) : null;
  if (!startDate && !endDate) return "";
  const format = (date: Date) =>
    `${date.getFullYear()}.${String(date.getMonth() + 1).padStart(2, "0")}.${String(date.getDate()).padStart(2, "0")}`;
  if (startDate && endDate) return `${format(startDate)} - ${format(endDate)}`;
  return format((startDate ?? endDate)!);
}

export function GuidesPage() {
  useVisit("guide");
  const query = useGuides();
  const guides = query.data?.items ?? [];
  const countries = new Set(guides.map((guide) => guide.country).filter(Boolean)).size;
  const days = guides.reduce((sum, guide) => sum + (Number(guide.days) || 0), 0);
  return (
    <>
      <PublicNav />
      <main className="min-h-screen bg-background pt-14 text-foreground">
        <section className="px-4 pb-8 pt-6 sm:px-6 sm:pb-10 sm:pt-8 lg:px-8">
          <div className="mx-auto flex max-w-6xl flex-wrap items-end justify-between gap-8 sm:flex-nowrap">
            <div className="text-left">
              <h1 className="text-3xl font-light leading-none tracking-tight text-foreground sm:text-4xl">攻略路书</h1>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground/80">探索我去过的地方，发现旅行故事与路书。</p>
            </div>
            <dl className="grid shrink-0 grid-cols-3 gap-4 sm:gap-8">
              <StatCell label="路书" value={guides.length} />
              <StatCell label="国家" value={countries} />
              <StatCell label="总天数" value={days} />
            </dl>
          </div>
        </section>

        <div className="h-px w-full bg-border/50" />

        <section className="px-4 py-12 sm:px-6 sm:py-16 lg:px-8">
          <div className="mx-auto max-w-6xl">
            {query.error ? <Alert type="error" message="Guide 加载失败" showIcon /> : null}
            {!query.isPending && !guides.length ? <Empty description="暂无公开 Guide" /> : null}
            <ul className="grid grid-cols-1 gap-10 sm:gap-12 md:grid-cols-2 lg:gap-14">
              {guides.map((guide, index) => {
                const range = formatRange(guide.start_date, guide.end_date);
                return (
                  <motion.li
                    key={guide.id}
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true, amount: 0.2 }}
                    transition={{ duration: 0.7, ease: [0.22, 0.61, 0.36, 1], delay: index * 0.05 }}
                  >
                    <AppLink href={`/guides/${guide.id}`} className="group block">
                      <article className="flex flex-col">
                        <div className="relative aspect-[4/3] overflow-hidden rounded-xl border border-border/60 bg-muted">
                          {guide.cover_url ? (
                            <img
                              src={guide.cover_url}
                              alt={guide.title}
                              loading="lazy"
                              decoding="async"
                              className="h-full w-full object-cover transition-transform duration-[1100ms] ease-out group-hover:scale-[1.03]"
                            />
                          ) : (
                            <div className="absolute inset-0 bg-gradient-to-br from-muted to-muted/60" />
                          )}
                          <div className="absolute inset-0 bg-gradient-to-t from-black/50 via-black/0 to-black/0" />
                          {guide.days > 0 ? (
                            <div className="absolute left-3 top-3 flex items-center gap-1.5 rounded-full border border-white/10 bg-black/50 px-2.5 py-1 text-[11px] font-medium text-white/90 backdrop-blur-md">
                              <CalendarOutlined aria-hidden />
                              <span className="tabular-nums">{guide.days} 天</span>
                            </div>
                          ) : null}
                        </div>

                        <header className="mt-4">
                          <div className="flex items-center gap-2 text-xs text-muted-foreground">
                            <EnvironmentOutlined aria-hidden />
                            <span className="font-medium text-foreground">{guide.country}</span>
                            <span aria-hidden>·</span>
                            <span className="truncate">{guide.city}</span>
                          </div>
                          <h2 className="mt-2 text-xl font-light leading-snug tracking-tight text-foreground transition-colors duration-300 group-hover:text-foreground/60 sm:text-2xl">
                            {guide.title}
                          </h2>
                          {range ? <p className="mt-1.5 text-xs tracking-wide text-muted-foreground tabular-nums">{range}</p> : null}
                        </header>
                      </article>
                    </AppLink>
                  </motion.li>
                );
              })}
            </ul>
          </div>
        </section>
      </main>
    </>
  );
}

function StatCell({ label, value }: { label: string; value: number }) {
  return (
    <div className="text-right">
      <dt className="text-[11px] uppercase tracking-[0.18em] text-muted-foreground/70">{label}</dt>
      <dd className="mt-1 text-xl font-light text-foreground tabular-nums">{value}</dd>
    </div>
  );
}
