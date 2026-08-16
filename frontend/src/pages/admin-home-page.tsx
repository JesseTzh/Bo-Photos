import { DeleteOutlined, EditOutlined, EyeOutlined, PlusOutlined, VideoCameraOutlined } from "@ant-design/icons";
import { Button, Form, Input, Space, Switch, Tag, Tooltip, Typography, message } from "antd";
import { useEffect, useMemo, useState } from "react";
import { useAllAdminAssets } from "../features/assets/admin-api";
import { AssetMedia } from "../features/assets/asset-media";
import { isVideoAsset, type Asset } from "../features/assets/schema";
import { HomeAssetPicker, HomeAssetPreview } from "../features/site/home-asset-picker";
import { useAdminSettings, useSaveSettings } from "../features/site/api";

type PickerTarget = { kind: "cover" } | { kind: "carousel"; index: number };

export function AdminHomePage() {
  const settings = useAdminSettings();
  const save = useSaveSettings();
  const assets = useAllAdminAssets({ status: "ready", visible: true, private: false });
  const [form] = Form.useForm();
  const [messageApi, contextHolder] = message.useMessage();
  const [coverId, setCoverId] = useState("");
  const [carouselIds, setCarouselIds] = useState<string[]>([]);
  const [pickerTarget, setPickerTarget] = useState<PickerTarget>();
  const [previewing, setPreviewing] = useState<Asset>();

  useEffect(() => {
    if (!settings.data) return;
    form.setFieldsValue(settings.data);
    setCoverId(settings.data.hero_asset_id || "");
    setCarouselIds((settings.data.hero_carousel_asset_ids ?? []).slice(0, 5));
  }, [form, settings.data]);

  const availableAssets = useMemo(
    () => (assets.data ?? []).filter((asset) => asset.status === "ready" && asset.visible && !asset.private),
    [assets.data]
  );
  const assetMap = useMemo(() => new Map(availableAssets.map((asset) => [asset.id, asset])), [availableAssets]);
  const currentPickerId = pickerTarget?.kind === "cover" ? coverId : pickerTarget ? carouselIds[pickerTarget.index] : undefined;
  const excludedIds = pickerTarget?.kind === "cover"
    ? carouselIds
    : pickerTarget
      ? [coverId, ...carouselIds.filter((_, index) => index !== pickerTarget.index)].filter(Boolean)
      : [];

  function selectAsset(asset: Asset) {
    if (!pickerTarget) return;
    if (pickerTarget.kind === "cover") {
      setCoverId(asset.id);
    } else {
      setCarouselIds((current) => {
        const next = [...current];
        next[pickerTarget.index] = asset.id;
        return next;
      });
    }
    setPickerTarget(undefined);
  }

  async function submit(values: Record<string, unknown>) {
    if (!settings.data) return;
    await save.mutateAsync({
      ...settings.data,
      ...values,
      hero_asset_id: coverId,
      hero_carousel_asset_ids: carouselIds.filter(Boolean).slice(0, 5)
    });
    messageApi.success("首页设置已保存");
  }

  return (
    <div className="admin-page-stack admin-home-page">
      {contextHolder}
      <div>
        <Typography.Title>首页管理</Typography.Title>
        <Typography.Text type="secondary">管理首页封面、轮播顺序、封面文字和视频呈现方式。</Typography.Text>
      </div>
      <Form form={form} layout="vertical" onFinish={submit}>
        <section className="home-admin-section">
          <Typography.Title level={4}>封面素材</Typography.Title>
          <Typography.Paragraph type="secondary">封面是访客进入首页后首先看到的素材，也会作为轮播的第一项。</Typography.Paragraph>
          <div className="home-cover-slot">
            <HomeAssetSlot
              asset={assetMap.get(coverId)}
              label="首页封面"
              loading={assets.isPending}
              onChoose={() => setPickerTarget({ kind: "cover" })}
              onPreview={setPreviewing}
              onClear={() => setCoverId("")}
            />
          </div>
        </section>

        <section className="home-admin-section">
          <Typography.Title level={4}>轮播大图</Typography.Title>
          <Typography.Paragraph type="secondary">按槽位顺序播放，空槽位会自动跳过，最多设置 5 项。</Typography.Paragraph>
          <div className="home-carousel-slots">
            {Array.from({ length: 5 }, (_, index) => (
              <HomeAssetSlot
                key={index}
                asset={assetMap.get(carouselIds[index])}
                label={`轮播 ${index + 1}`}
                loading={assets.isPending}
                onChoose={() => setPickerTarget({ kind: "carousel", index })}
                onPreview={setPreviewing}
                onClear={() => setCarouselIds((current) => current.filter((_, itemIndex) => itemIndex !== index))}
              />
            ))}
          </div>
        </section>

        <section className="home-admin-section">
          <Typography.Title level={4}>封面文字</Typography.Title>
          <Form.Item name="hero_show_text" label="显示封面文字" valuePropName="checked"><Switch /></Form.Item>
          <div className="home-copy-grid">
            <Form.Item name="hero_eyebrow" label="顶部标识"><Input maxLength={60} placeholder="Photography" /></Form.Item>
            <Form.Item name="hero_kicker" label="标题上方文字"><Input maxLength={80} placeholder="Visual Storytelling" /></Form.Item>
            <Form.Item name="hero_title" label="主标题"><Input maxLength={80} placeholder="Every Moment" /></Form.Item>
            <Form.Item name="hero_accent_title" label="强调标题"><Input maxLength={80} placeholder="Tells a Story" /></Form.Item>
          </div>
          <Form.Item name="hero_description" label="说明文字"><Input.TextArea rows={3} maxLength={240} showCount /></Form.Item>
        </section>

        <section className="home-admin-section">
          <Typography.Title level={4}>视频设置</Typography.Title>
          <Form.Item name="hero_video_overlay" label="视频显示遮罩" valuePropName="checked" extra="关闭后，首页视频上不再叠加渐暗层和暗角。">
            <Switch />
          </Form.Item>
        </section>

        <Button type="primary" htmlType="submit" loading={save.isPending} disabled={!settings.data}>保存首页设置</Button>
      </Form>

      <HomeAssetPicker
        open={Boolean(pickerTarget)}
        assets={availableAssets}
        currentId={currentPickerId}
        excludedIds={excludedIds}
        onCancel={() => setPickerTarget(undefined)}
        onSelect={selectAsset}
      />
      <HomeAssetPreview asset={previewing} onClose={() => setPreviewing(undefined)} />
    </div>
  );
}

