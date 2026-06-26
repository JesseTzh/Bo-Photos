import { CameraOutlined, EyeOutlined, FolderOpenOutlined, LineChartOutlined, PictureOutlined, ReadOutlined } from "@ant-design/icons";
import { Card, Empty, Skeleton, Space, Typography } from "antd";
import { Link } from "react-router-dom";
import { useDashboard, useDisk, type CountPoint, type Dashboard, type NamedCount } from "../features/site/api";

export function AdminDashboardPage() {
  const dashboard = useDashboard();
  const disk = useDisk();
  const data = dashboard.data;
  const diskLabel = disk.data ? `${(disk.data.used / 1073741824).toFixed(1)} / ${(disk.data.total / 1073741824).toFixed(1)} GB` : "-";

  if (dashboard.isPending) return <Skeleton active />;

  return (
    <div className="admin-page-stack">
      <Typography.Title>仪表盘</Typography.Title>
      {data ? (
        <>
          <div className="admin-dashboard-grid">
            {statCards(data, diskLabel).map((item) => (
              <Card className="admin-dashboard-card" key={item.title}>
                {item.href ? <Link to={item.href}><StatContent {...item} /></Link> : <StatContent {...item} />}
              </Card>
            ))}
          </div>
          <div className="admin-dashboard-sections">
            <Card title="最近 7 天访问"><TrendChart data={data.last_7_days} /></Card>
            <Card title="相机排行"><RankingList data={data.TopCameras} /></Card>
            <Card title="镜头排行"><RankingList data={data.TopLenses} /></Card>
            <Card title="年份分布"><RankingList data={data.PhotosByYear} /></Card>
          </div>
        </>
      ) : (
        <Empty description="暂无仪表盘数据" />
      )}
    </div>
  );
}

function statCards(data: Dashboard, diskLabel: string) {
  return [
    { title: "图片", value: `${data.ImagesTotal}`, meta: `公开 ${data.ImagesPublic}`, icon: <PictureOutlined />, href: "/admin/images" },
    { title: "相册", value: `${data.AlbumsTotal}`, meta: "全部相册", icon: <FolderOpenOutlined />, href: "/admin/albums" },
    { title: "Guides", value: `${data.GuidesTotal}`, meta: `发布 ${data.GuidesPublic}`, icon: <ReadOutlined />, href: "/admin/guides" },
    { title: "相机", value: `${data.CamerasTotal}`, meta: "已识别型号", icon: <CameraOutlined /> },
    { title: "镜头", value: `${data.LensesTotal}`, meta: "已识别镜头", icon: <CameraOutlined /> },
    { title: "今日访问", value: `${data.VisitsToday}`, meta: `昨日 ${data.VisitsYesterday} / 总计 ${data.VisitsTotal}`, icon: <EyeOutlined />, href: "/admin/analytics" },
    { title: "磁盘", value: diskLabel, meta: "本地存储", icon: <LineChartOutlined /> }
  ];
}

function StatContent({ title, value, meta, icon }: { title: string; value: string; meta: string; icon: React.ReactNode }) {
  return (
    <Space align="start" size="middle">
      <span className="admin-stat-icon">{icon}</span>
      <span>
        <small>{title}</small>
        <strong>{value}</strong>
        <em>{meta}</em>
      </span>
    </Space>
  );
}

function TrendChart({ data }: { data?: CountPoint[] }) {
  const max = Math.max(1, ...(data ?? []).map((item) => item.count));
  if (!data?.length) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无访问数据" />;
  return (
    <div className="admin-trend-chart">
      {data.map((item) => (
        <div className="admin-trend-bar" key={item.date}>
          <span style={{ height: `${Math.max(8, (item.count / max) * 100)}%` }} />
          <small>{item.date.slice(5)}</small>
        </div>
      ))}
    </div>
  );
}

function RankingList({ data }: { data?: NamedCount[] }) {
  const max = Math.max(1, ...(data ?? []).map((item) => item.count));
  if (!data?.length) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无数据" />;
  return (
    <div className="admin-ranking-list">
      {data.map((item) => (
        <div className="admin-ranking-row" key={item.name}>
          <span className="admin-ranking-name">{item.name || "Unknown"}</span>
          <span className="admin-ranking-meter"><i style={{ width: `${(item.count / max) * 100}%` }} /></span>
          <strong>{item.count}</strong>
        </div>
      ))}
    </div>
  );
}
