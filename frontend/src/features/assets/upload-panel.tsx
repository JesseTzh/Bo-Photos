import { CopyOutlined, DeleteOutlined, EditOutlined, InboxOutlined } from "@ant-design/icons";
import { Alert, Button, Collapse, Form, Image, Input, InputNumber, List, Modal, Progress, Select, Space, Switch, Tag, Upload, message } from "antd";
import { useEffect, useMemo, useState } from "react";
import { useAlbums } from "../albums/api";
import { useAdminSettings } from "../site/api";
import { flattenTags, useSaveAssetTags, useTags } from "../tags/api";
import { type AssetUpdate, useDeleteAssets, useSaveAssetAlbums, useUpdateAsset } from "./admin-api";
import { readReferenceExif } from "./exif";
import { defaultUploadMetadata, isRawUploadFile, RAW_UPLOAD_ACCEPT, type UploadMetadataDraft, type UploadQueueItem, useUploadQueue } from "./upload-queue";

interface UploadPanelProps {
  onAssetReady?: () => void;
}

export function UploadPanel({ onAssetReady }: UploadPanelProps) {
  const [messageApi, contextHolder] = message.useMessage();
  const [uploadDefaults, setUploadDefaults] = useState<UploadMetadataDraft>(defaultUploadMetadata);
  const [editingKey, setEditingKey] = useState<string>();
  const [defaultsForm] = Form.useForm();
  const [itemForm] = Form.useForm();
  const settings = useAdminSettings();
  const albums = useAlbums(true);
  const tags = useTags(true);
  const updateAsset = useUpdateAsset();
  const saveAssetTags = useSaveAssetTags();
  const saveAssetAlbums = useSaveAssetAlbums();
  const deleteAssets = useDeleteAssets();

  const albumOptions = useMemo(
    () => (albums.data?.items ?? []).map((item) => ({ value: item.id, label: item.name })),
    [albums.data]
  );
  const tagOptions = useMemo(
    () => flattenTags(tags.data?.items ?? []).map((item) => ({
      value: item.id,
      label: `${"—".repeat(item.depth)} ${item.name}`
    })),
    [tags.data]
  );

  const queue = useUploadQueue(2, uploadDefaults, async (item) => {
    if (!item.assetId) return;
    await updateAsset.mutateAsync({ id: item.assetId, input: toAssetUpdate(item.metadata) });
    await saveAssetTags.mutateAsync({ assetId: item.assetId, tagIds: item.metadata.tag_ids });
    await saveAssetAlbums.mutateAsync({ assetId: item.assetId, albumIds: item.metadata.album_ids });
    messageApi.success(`${item.file.name} 上传完成`);
    onAssetReady?.();
  });

  const editingItem = queue.items.find((item) => item.key === editingKey);

  useEffect(() => {
    defaultsForm.setFieldsValue(uploadDefaults);
  }, [defaultsForm, uploadDefaults]);

  useEffect(() => {
    if (editingItem) itemForm.setFieldsValue(editingItem.metadata);
  }, [editingItem, itemForm]);

  const maxUploadFiles = settings.data?.max_upload_files ?? 5;
  const activeCount = queue.items.filter((item) => !["ready", "failed"].includes(item.status)).length;

  function saveItemMetadata(values: UploadMetadataDraft) {
    if (!editingItem) return;
    queue.updateMetadata(editingItem.key, normalizeUploadDefaults(values));
    setEditingKey(undefined);
  }

  async function copyDuplicateIds(ids: string[]) {
    await navigator.clipboard?.writeText(ids.join("\n"));
    messageApi.success("已复制重复图片 ID");
  }

  async function deleteUploadedDuplicate(item: UploadQueueItem) {
    if (!item.assetId) return;
    try {
      await deleteAssets.mutateAsync([item.assetId]);
      queue.update(item.key, { status: "deleted" });
      messageApi.success("本次上传已删除");
    } catch (error) {
      messageApi.error(error instanceof Error ? error.message : "删除失败");
    }
  }

  return (
    <section className="asset-panel">
      {contextHolder}
      <Collapse
        size="small"
        className="upload-defaults"
        items={[{
          key: "defaults",
          label: "上传默认信息",
          forceRender: true,
          children: (
            <Form
              form={defaultsForm}
              layout="vertical"
              initialValues={uploadDefaults}
              onValuesChange={(_, values) => setUploadDefaults(normalizeUploadDefaults(values))}
            >
              <Form.Item name="album_ids" label="默认相册"><Select mode="multiple" allowClear options={albumOptions} loading={albums.isPending} /></Form.Item>
              <Form.Item name="tag_ids" label="默认标签"><Select mode="multiple" allowClear options={tagOptions} loading={tags.isPending} /></Form.Item>
              <Form.Item name="description" label="默认描述"><Input.TextArea rows={2} /></Form.Item>
              <Space wrap>
                <Form.Item name="visible" label="公开" valuePropName="checked"><Switch /></Form.Item>
                <Form.Item name="private" label="隐私" valuePropName="checked"><Switch /></Form.Item>
                <Form.Item name="show_on_homepage" label="首页显示" valuePropName="checked"><Switch /></Form.Item>
                <Form.Item name="sort" label="排序"><InputNumber /></Form.Item>
              </Space>
            </Form>
          )
        }]}
      />
      <Upload.Dragger
        className="asset-upload-dragger"
        multiple
        accept={RAW_UPLOAD_ACCEPT}
        showUploadList={false}
        beforeUpload={(file) => {
          if (activeCount >= maxUploadFiles) {
            messageApi.warning(`单次最多排队 ${maxUploadFiles} 个文件`);
            return Upload.LIST_IGNORE;
          }
          queue.addFiles([file as File], uploadDefaults);
          return Upload.LIST_IGNORE;
        }}
      >
        <p className="ant-upload-drag-icon"><InboxOutlined /></p>
        <p className="ant-upload-text">拖入图片、RAW 或视频，或点击选择多个文件</p>
      </Upload.Dragger>
      {queue.items.length ? (
        <>
          <div className="panel-actions">
            <strong>上传队列</strong>
            <Button size="small" onClick={queue.clearFinished}>清除已完成</Button>
          </div>
          <List
            className="upload-queue-list"
            dataSource={queue.items}
            renderItem={(item) => (
              <List.Item className="upload-queue-item">
                <List.Item.Meta
                  avatar={item.file.type.startsWith("video/") ? (
                    <video src={item.previewUrl} className="upload-preview" width={64} height={64} muted />
                  ) : isRawUploadFile(item.file) ? (
                    <div className="upload-preview upload-preview-raw" aria-hidden>RAW</div>
                  ) : <Image src={item.previewUrl} alt={item.file.name} className="upload-preview" width={64} height={64} />}
                  title={<Space wrap><span>{item.file.name}</span><Button size="small" icon={<EditOutlined />} onClick={() => setEditingKey(item.key)}>信息</Button></Space>}
                  description={<UploadItemDescription item={item} onCopy={copyDuplicateIds} onDelete={deleteUploadedDuplicate} />}
                />
                <div className="upload-status">
                  {item.status === "uploading" ? <Progress type="circle" percent={50} size={28} /> : null}
                  <Tag className={`asset-status-tag asset-status-tag--${item.status}`}>{item.status}</Tag>
                </div>
              </List.Item>
            )}
          />
        </>
      ) : null}
      <Modal
        open={Boolean(editingItem)}
        title="素材信息"
        onCancel={() => setEditingKey(undefined)}
        onOk={() => itemForm.submit()}
        width={760}
      >
        <Form form={itemForm} layout="vertical" onFinish={saveItemMetadata}>
          <Form.Item name="title" label="标题"><Input /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="album_ids" label="相册"><Select mode="multiple" allowClear options={albumOptions} /></Form.Item>
          <Form.Item name="tag_ids" label="标签"><Select mode="multiple" allowClear options={tagOptions} /></Form.Item>
          <Space wrap>
            <Form.Item name="visible" label="公开" valuePropName="checked"><Switch /></Form.Item>
            <Form.Item name="private" label="隐私" valuePropName="checked"><Switch /></Form.Item>
            <Form.Item name="show_on_homepage" label="首页显示" valuePropName="checked"><Switch /></Form.Item>
            <Form.Item name="sort" label="排序"><InputNumber /></Form.Item>
          </Space>
          <div className="upload-editor-grid">
            <Form.Item name="shoot_at" label="拍摄时间"><Input placeholder="2026-06-26T01:02:03Z" /></Form.Item>
            <Form.Item name="camera" label="相机"><Input /></Form.Item>
            <Form.Item name="lens" label="镜头"><Input /></Form.Item>
            <Form.Item name="exposure_time" label="快门"><Input /></Form.Item>
            <Form.Item name="aperture" label="光圈"><Input /></Form.Item>
            <Form.Item name="iso" label="ISO"><Input /></Form.Item>
            <Form.Item name="focal_length" label="焦距"><Input /></Form.Item>
            <Form.Item name="latitude" label="纬度"><InputNumber style={{ width: "100%" }} /></Form.Item>
            <Form.Item name="longitude" label="经度"><InputNumber style={{ width: "100%" }} /></Form.Item>
          </div>
          <Form.Item name="exif_json" label="EXIF JSON"><Input.TextArea rows={4} /></Form.Item>
          <Upload
            showUploadList={false}
            beforeUpload={(file) => {
              void readReferenceExif(file as File)
                .then((summary) => {
                  itemForm.setFieldsValue({ ...itemForm.getFieldsValue(), ...summary });
                  messageApi.success("EXIF 已导入");
                })
                .catch((error) => messageApi.error(error instanceof Error ? error.message : "EXIF 读取失败"));
              return Upload.LIST_IGNORE;
            }}
          >
            <Button>导入参考 EXIF</Button>
          </Upload>
        </Form>
      </Modal>
    </section>
  );
}

