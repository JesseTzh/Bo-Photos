import { Button, Form, Input, Typography, message } from "antd";
import { useLocation, useNavigate } from "react-router-dom";
import { ApiError } from "../api/client";
import { useLogin } from "../features/auth/api";
import { AuthLayout } from "./auth-layout";

interface LoginForm {
  password: string;
}

export function LoginPage() {
  const login = useLogin();
  const location = useLocation();
  const navigate = useNavigate();
  const [messageApi, contextHolder] = message.useMessage();

  async function submit(values: LoginForm) {
    try {
      await login.mutateAsync(values);
      const target =
        typeof location.state === "object" &&
        location.state !== null &&
        "from" in location.state &&
        typeof location.state.from === "string"
          ? location.state.from
          : "/admin";
      navigate(target, { replace: true });
    } catch (error) {
      messageApi.error(error instanceof ApiError ? error.message : "登录失败");
    }
  }

  return (
    <AuthLayout>
      {contextHolder}
      <Typography.Title level={2}>管理员登录</Typography.Title>
      <Form<LoginForm> layout="vertical" onFinish={submit} requiredMark={false}>
        <Form.Item
          label="密码"
          name="password"
          rules={[{ required: true, message: "请输入密码" }]}
        >
          <Input.Password size="large" autoComplete="current-password" autoFocus />
        </Form.Item>
        <Button type="primary" htmlType="submit" size="large" block loading={login.isPending}>
          登录
        </Button>
      </Form>
    </AuthLayout>
  );
}
