import { StarOutlined, UploadOutlined } from "@ant-design/icons";
import { Button, Divider, Drawer, Form, Input, InputNumber, Select, Space, Switch, Upload, message } from "antd";
import { useEffect, useMemo, useState } from "react";
import { type Album, useAlbums, useSetAlbumCover } from "../albums/api";
import { flattenTags, useAssetTags, useSaveAssetTags, useTags } from "../tags/api";
import { type AssetUpdate, useAssetAlbums, useSaveAssetAlbums, useUpdateAsset } from "./admin-api";
import { readReferenceExif } from "./exif";
import { isVideoAsset, type Asset } from "./schema";

interface AdminAssetEditDrawerProps {
  asset?: Asset;
  open: boolean;
  onClose: () => void;
  onSaved: () => void;
}

export function AdminAssetEditDrawer({ asset, open, onClose, onSaved }: AdminAssetEditDrawerProps) {
  const [form] = Form.useForm();
  const [messageApi, contextHolder] = message.useMessage();
  const [coverAlbumId, setCoverAlbumId] = useState<string>();
  const tags = useTags(true);
  const albums = useAlbums(true);
  const assetTags = useAssetTags(asset?.id);
  const assetAlbums = useAssetAlbums(asset?.id);
  const updateAsset = useUpdateAsset();
  const saveAssetTags = useSaveAssetTags();
  const saveAssetAlbums = useSaveAssetAlbums();
  const setAlbumCover = useSetAlbumCover();

  const tagOptions = useMemo(
    () => flattenTags(tags.data?.items ?? []).map((item) => ({ value: item.id, label: `${"—".repeat(item.depth)} ${item.name}` })),
    [tags.data]
  );
  const albumItems = albums.data?.items ?? [];
  const albumOptions = albumItems.map((item) => ({ value: item.id, label: item.name }));
  const boundAlbums = albumItems.filter((album) => assetAlbums.data?.album_ids.includes(album.id));

  useEffect(() => {
    if (!asset) return;
    form.setFieldsValue({
      ...asset,
      tag_ids: assetTags.data?.tag_ids ?? [],
      album_ids: assetAlbums.data?.album_ids ?? []
    });
    setCoverAlbumId(assetAlbums.data?.album_ids[0]);
  }, [asset, assetAlbums.data, assetTags.data, form]);

  async function save(values: Record<string, unknown>) {
    if (!asset) return;
    const tagIds = (values.tag_ids as string[] | undefined) ?? [];
    const albumIds = (values.album_ids as string[] | undefined) ?? [];
    await updateAsset.mutateAsync({ id: asset.id, input: toAssetInput(values) });
    await saveAssetTags.mutateAsync({ assetId: asset.id, tagIds });
    await saveAssetAlbums.mutateAsync({ assetId: asset.id, albumIds });
    messageApi.success("素材信息已保存");
    onSaved();
    onClose();
  }

  async function makeCover() {
    if (!asset || !coverAlbumId) return;
    await setAlbumCover.mutateAsync({ albumId: coverAlbumId, assetId: asset.id });
    messageApi.success("相册封面已更新");
  }

  return (
    <Drawer open={open} size={720} title="编辑素材" onClose={onClose}>
      {contextHolder}
      {asset ? (
        <Form form={form} layout="vertical" onFinish={save}>
          <Divider>基本信息</Divider>
          <Form.Item name="title" label="标题"><Input /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="tag_ids" label="标签"><Select mode="multiple" allowClear options={tagOptions} /></Form.Item>
          <Form.Item name="album_ids" label="相册"><Select mode="multiple" allowClear options={albumOptions} /></Form.Item>
          <Space wrap>
            <Form.Item name="visible" label="公开" valuePropName="checked"><Switch /></Form.Item>
            <Form.Item name="private" label="隐私" valuePropName="checked"><Switch /></Form.Item>
            <Form.Item name="show_on_homepage" label="首页显示" valuePropName="checked"><Switch /></Form.Item>
            <Form.Item name="sort" label="排序"><InputNumber /></Form.Item>
          </Space>
          {boundAlbums.length && !isVideoAsset(asset) ? (
            <Space wrap className="asset-cover-action">
              <Select style={{ minWidth: 220 }} value={coverAlbumId} options={boundAlbums.map(albumOption)} onChange={setCoverAlbumId} />
              <Button icon={<StarOutlined />} onClick={() => void makeCover()} loading={setAlbumCover.isPending}>设为相册封面</Button>
            </Space>
          ) : null}

          <Divider>尺寸与位置</Divider>
          <div className="upload-editor-grid">
            <Form.Item name="width" label="宽度"><InputNumber min={0} style={{ width: "100%" }} /></Form.Item>
            <Form.Item name="height" label="高度"><InputNumber min={0} style={{ width: "100%" }} /></Form.Item>
            <Form.Item name="longitude" label="经度"><InputNumber style={{ width: "100%" }} /></Form.Item>
            <Form.Item name="latitude" label="纬度"><InputNumber style={{ width: "100%" }} /></Form.Item>
          </div>

          <Divider>EXIF 信息</Divider>
          <div className="upload-editor-grid">
            <Form.Item name="camera" label="相机"><Input /></Form.Item>
            <Form.Item name="lens" label="镜头"><Input /></Form.Item>
            <Form.Item name="focal_length" label="焦距"><Input /></Form.Item>
            <Form.Item name="aperture" label="光圈"><Input /></Form.Item>
            <Form.Item name="exposure_time" label="快门"><Input /></Form.Item>
            <Form.Item name="iso" label="ISO"><Input /></Form.Item>
            <Form.Item name="shoot_at" label="拍摄时间"><Input /></Form.Item>
          </div>
          <Upload
            showUploadList={false}
            beforeUpload={(file) => {
              void readReferenceExif(file as File)
                .then((summary) => {
                  form.setFieldsValue({ ...form.getFieldsValue(), ...summary });
                  messageApi.success("EXIF 已导入");
                })
                .catch((error) => messageApi.error(error instanceof Error ? error.message : "EXIF 读取失败"));
              return Upload.LIST_IGNORE;
            }}
          >
            <Button icon={<UploadOutlined />}>导入参考 EXIF</Button>
          </Upload>

          <Divider>高级</Divider>
          <Form.Item name="exif_json" label="EXIF JSON"><Input.TextArea rows={5} /></Form.Item>
          <div className="drawer-actions">
            <Button onClick={onClose}>取消</Button>
            <Button type="primary" loading={updateAsset.isPending || saveAssetTags.isPending || saveAssetAlbums.isPending} onClick={() => form.submit()}>保存</Button>
          </div>
        </Form>
      ) : null}
    </Drawer>
  );
}

function albumOption(item: Album) {
  return { value: item.id, label: item.name };
}

function toAssetInput(values: Record<string, unknown>): AssetUpdate {
  return {
    title: stringValue(values.title),
    description: stringValue(values.description),
    width: numberValue(values.width),
    height: numberValue(values.height),
    longitude: nullableNumber(values.longitude),
    latitude: nullableNumber(values.latitude),
    shoot_at: stringValue(values.shoot_at),
    camera: stringValue(values.camera),
    lens: stringValue(values.lens),
    exposure_time: stringValue(values.exposure_time),
    aperture: stringValue(values.aperture),
    iso: stringValue(values.iso),
    focal_length: stringValue(values.focal_length),
    exif_json: stringValue(values.exif_json),
    visible: values.visible as boolean | undefined,
    private: values.private as boolean | undefined,
    show_on_homepage: values.show_on_homepage as boolean | undefined,
    sort: numberValue(values.sort)
  };
}

function stringValue(value: unknown) {
  return typeof value === "string" ? value : undefined;
}

function numberValue(value: unknown) {
  return typeof value === "number" ? value : undefined;
}

function nullableNumber(value: unknown) {
  if (value === null) return null;
  return typeof value === "number" ? value : undefined;
}