function HomeAssetSlot({ asset, label, loading, onChoose, onPreview, onClear }: {
  asset?: Asset;
  label: string;
  loading: boolean;
  onChoose: () => void;
  onPreview: (asset: Asset) => void;
  onClear: () => void;
}) {
  return (
    <div className={`home-asset-slot${asset ? " has-asset" : ""}`}>
      <button type="button" className="home-asset-slot-preview" onClick={onChoose} disabled={loading}>
        {asset ? (
          <>
            <AssetMedia asset={asset} muted />
            <span className="home-asset-slot-change"><EditOutlined /> 更换</span>
            {isVideoAsset(asset) ? <Tag icon={<VideoCameraOutlined />}>视频</Tag> : null}
          </>
        ) : (
          <span className="home-asset-slot-empty"><PlusOutlined />{loading ? "加载素材中" : "选择素材"}</span>
        )}
      </button>
      <div className="home-asset-slot-footer">
        <div>
          <strong>{label}</strong>
          <span title={asset?.title || asset?.original_name}>{asset ? asset.title || asset.original_name : "未设置"}</span>
        </div>
        {asset ? (
          <Space size={4}>
            <Tooltip title="查看"><Button type="text" icon={<EyeOutlined />} onClick={() => onPreview(asset)} /></Tooltip>
            <Tooltip title="清除"><Button type="text" danger icon={<DeleteOutlined />} onClick={onClear} /></Tooltip>
          </Space>
        ) : null}
      </div>
    </div>
  );
}
