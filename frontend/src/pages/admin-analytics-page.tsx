import { DownloadOutlined } from "@ant-design/icons";
import { Button, Card, Table, Typography } from "antd";
import { useAnalytics, type Analytics, type NamedCount } from "../features/site/api";

const namedColumns = [
  { title: "名称", dataIndex: "name" },
  { title: "访问", dataIndex: "count" }
];

export function AdminAnalyticsPage() {
  const query = useAnalytics();
  const analytics = query.data;
  return (
    <div className="admin-page-stack">
      <div className="admin-page-heading">
        <Typography.Title>访问统计</Typography.Title>
        <Button
          icon={<DownloadOutlined />}
          disabled={!analytics}
          onClick={() => analytics && downloadAnalyticsCSV(analytics)}
        >
          导出 CSV
        </Button>
      </div>
      <div className="album-grid">
        <Card title="累计">{analytics?.dashboard.VisitsTotal ?? 0}</Card>
        <Card title="今日">{analytics?.dashboard.VisitsToday ?? 0}</Card>
        <Card title="昨日">{analytics?.dashboard.VisitsYesterday ?? 0}</Card>
        <Card title="唯一访客">{analytics?.unique_visitors ?? 0}</Card>
      </div>
      <Typography.Title level={3}>最近 7 天</Typography.Title>
      <Table
        rowKey="date"
        pagination={false}
        dataSource={analytics?.dashboard.last_7_days}
        columns={[
          { title: "日期", dataIndex: "date" },
          { title: "访问", dataIndex: "count" }
        ]}
      />
      <Typography.Title level={3}>最近 24 小时</Typography.Title>
      <Table
        rowKey="date"
        pagination={false}
        dataSource={analytics?.hourly}
        columns={[
          { title: "小时", dataIndex: "date", render: (value: string) => new Date(value).toLocaleString() },
          { title: "访问", dataIndex: "count" }
        ]}
      />
      <Typography.Title level={3}>来源</Typography.Title>
      <Table rowKey="name" pagination={false} dataSource={analytics?.sources} columns={namedColumns} />
      <Typography.Title level={3}>页面类型</Typography.Title>
      <Table rowKey="name" pagination={false} dataSource={analytics?.pages} columns={namedColumns} />
    </div>
  );
}

function downloadAnalyticsCSV(analytics: Analytics) {
  const rows = [
    ["分类", "名称", "访问量"],
    ...analytics.dashboard.last_7_days.map((point) => ["最近 7 天", point.date, point.count]),
    ...analytics.hourly.map((point) => ["最近 24 小时", point.date, point.count]),
    ...namedRows("来源", analytics.sources),
    ...namedRows("页面类型", analytics.pages)
  ];
  const csv = `\uFEFF${rows.map((row) => row.map(csvCell).join(",")).join("\n")}`;
  const url = URL.createObjectURL(new Blob([csv], { type: "text/csv;charset=utf-8" }));
  const link = document.createElement("a");
  link.href = url;
  link.download = `bophotos-analytics-${new Date().toISOString().slice(0, 10)}.csv`;
  link.click();
  URL.revokeObjectURL(url);
}

function namedRows(category: string, values: NamedCount[]) {
  return values.map((value) => [category, value.name, value.count]);
}

function csvCell(value: string | number) {
  return `"${String(value).replaceAll('"', '""')}"`;
}
