import { Button, Form, Input, InputNumber, Select, Switch, Typography, message } from "antd";
import { useEffect, useMemo } from "react";
import { useAdminAssets } from "../features/assets/admin-api";
import { useAdminSettings, useSaveSettings } from "../features/site/api";

export function AdminSettingsPage() {
  const settings = useAdminSettings();
  const save = useSaveSettings();
  const assets = useAdminAssets({ page: 1, pageSize: 200, status: "ready" });
  const [form] = Form.useForm();
  const [messageApi, contextHolder] = message.useMessage();

  useEffect(() => {
    if (settings.data) form.setFieldsValue(settings.data);
  }, [form, settings.data]);

  const assetOptions = useMemo(
    () => (assets.data?.items ?? []).filter((asset) => !asset.mime_type?.startsWith("video/")).map((asset) => ({
      value: asset.id,
      label: `[图片] ${asset.title || asset.original_name}`
    })),
    [assets.data?.items]
  );
  return (
    <div className="admin-page-stack">
      {contextHolder}
      <Typography.Title>站点设置</Typography.Title>
      <Form
        form={form}
        layout="vertical"
        onFinish={async (values) => {
          if (!settings.data) return;
          await save.mutateAsync({ ...settings.data, ...values });
          messageApi.success("设置已保存");
        }}
      >
        <Typography.Title level={4}>站点信息</Typography.Title>
        <Form.Item name="site_title" label="站点标题"><Input /></Form.Item>
        <Form.Item name="site_author" label="作者"><Input /></Form.Item>
        <Form.Item name="site_favicon_url" label="图标 URL"><Input /></Form.Item>
        <Form.Item name="about_intro" label="About 介绍"><Input.TextArea /></Form.Item>
        <Form.Item name="about_gallery_asset_ids" label="About 图片"><Select mode="multiple" options={assetOptions} /></Form.Item>
        {["about_social_instagram", "about_social_xiaohongshu", "about_social_weibo", "about_social_github"].map((name) => (
          <Form.Item name={name} label={name} key={name}><Input /></Form.Item>
        ))}

        <Typography.Title level={4}>图库与处理</Typography.Title>
        <Form.Item name="gallery_layout" label="图库布局"><Select options={[{ value: "grid", label: "网格" }, { value: "single", label: "单列" }]} /></Form.Item>
        <Form.Item name="public_original_download" label="允许公开下载原图" valuePropName="checked"><Switch /></Form.Item>
        <Form.Item name="admin_images_per_page" label="后台每页素材"><InputNumber min={1} max={200} /></Form.Item>
        <Form.Item name="max_upload_files" label="一次最多上传"><InputNumber min={1} max={100} /></Form.Item>
        <Form.Item name="preview_quality" label="预览质量"><InputNumber min={1} max={100} /></Form.Item>
        <Form.Item name="preview_max_width" label="预览最大宽度"><InputNumber min={320} max={16384} /></Form.Item>
        <Form.Item name="analytics_enabled" label="访问统计" valuePropName="checked"><Switch /></Form.Item>
        <Form.Item name="analytics_retention_days" label="日志保留天数"><InputNumber min={1} max={3650} /></Form.Item>
        <Form.Item name="analytics_timezone" label="统计时区"><Input /></Form.Item>
        <Button htmlType="submit" type="primary" loading={save.isPending}>保存</Button>
      </Form>
    </div>
  );
}
