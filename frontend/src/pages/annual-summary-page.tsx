import { ArrowLeft, ArrowRight } from "lucide-react";
import { useMemo, useState } from "react";
import { AssetMedia } from "../features/assets/asset-media";
import { useHomeAssets, useAnnualSummary, useVisit, type AnnualSummarySlot } from "../features/site/api";
import { PublicNav } from "../features/site/public-nav";
import { AppLink } from "../shared/adapters/link";

export function AnnualSummaryPage() {
  useVisit("annual-summary");
  const currentYear = new Date().getFullYear();
  const [year, setYear] = useState(currentYear);
  const summary = useAnnualSummary(year);
  const assetIds = useMemo(() => (summary.data?.slots ?? []).map((slot) => slot.asset_id).filter((id): id is string => Boolean(id)), [summary.data?.slots]);
  const assets = useHomeAssets(assetIds);
  const byId = useMemo(() => new Map((assets.data ?? []).map((asset) => [asset.id, asset])), [assets.data]);
  const years = Array.from(new Set([...(summary.data?.years ?? []), currentYear])).sort((a, b) => a - b);
  const slots: AnnualSummarySlot[] = summary.data?.slots ?? Array.from({ length: 10 }, (_, slot) => ({ slot, comment: "" }));

  return <main className="min-h-screen bg-background text-foreground"><PublicNav /><div className="mx-auto flex max-w-7xl gap-8 px-5 pb-20 pt-28 md:gap-16 md:px-10">
    <aside className="annual-years-panel"><p className="mb-5 text-xs uppercase tracking-[0.24em] text-muted-foreground">年度精选</p><div className="annual-years-scroll">
      {years.map((item) => <button type="button" key={item} className={`annual-year${item === year ? " is-active" : ""}`} onClick={() => setYear(item)}>{item}</button>)}
    </div><AppLink href="/" className="mt-8 inline-flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"><ArrowLeft className="h-4 w-4" />返回首页</AppLink></aside>
    <section className="min-w-0 flex-1"><div className="mb-10 flex items-end justify-between border-b border-border/60 pb-5"><div><p className="mb-2 text-sm text-primary">YEAR IN REVIEW</p><h1 className="font-serif text-4xl font-normal md:text-6xl">{year}</h1></div><span className="text-sm text-muted-foreground">{slots.filter((slot) => slot.asset_id).length} / 10</span></div>
      <div className="annual-summary-grid">{slots.map((slot, index) => { const asset = slot.asset_id ? byId.get(slot.asset_id) : undefined; return <article key={slot.slot} className="annual-summary-item"><div className="annual-summary-media">{asset ? <AssetMedia asset={asset} className="h-full w-full object-cover" /> : <div className="annual-summary-empty"><span>{String(index + 1).padStart(2, "0")}</span></div>}</div><div className="mt-3 flex gap-3"><span className="font-mono text-xs text-primary">{String(index + 1).padStart(2, "0")}</span><p className="m-0 text-sm leading-relaxed text-muted-foreground">{slot.comment || "这一刻，值得被记住。"}</p></div></article>; })}</div>
      {summary.isLoading ? <div className="py-16 text-center text-sm text-muted-foreground">正在整理这一年的故事...</div> : null}
      <div className="mt-12 flex justify-between border-t border-border/60 pt-5"><button type="button" className="annual-nav-button" onClick={() => setYear((value) => Math.max(1900, value - 1))}>上一年<ArrowLeft className="h-4 w-4" /></button><button type="button" className="annual-nav-button" onClick={() => setYear((value) => value + 1)}><ArrowRight className="h-4 w-4" />下一年</button></div>
    </section></div></main>;
}
