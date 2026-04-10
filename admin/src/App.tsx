import { App as AntdApp, ConfigProvider, theme } from "antd";
import zhCN from "antd/locale/zh_CN";
import { Navigate, Route, BrowserRouter as Router, Routes } from "react-router-dom";
import { AdminLayout } from "./layout/AdminLayout";
import { AuditLogsPage } from "./pages/AuditLogsPage";
import { CategoriesPage } from "./pages/CategoriesPage";
import { DashboardPage } from "./pages/DashboardPage";
import { LoginPage } from "./pages/LoginPage";
import { NotFoundPage } from "./pages/NotFoundPage";
import { ParamTemplatesPage } from "./pages/ParamTemplatesPage";
import { SystemSettingsPage } from "./pages/SystemSettingsPage";
import { VehiclesPage } from "./pages/VehiclesPage";

export default function App() {
  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: theme.defaultAlgorithm,
        token: {
          colorPrimary: "#f97316", // Orange-500: sharp active personality
          colorBgContainer: "#ffffff",
          borderRadius: 8, // Just a tiny bit of roundness
          fontFamily: "var(--font-ui)"
        },
        components: {
          Layout: {
            bodyBg: "#f0f2f5",
            headerBg: "#ffffff",
            siderBg: "#0f172a", // Sleek dark sidebar
          },
          Menu: {
            itemBorderRadius: 8,
            darkItemBg: "#0f172a",
            darkSubMenuItemBg: "#0f172a",
          },
          Card: {
            headerBg: "transparent"
          }
        }
      }}
    >
      <AntdApp>
        <Router>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/" element={<AdminLayout />}>
              <Route index element={<Navigate to="/dashboard" replace />} />
              <Route path="dashboard" element={<DashboardPage />} />
              <Route path="vehicles" element={<VehiclesPage />} />
              <Route path="categories" element={<CategoriesPage />} />
              <Route path="param-templates" element={<ParamTemplatesPage />} />
              <Route path="system-settings" element={<SystemSettingsPage />} />
              <Route path="audit-logs" element={<AuditLogsPage />} />
            </Route>
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </Router>
      </AntdApp>
    </ConfigProvider>
  );
}
