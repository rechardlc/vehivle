import {
  AppstoreOutlined,
  BellOutlined,
  CarOutlined,
  DashboardOutlined,
  FileTextOutlined,
  LogoutOutlined,
  SettingOutlined,
  TagsOutlined
} from "@ant-design/icons";
import { Avatar, Button, Layout, Menu, Space, Typography } from "antd";
import { Navigate, Outlet, useLocation, useNavigate } from "react-router-dom";
import { clearAuthState, getAuthState } from "../state/auth";

const { Header, Sider, Content } = Layout;

const menus = [
  { key: "/dashboard", icon: <DashboardOutlined />, label: "仪表盘" },
  { key: "/vehicles", icon: <CarOutlined />, label: "车型管理" },
  { key: "/categories", icon: <AppstoreOutlined />, label: "分类管理" },
  { key: "/param-templates", icon: <TagsOutlined />, label: "参数模板" },
  { key: "/system-settings", icon: <SettingOutlined />, label: "系统设置" },
  { key: "/audit-logs", icon: <FileTextOutlined />, label: "审计日志" }
];

function resolveSelectedKey(pathname: string): string {
  const matched = menus.find((item) => pathname.startsWith(item.key));
  return matched?.key ?? "/dashboard";
}

export function AdminLayout() {
  const auth = getAuthState();
  const navigate = useNavigate();
  const location = useLocation();
  const selectedKey = resolveSelectedKey(location.pathname);
  const currentMenu = menus.find((item) => item.key === selectedKey);

  if (!auth) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="admin-shell">
      <div className="admin-ambient" />
      <Layout className="admin-layout">
        <Sider width={276} theme="light" className="admin-sider">
          <div className="brand">
            <div className="brand-pill">VEHIVLE</div>
            <Typography.Title level={5} className="brand-title">
              车辆电子展厅后台
            </Typography.Title>
            <Typography.Text className="brand-subtitle">Light Skeuomorphic Console</Typography.Text>
          </div>
          <Menu
            mode="inline"
            className="admin-menu"
            selectedKeys={[selectedKey]}
            items={menus}
            onClick={(info) => {
              navigate(info.key);
            }}
          />
        </Sider>
        <Layout className="admin-main">
          <Header className="top-header">
            <div className="top-header-main">
              <Typography.Text className="top-header-title">{currentMenu?.label}</Typography.Text>
              <Typography.Text className="top-header-subtitle">触感面板 · 运营工作台</Typography.Text>
            </div>
            <Space size={10}>
              <Button className="pressable soft" type="text" icon={<BellOutlined />} />
              <div className="user-chip">
                <Avatar>{auth.user.username.slice(0, 1).toUpperCase()}</Avatar>
                <div className="user-chip-meta">
                  <Typography.Text className="user-chip-name">{auth.user.username}</Typography.Text>
                  <Typography.Text className="user-chip-role">
                    {auth.user.role === "super_admin" ? "超级管理员" : "运营编辑"}
                  </Typography.Text>
                </div>
              </div>
              <Button
                className="pressable soft"
                type="text"
                icon={<LogoutOutlined />}
                onClick={() => {
                  clearAuthState();
                  navigate("/login", { replace: true });
                }}
              >
                退出登录
              </Button>
            </Space>
          </Header>
          <Content className="admin-content">
            <Outlet />
          </Content>
        </Layout>
      </Layout>
    </div>
  );
}
