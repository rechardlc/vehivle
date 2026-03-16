import { useQuery } from "@tanstack/react-query";
import { Table, Tag } from "antd";
import dayjs from "dayjs";
import { useState } from "react";
import { auditLogsApi } from "../api/auditLogs";
import { PageCard } from "../components/PageCard";

export function AuditLogsPage() {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const logsQuery = useQuery({
    queryKey: ["audit-logs", page, pageSize],
    queryFn: () => auditLogsApi.list(page, pageSize)
  });

  return (
    <PageCard title="审计日志">
      <Table
        rowKey="id"
        loading={logsQuery.isLoading}
        dataSource={logsQuery.data?.list ?? []}
        columns={[
          {
            title: "时间",
            dataIndex: "timestamp",
            width: 200,
            render: (timestamp: string) => dayjs(timestamp).format("YYYY-MM-DD HH:mm:ss")
          },
          {
            title: "操作类型",
            dataIndex: "action",
            width: 140,
            render: (action: string) => <Tag color="blue">{action}</Tag>
          },
          {
            title: "操作对象",
            width: 220,
            render: (_, record: { targetType: string; targetId: string }) => `${record.targetType} / ${record.targetId}`
          },
          {
            title: "详情",
            dataIndex: "detail"
          }
        ]}
        pagination={{
          current: page,
          pageSize,
          total: logsQuery.data?.total ?? 0,
          showSizeChanger: true,
          onChange: (nextPage, nextPageSize) => {
            setPage(nextPage);
            setPageSize(nextPageSize);
          }
        }}
      />
    </PageCard>
  );
}
