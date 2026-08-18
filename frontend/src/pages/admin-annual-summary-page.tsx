import { EditOutlined, SaveOutlined } from "@ant-design/icons";
import { Button, Form, Input, InputNumber, Space, Typography, message } from "antd";
import { useEffect, useMemo, useState } from "react";
import { useAllAdminAssets } from "../features/assets/admin-api";
import { AssetMedia } from "../features/assets/asset-media";
import type { Asset } from "../features/assets/schema";
import { HomeAssetPicker } from "../features/site/home-asset-picker";
import { useAdminAnnualSummary, useSaveAnnualSummary, type AnnualSummarySlot } from "../features/site/api";

export function AdminAnnualSummaryPage() {
  const currentYear = new Date().getFullYear(); const [year, setYear] = useState(currentYear); const data = useAdminAnnualSummary(year); const save = useSaveAnnualSummary(); const assets = useAllAdminAssets({ status: "ready", visible: true, private: false }); const [slots, setSlots] = useState<AnnualSummarySlot[]>(() => Array.from({ length: 10 }, (_, slot) => ({ slot, comment: "" }))); const [picker, setPicker] = useState<number>(); const [msg, holder] = message.useMessage();
  useEffect(() => { if (data.data) setSlots(data.data.slots); }, [data.data]);
  const byId = useMemo(() => new Map((assets.data ?? []).map((asset) => [asset.id, asset])), [assets.data]);
  const update = (index: number, patch: Partial<AnnualSummarySlot>) => setSlots((items) => items.map((slot, i) => i === index ? { ...slot, ...patch } : slot));
  async function submit() { await save.mutateAsync({ year, years: data.data?.years ?? [], slots }); msg.success("年度总结已保存"); }
  return <div className="admin-page-stack">{holder}<div className="admin-page-heading"><div><Typography.Title level={2}>年度总结</Typography.Title><Typography.Text type="secondary">为每个年份配置十张照片和对应评语</Typography.Text></div><Space><InputNumber min={1900} max={2200} value={year} onChange={(value) => value && setYear(value)} /><Button type="primary" icon={<SaveOutlined />} loading={save.isPending} onClick={submit}>保存 {year} 年</Button></Space></div>
    <div className="annual-admin-grid">{slots.map((slot, index) => { const asset = slot.asset_id ? byId.get(slot.asset_id) : undefined; return <div className="annual-admin-slot" key={index}><div className="annual-admin-slot-media">{asset ? <AssetMedia asset={asset} className="h-full w-full object-cover" /> : <span>第 {index + 1} 张</span>}</div><Button block icon={<EditOutlined />} onClick={() => setPicker(index)}>{asset ? "更换照片" : "选择照片"}</Button><Form.Item className="mb-0" label={`第 ${index + 1} 张评语`}><Input.TextArea rows={2} value={slot.comment} onChange={(event) => update(index, { comment: event.target.value })} placeholder="写下一句评语" /></Form.Item></div>; })}</div>
    <HomeAssetPicker open={picker !== undefined} assets={assets.data ?? []} currentId={picker === undefined ? undefined : slots[picker]?.asset_id} excludedIds={slots.map((slot) => slot.asset_id).filter((id): id is string => Boolean(id))} onCancel={() => setPicker(undefined)} onSelect={(asset: Asset) => { if (picker !== undefined) update(picker, { asset_id: asset.id }); setPicker(undefined); }} />
  </div>;
}
