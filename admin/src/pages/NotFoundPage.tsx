import { Button, Result } from "antd";
import { useNavigate } from "react-router-dom";

export function NotFoundPage() {
  const navigate = useNavigate();
  return (
    <Result
      status="404"
      title="404"
      subTitle="\u9875\u9762\u4e0d\u5b58\u5728"
      extra={
        <Button type="primary" onClick={() => navigate("/dashboard")}>
          {"\u8fd4\u56de\u4eea\u8868\u76d8"}
        </Button>
      }
    />
  );
}
