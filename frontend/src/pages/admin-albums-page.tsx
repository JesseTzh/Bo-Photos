import { EditOutlined, PictureOutlined, PlusOutlined } from "@ant-design/icons";
import { Button, Image, Space, Table, Tag, Typography } from "antd";
import { useNavigate } from "react-router-dom";
import { type Album, useAlbums } from "../features/albums/api";

export function AdminAlbumsPage() {
  const navigate = useNavigate();
  const albums = useAlbums(true);

  return (
    <div className="admin-page-stack">
      <div className="panel-actions">
        <div>
          <Typography.Title>相册管理</Typography.Title>
          <Typography.Paragraph type="secondary">查看相册，并分别管理基本信息和相册图片。</Typography.Paragraph>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate("/admin/albums/new")}>新建相册</Button>
      </div>
      <Table<Album>
        rowKey="id"
        loading={albums.isPending}
        dataSource={albums.data?.items ?? []}
        scroll={{ x: 760 }}
        columns={[
          {
            title: "封面",
            width: 112,
            render: (_, item) => item.cover_url ? (
              <Image
                className="admin-album-cover"
                width={88}
                height={56}
                preview={false}
                src={item.cover_url}
                alt={item.name}
              />
            ) : <div className="admin-album-cover" />
          },
          { title: "名称", dataIndex: "name" },
          { title: "路由", dataIndex: "album_value", render: (value) => `/${value}` },
          { title: "图片", dataIndex: "asset_count", width: 90 },
          { title: "状态", width: 100, render: (_, item) => item.visible ? <Tag className="visibility-status-tag visibility-status-tag--visible">公开</Tag> : <Tag className="visibility-status-tag visibility-status-tag--hidden">隐藏</Tag> },
          {
            title: "操作",
            width: 280,
            render: (_, item) => (
              <Space wrap>
                <Button icon={<EditOutlined />} onClick={() => navigate(`/admin/albums/${item.id}/edit`)}>基本信息编辑</Button>
                <Button icon={<PictureOutlined />} onClick={() => navigate(`/admin/albums/${item.id}/images`)}>图片管理</Button>
              </Space>
            )
          }
        ]}
      />
    </div>
  );
}
