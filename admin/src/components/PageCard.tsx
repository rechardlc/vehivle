import { Card, Space, Typography } from "antd";
import type { PropsWithChildren, ReactNode } from "react";

interface PageCardProps {
  title: string;
  subtitle?: string;
  extra?: ReactNode;
}

export function PageCard({ title, subtitle, extra, children }: PropsWithChildren<PageCardProps>) {
  return (
    <Card
      className="page-card"
      styles={{
        header: { borderBottom: "none", paddingBottom: 4 },
        body: { paddingTop: 16 }
      }}
      title={
        <Space className="page-card-title" direction="vertical" size={3}>
          <Typography.Text className="page-card-kicker">Control Surface</Typography.Text>
          <Typography.Title level={4} style={{ margin: 0 }} className="page-card-main-title">
            {title}
          </Typography.Title>
          {subtitle ? (
            <Typography.Text className="page-card-subtitle" type="secondary">
              {subtitle}
            </Typography.Text>
          ) : null}
        </Space>
      }
      extra={extra}
    >
      {children}
    </Card>
  );
}
