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
          colorPrimary: "#ff6a00",
          colorInfo: "#0ea5e9",
          colorSuccess: "#16a34a",
          colorWarning: "#f59e0b",
          colorError: "#ef4444",
          borderRadius: 14,
          fontFamily: "var(--font-ui)"
        },
        components: {
          Layout: {
            bodyBg: "transparent",
            headerBg: "transparent",
            siderBg: "transparent",
            triggerBg: "transparent"
          },
          Menu: {
            itemBg: "transparent",
            itemHoverBg: "rgba(255, 106, 0, 0.1)",
            itemSelectedBg: "rgba(255, 106, 0, 0.18)",
            itemSelectedColor: "rgba(11, 19, 32, 0.96)",
            itemColor: "rgba(11, 19, 32, 0.8)"
          },
          Card: {
            headerBg: "transparent"
          },
          Button: {
            borderRadius: 12,
            controlHeight: 38
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
