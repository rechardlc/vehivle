import {
  CopyOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeInvisibleOutlined,
  PlusOutlined,
  ReloadOutlined,
  RocketOutlined
} from "@ant-design/icons";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  App,
  Button,
  Col,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Table,
  Tag
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { useMemo, useState } from "react";
import { categoriesApi } from "../api/categories";
import { ImageUploader } from "../components/ImageUploader";
import { PageCard } from "../components/PageCard";
import { vehiclesApi } from "../api/vehicles";
import type { PriceMode, Vehicle, VehicleListItem, VehicleStatus } from "../types";

interface FilterState {
  page: number;
  pageSize: number;
  keyword?: string;
  status?: VehicleStatus;
  categoryId?: string;
}

type VehicleFormValue = Pick<Vehicle, "name" | "categoryId" | "coverMediaId" | "msrpPrice" | "sellingPoints" | "sortOrder"> & {
  priceMode: PriceMode;
};

const statusOptions: Array<{ label: string; value: VehicleStatus }> = [
  { label: "草稿", value: "draft" },
  { label: "已发布", value: "published" },
  { label: "已下架", value: "unpublished" }
];

function renderStatusTag(status: VehicleStatus) {
  if (status === "published") {
    return <Tag color="green">已发布</Tag>;
  }
  if (status === "draft") {
    return <Tag color="blue">草稿</Tag>;
  }
  if (status === "unpublished") {
    return <Tag color="orange">已下架</Tag>;
  }
  return <Tag>已删除</Tag>;
}

