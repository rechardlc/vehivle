import { LockOutlined, UserOutlined } from "@ant-design/icons";
import { useMutation } from "@tanstack/react-query";
import { App, Button, Card, Form, Input, Space, Typography } from "antd";
import { Navigate, useNavigate } from "react-router-dom";
import { authApi } from "../api/auth";
import { getStoredUser, setStoredUser } from "../state/auth";

interface LoginFormValue {
  username: string;
  password: string;
}

export function LoginPage() {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const storedUser = getStoredUser();

  const mutation = useMutation({
    mutationFn: async (payload: LoginFormValue) => {
      await authApi.login(payload);
      const user = await authApi.me();
      return user;
    },
    onSuccess: (user) => {
      setStoredUser(user);
      message.success("登录成功");
      navigate("/dashboard", { replace: true });
    },
    onError: (error: Error) => {
      message.error(error.message);
    }
  });

  if (storedUser) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <div className="login-shell">
      <div className="login-panel">
        <div className="login-intro">
          <Typography.Text className="login-kicker">VEHIVLE CONSOLE</Typography.Text>
          <Typography.Title level={2} className="login-title">
            触感后台
          </Typography.Title>
          <Typography.Paragraph className="login-desc">
            轻拟物界面已启用。你将进入统一的运营工作台，完成车型、分类、模板和系统配置管理。
          </Typography.Paragraph>
        </div>

        <Card className="login-card" style={{ width: "100%" }}>
            <Space direction="vertical" size={16} style={{ width: "100%" }}>
            <Typography.Title level={3} style={{ marginBottom: 0 }}>
              后台登录
            </Typography.Title>
            <Form<LoginFormValue>
              layout="vertical"
              onFinish={(values) => mutation.mutate(values)}
            >
              <Form.Item
                label="用户名"
                name="username"
                rules={[{ required: true, message: "请输入用户名" }]}
              >
                <Input prefix={<UserOutlined />} placeholder="请输入用户名" />
              </Form.Item>
              <Form.Item
                label="密码"
                name="password"
                rules={[{ required: true, message: "请输入密码" }]}
              >
                <Input.Password prefix={<LockOutlined />} placeholder="请输入密码" />
              </Form.Item>
              <Button
                className="pressable"
                type="primary"
                htmlType="submit"
                block
                loading={mutation.isPending}
              >
                登录
              </Button>
            </Form>
          </Space>
        </Card>
      </div>
    </div>
  );
}
