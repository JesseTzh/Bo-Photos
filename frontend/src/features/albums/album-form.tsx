import { Form, Input, InputNumber, Select, Space, Switch } from "antd";

interface AlbumFormProps {
  assetOptions: Array<{ value: string; label: string }>;
}

const themeOptions = [
  { value: "0", label: "默认样式" },
  { value: "1", label: "简洁样式" }
];

const imageSortingOptions = [
  { value: 1, label: "上传时间从新到旧" },
  { value: 2, label: "拍摄时间从新到旧" },
  { value: 3, label: "上传时间从旧到新" },
  { value: 4, label: "拍摄时间从旧到新" }
];

export function AlbumForm({ assetOptions }: AlbumFormProps) {
  return (
    <>
      <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="album_value" label="路由值" rules={[{ required: true }]}><Input /></Form.Item>
      <Form.Item name="detail" label="描述"><Input.TextArea rows={3} /></Form.Item>
      <Space wrap>
        <Form.Item name="sort" label="排序"><InputNumber /></Form.Item>
        <Form.Item name="theme" label="主题"><Select style={{ width: 160 }} options={themeOptions} /></Form.Item>
        <Form.Item name="image_sorting" label="图片排序"><Select style={{ width: 200 }} options={imageSortingOptions} /></Form.Item>
        <Form.Item name="visible" label="公开" valuePropName="checked"><Switch /></Form.Item>
        <Form.Item name="random_show" label="随机显示" valuePropName="checked"><Switch /></Form.Item>
      </Space>
      <Form.Item name="license" label="授权协议"><Input /></Form.Item>
      <Form.Item name="cover_asset_id" label="封面"><Select allowClear showSearch options={assetOptions} /></Form.Item>
      <Form.Item name="asset_ids" label="关联图片"><Select mode="multiple" allowClear showSearch options={assetOptions} /></Form.Item>
    </>
  );
}