export function VehiclesPage() {
  const { message } = App.useApp();
  const queryClient = useQueryClient();
  const [form] = Form.useForm<VehicleFormValue>();
  const [filter, setFilter] = useState<FilterState>({ page: 1, pageSize: 10 });
  /** 递增后用于重置筛选区未受控组件的展示（与 filter 清空同步） */
  const [filterResetKey, setFilterResetKey] = useState(0);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<VehicleListItem | null>(null);

  const vehiclesQuery = useQuery({
    queryKey: ["vehicles", filter],
    queryFn: () => vehiclesApi.list(filter)
  });

  const categoriesQuery = useQuery({
    queryKey: ["categories"],
    queryFn: () => categoriesApi.list()
  });

  const refreshCoreQueries = () => {
    queryClient.invalidateQueries({ queryKey: ["vehicles"] });
    queryClient.invalidateQueries({ queryKey: ["dashboard-summary"] });
    queryClient.invalidateQueries({ queryKey: ["audit-logs"] });
  };

  const createMutation = useMutation({
    mutationFn: vehiclesApi.create,
    onSuccess: () => {
      message.success("车型创建成功");
      refreshCoreQueries();
      closeModal();
    },
    onError: (error: Error) => message.error(error.message)
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: Partial<Vehicle> }) => vehiclesApi.update(id, payload),
    onSuccess: () => {
      message.success("车型更新成功");
      refreshCoreQueries();
      closeModal();
    },
    onError: (error: Error) => message.error(error.message)
  });

  const publishMutation = useMutation({
    mutationFn: ({ id, action }: { id: string; action: "publish" | "unpublish" }) =>
      action === "publish" ? vehiclesApi.publish(id) : vehiclesApi.unpublish(id),
    onSuccess: (_, vars) => {
      message.success(vars.action === "publish" ? "车型发布成功" : "车型下架成功");
      refreshCoreQueries();
    },
    onError: (error: Error) => message.error(error.message)
  });

  const duplicateMutation = useMutation({
    mutationFn: (id: string) => vehiclesApi.duplicate(id),
    onSuccess: () => {
      message.success("车型复制成功");
      refreshCoreQueries();
    },
    onError: (error: Error) => message.error(error.message)
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => vehiclesApi.remove(id),
    onSuccess: () => {
      message.success("车型删除成功");
      refreshCoreQueries();
    },
    onError: (error: Error) => message.error(error.message)
  });

  const batchMutation = useMutation({
    mutationFn: ({ ids, status }: { ids: string[]; status: VehicleStatus }) => vehiclesApi.batchStatus(ids, status),
    onSuccess: () => {
      message.success("批量操作成功");
      setSelectedIds([]);
      refreshCoreQueries();
    },
    onError: (error: Error) => message.error(error.message)
  });

  const categoryOptions = useMemo(
    () => (categoriesQuery.data?.list ?? []).map((item) => ({ label: item.name, value: item.id })),
    [categoriesQuery.data]
  );

  function openCreateModal() {
    setEditing(null);
    setModalOpen(true);
    form.setFieldsValue({
      name: "",
      categoryId: undefined,
      coverMediaId: "",
      msrpPrice: 0,
      priceMode: "msrp",
      sellingPoints: "",
      sortOrder: 1
    });
  }

  function openEditModal(record: VehicleListItem) {
    setEditing(record);
    setModalOpen(true);
    form.setFieldsValue({
      name: record.name,
      categoryId: record.categoryId,
      coverMediaId: record.coverMediaId,
      msrpPrice: record.msrpPrice,
      priceMode: record.priceMode,
      sellingPoints: record.sellingPoints,
      sortOrder: record.sortOrder
    });
  }

  function closeModal() {
    setEditing(null);
    setModalOpen(false);
    form.resetFields();
  }

  function submitForm(values: VehicleFormValue) {
    if (editing) {
      updateMutation.mutate({ id: editing.id, payload: values });
      return;
    }
    createMutation.mutate(values);
  }

  /** 清空筛选条件并重新拉取列表（重置分页与表格勾选） */
  function handleRefreshList() {
    setFilter({ page: 1, pageSize: 10 });
    setFilterResetKey((k) => k + 1);
    setSelectedIds([]);
    void queryClient.invalidateQueries({ queryKey: ["vehicles"] });
  }

  const columns: ColumnsType<VehicleListItem> = [
    { title: "车型名称", dataIndex: "name" },
    { title: "所属分类", dataIndex: "categoryName", width: 160 },
    {
      title: "价格",
      width: 160,
      render: (_, record) => (record.priceMode === "negotiable" ? "面议" : `CNY ${record.msrpPrice}`)
    },
    {
      title: "状态",
      width: 120,
      dataIndex: "status",
      render: (status) => renderStatusTag(status as VehicleStatus)
    },
    {
      title: "操作",
      width: 400,
      render: (_, record) => (
        <Space wrap>
          <Button size="small" icon={<EditOutlined />} onClick={() => openEditModal(record)}>
            编辑
          </Button>
          {record.status === "published" ? (
            <Button
              size="small"
              icon={<EyeInvisibleOutlined />}
              onClick={() => publishMutation.mutate({ id: record.id, action: "unpublish" })}
            >
              下架
            </Button>
          ) : (
            <Button size="small" icon={<RocketOutlined />} onClick={() => publishMutation.mutate({ id: record.id, action: "publish" })}>
              发布
            </Button>
          )}
          <Button size="small" icon={<CopyOutlined />} onClick={() => duplicateMutation.mutate(record.id)}>
            复制
          </Button>
          <Popconfirm title="确定删除该车型吗？" onConfirm={() => deleteMutation.mutate(record.id)}>
            <Button size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  return (
    <PageCard
      title="车型管理"
      extra={
        <Space>
          <Button
            disabled={selectedIds.length === 0}
            onClick={() => batchMutation.mutate({ ids: selectedIds, status: "published" })}
          >
            批量发布
          </Button>
          <Button
            disabled={selectedIds.length === 0}
            onClick={() => batchMutation.mutate({ ids: selectedIds, status: "unpublished" })}
          >
            批量下架
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
            新建车型
          </Button>
        </Space>
      }
    >
      <Space key={filterResetKey} wrap style={{ marginBottom: 16 }}>
        <Input.Search
          allowClear
          placeholder="请输入车型名称"
          style={{ width: 240 }}
          onSearch={(keyword) => setFilter((prev) => ({ ...prev, page: 1, keyword }))}
        />
        <Select
          allowClear
          placeholder="请选择状态"
          style={{ width: 180 }}
          options={statusOptions}
          onChange={(status) => setFilter((prev) => ({ ...prev, page: 1, status }))}
        />
        <Select
          allowClear
          placeholder="请选择分类"
          style={{ width: 220 }}
          options={categoryOptions}
          onChange={(categoryId) => setFilter((prev) => ({ ...prev, page: 1, categoryId }))}
        />
        <Button icon={<ReloadOutlined />} loading={vehiclesQuery.isFetching} onClick={handleRefreshList}>
          重置
        </Button>
      </Space>

      <Table
        rowKey="id"
        loading={vehiclesQuery.isLoading}
        dataSource={vehiclesQuery.data?.list ?? []}
        columns={columns}
        rowSelection={{
          selectedRowKeys: selectedIds,
          onChange: (keys) => setSelectedIds(keys.map(String))
        }}
        pagination={{
          current: filter.page,
          pageSize: filter.pageSize,
          total: vehiclesQuery.data?.total ?? 0,
          showSizeChanger: true,
          onChange: (page, pageSize) => setFilter((prev) => ({ ...prev, page, pageSize }))
        }}
      />

      <Modal
        title={editing ? "编辑车型" : "新建车型"}
        open={modalOpen}
        onCancel={closeModal}
        onOk={() => form.submit()}
        width={720}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form<VehicleFormValue> form={form} layout="vertical" onFinish={submitForm}>
          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item name="name" label="车型名称" rules={[{ required: true, message: "请输入车型名称" }]}>
                <Input placeholder="请输入车型名称" />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item name="categoryId" label="所属分类" rules={[{ required: true, message: "请选择分类" }]}>
                <Select placeholder="请选择所属分类" options={categoryOptions} />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item name="priceMode" label="价格模式" rules={[{ required: true, message: "请选择价格模式" }]}>
                <Select
                  placeholder="请选择价格模式"
                  options={[
                    { label: "建议零售价", value: "msrp" },
                    { label: "面议", value: "negotiable" }
                  ]}
                />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item name="msrpPrice" label="建议零售价">
                <InputNumber style={{ width: "100%" }} min={0} placeholder="请输入建议零售价" />
              </Form.Item>
            </Col>
            <Col xs={24}>
              <Form.Item name="coverMediaId" label="封面图" rules={[{ required: true, message: "请上传封面图" }]}>
                <ImageUploader placeholder="请上传车型封面图" />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item name="sortOrder" label="排序值">
                <InputNumber style={{ width: "100%" }} min={0} placeholder="请输入排序值" />
              </Form.Item>
            </Col>
            <Col xs={24}>
              <Form.Item name="sellingPoints" label="卖点描述">
                <Input.TextArea rows={4} placeholder="请输入卖点描述" />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </PageCard>
  );
}

