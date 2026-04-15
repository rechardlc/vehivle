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
import { useQuery } from "@tanstack/react-query";
import { App, Avatar, Button, Layout, Menu, Space, Spin, Typography } from "antd";
import { Navigate, Outlet, useLocation, useNavigate } from "react-router-dom";
import { authApi } from "../api/auth";
import { clearStoredUser, getStoredAccessToken, getStoredUser, setStoredUser } from "../state/auth";

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
  const navigate = useNavigate();
  const location = useLocation();
  const { message } = App.useApp();
  const selectedKey = resolveSelectedKey(location.pathname);
  const currentMenu = menus.find((item) => item.key === selectedKey);

  /**
   * 启动时通过 /auth/me 校验 Access Token 有效性并获取最新用户信息。
   * initialData 使用 localStorage 缓存，避免每次刷新页面出现 loading 闪烁。
   */
  const cachedAccessToken = getStoredAccessToken();
  const cachedUser = cachedAccessToken ? getStoredUser() : null;
  const { data: user, isLoading, isError } = useQuery({
    queryKey: ["auth", "me"],
    queryFn: async () => {
      const result = await authApi.me();
      setStoredUser(result);
      return result;
    },
    initialData: cachedUser ?? undefined,
    retry: false,
    staleTime: 5 * 60 * 1000
  });

  if (isLoading && !cachedUser) {
    return (
      <div className="admin-shell" style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "100vh" }}>
        <Spin size="large" tip="正在验证登录状态…" />
      </div>
    );
  }

  if (isError && !cachedUser) {
    clearStoredUser();
    return <Navigate to="/login" replace />;
  }

  const displayUser = user ?? cachedUser;
  if (!displayUser) {
    return <Navigate to="/login" replace />;
  }

  /** 调用后端登出接口清除 Cookie，同时清除本地缓存 */
  async function handleLogout() {
    try {
      await authApi.logout();
    } catch {
      // 即使后端登出失败，也清除本地状态（Cookie 可能已失效）
    }
    clearStoredUser();
    message.success("已退出登录");
    navigate("/login", { replace: true });
  }

  return (
    <div className="admin-shell">
      <Layout className="admin-layout">
        <Sider width={276} theme="light" className="admin-sider">
          <div className="brand">
            <Typography.Title level={4} className="brand-title" style={{ margin: 0 }}>
              车辆电子展厅
            </Typography.Title>
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
                <Avatar>{displayUser.username.slice(0, 1).toUpperCase()}</Avatar>
                <div className="user-chip-meta">
                  <Typography.Text className="user-chip-name">{displayUser.username}</Typography.Text>
                  <Typography.Text className="user-chip-role">
                    {displayUser.role === "super_admin" ? "超级管理员" : "运营编辑"}
                  </Typography.Text>
                </div>
              </div>
              <Button
                className="pressable soft"
                type="text"
                icon={<LogoutOutlined />}
                onClick={handleLogout}
              >
                退出登录
              </Button>
            </Space>
          </Header>
          <Content className="admin-content">
            <div className="admin-route-outlet">
              <Outlet />
            </div>
          </Content>
        </Layout>
      </Layout>
    </div>
  );
}