function UploadItemDescription({
  item,
  onCopy,
  onDelete
}: {
  item: UploadQueueItem;
  onCopy: (ids: string[]) => void | Promise<void>;
  onDelete: (item: UploadQueueItem) => void | Promise<void>;
}) {
  if (item.error) return <span>{item.error}</span>;
  if (!item.duplicateAssetIds.length) return item.metadataApplied ? <span>信息已写入</span> : null;
  return (
    <Alert
      type="warning"
      showIcon
      message={`发现 ${item.duplicateAssetIds.length} 个相同内容`}
      action={(
        <Space>
          <Button size="small" icon={<CopyOutlined />} onClick={() => void onCopy(item.duplicateAssetIds)} />
          <Button size="small" danger icon={<DeleteOutlined />} disabled={!item.assetId} onClick={() => void onDelete(item)} />
        </Space>
      )}
    />
  );
}

function normalizeUploadDefaults(values: Partial<UploadMetadataDraft>): UploadMetadataDraft {
  return {
    album_ids: values.album_ids ?? [],
    tag_ids: values.tag_ids ?? [],
    title: emptyToUndefined(values.title),
    description: emptyToUndefined(values.description),
    visible: values.visible ?? true,
    private: values.private ?? false,
    show_on_homepage: values.show_on_homepage ?? true,
    sort: values.sort ?? 0,
    shoot_at: emptyToUndefined(values.shoot_at),
    camera: emptyToUndefined(values.camera),
    lens: emptyToUndefined(values.lens),
    exposure_time: emptyToUndefined(values.exposure_time),
    aperture: emptyToUndefined(values.aperture),
    iso: emptyToUndefined(values.iso),
    focal_length: emptyToUndefined(values.focal_length),
    latitude: values.latitude,
    longitude: values.longitude,
    exif_json: emptyToUndefined(values.exif_json)
  };
}

function toAssetUpdate(metadata: UploadMetadataDraft): AssetUpdate {
  return {
    title: metadata.title,
    description: metadata.description,
    visible: metadata.visible,
    private: metadata.private,
    show_on_homepage: metadata.show_on_homepage,
    sort: metadata.sort,
    shoot_at: metadata.shoot_at,
    camera: metadata.camera,
    lens: metadata.lens,
    exposure_time: metadata.exposure_time,
    aperture: metadata.aperture,
    iso: metadata.iso,
    focal_length: metadata.focal_length,
    latitude: metadata.latitude,
    longitude: metadata.longitude,
    exif_json: metadata.exif_json
  };
}

function emptyToUndefined(value?: string) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}
