import {
  BarChartOutlined,
  BookOutlined,
  CameraOutlined,
  DashboardOutlined,
  FolderOpenOutlined,
  LogoutOutlined,
  PictureOutlined,
  SettingOutlined,
  TagsOutlined,
  UserOutlined
} from "@ant-design/icons";
import { Button, Layout, Menu, Typography } from "antd";
import { useNavigate } from "react-router-dom";
import { Outlet, useLocation } from "react-router-dom";
import { useLogout } from "../features/auth/api";

const navigation = [
  { key: "dashboard", icon: <DashboardOutlined />, label: "仪表盘" },
  { key: "images", icon: <PictureOutlined />, label: "图片" },
  { key: "albums", icon: <FolderOpenOutlined />, label: "相册" },
  { key: "tags", icon: <TagsOutlined />, label: "标签" },
  { key: "guides", icon: <BookOutlined />, label: "Guides" },
  { key: "analytics", icon: <BarChartOutlined />, label: "访问统计" },
  { key: "settings", icon: <SettingOutlined />, label: "站点设置" },
  { key: "account", icon: <UserOutlined />, label: "账户" }
];

export function AdminPage() {
  const logout = useLogout();
  const navigate = useNavigate();
  const location = useLocation();

  async function signOut() {
    await logout.mutateAsync();
    navigate("/login", { replace: true });
  }

  return (
    <Layout className="admin-shell">
      <Layout.Sider width={240} breakpoint="lg" collapsedWidth="0" className="admin-sidebar">
        <button className="admin-brand" type="button" onClick={() => navigate("/")}>
          <CameraOutlined />
          <span>BoPhoto</span>
        </button>
        <Menu
          theme="light"
          mode="inline"
          selectedKeys={[location.pathname === "/admin" ? "dashboard" : location.pathname.split("/")[2] || "dashboard"]}
          items={navigation.map((item) => ({
            ...item,
            onClick: () => navigate(item.key === "dashboard" ? "/admin" : `/admin/${item.key}`)
          }))}
        />
      </Layout.Sider>
      <Layout>
        <Layout.Header className="admin-header">
          <Typography.Text>控制台</Typography.Text>
          <Button icon={<LogoutOutlined />} loading={logout.isPending} onClick={signOut}>
            退出
          </Button>
        </Layout.Header>
        <Layout.Content className="admin-content">
          <Outlet />
        </Layout.Content>
      </Layout>
    </Layout>
  );
}
