import { useQuery } from "@tanstack/react-query";
import { Card, Col, Row, Skeleton, Statistic, Tag, Typography } from "antd";
import dayjs from "dayjs";
import { dashboardApi } from "../api/dashboard";
import { PageCard } from "../components/PageCard";

interface StatCardItem {
  key: string;
  title: string;
  value: number;
  tone: "orange" | "sky" | "mint" | "slate";
}

export function DashboardPage() {
  const summaryQuery = useQuery({
    queryKey: ["dashboard-summary"],
    queryFn: () => dashboardApi.summary()
  });

  const data = summaryQuery.data;
  const statCards: StatCardItem[] = data
    ? [
        { key: "vehicles", title: "车型总数", value: data.vehicleCount, tone: "orange" },
        { key: "published", title: "已发布", value: data.publishedCount, tone: "sky" },
        { key: "draft", title: "草稿中", value: data.draftCount, tone: "mint" },
        { key: "categories", title: "分类数量", value: data.categoryCount, tone: "slate" }
      ]
    : [];

  const maxValue = Math.max(...statCards.map((item) => item.value), 1);

  return (
    <PageCard title="仪表盘" subtitle="运营数据总览与最近操作节奏">
      {summaryQuery.isLoading || !data ? (
        <Skeleton active />
      ) : (
        <>
          <div className="dashboard-hero">
            <div>
              <Typography.Title level={4} style={{ marginTop: 0, marginBottom: 8 }}>
                今日工作台状态正常
              </Typography.Title>
              <Typography.Text type="secondary">
                最近操作时间：{dayjs(data.latestOperationAt).format("YYYY-MM-DD HH:mm:ss")}
              </Typography.Text>
            </div>
            <Tag className="dashboard-hero-tag">运营中</Tag>
          </div>

          <Row gutter={[16, 16]}>
            {statCards.map((item) => (
              <Col key={item.key} xs={24} sm={12} xl={6}>
                <Card className={`kpi-card kpi-${item.tone}`}>
                  <Statistic title={item.title} value={item.value} />
                  <div className="kpi-track">
                    <span
                      style={{
                        width: `${Math.max(8, Math.round((item.value / maxValue) * 100))}%`
                      }}
                    />
                  </div>
                </Card>
              </Col>
            ))}
            <Col xs={24}>
              <Card className="timeline-card">
                <Typography.Text type="secondary">
                  节奏提示：已发布 / 草稿比率为{" "}
                  {data.vehicleCount > 0 ? `${Math.round((data.publishedCount / data.vehicleCount) * 100)}%` : "0%"}，
                  建议优先处理草稿内容，保持展厅更新频率。
                </Typography.Text>
              </Card>
            </Col>
          </Row>
        </>
      )}
    </PageCard>
  );
}
