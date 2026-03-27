import { DeleteOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";
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
  Switch,
  Table
} from "antd";
import { useMemo, useState } from "react";
import { categoriesApi } from "../api/categories";
import { PageCard } from "../components/PageCard";
import type { Category, CategoryStatus } from "../types";

type CategoryFormValue = Pick<Category, "name" | "level" | "parentId" | "sortOrder"> & { statusEnabled: boolean };

/** 列表筛选（客户端过滤，与 VehiclesPage 重置行为对齐） */
interface FilterState {
  keyword?: string;
  level?: Category["level"];
  status?: CategoryStatus;
}

const levelFilterOptions: Array<{ label: string; value: Category["level"] }> = [
  { label: "一级分类", value: 1 },
  { label: "二级分类", value: 2 }
];

const statusFilterOptions: Array<{ label: string; value: CategoryStatus }> = [
  { label: "启用", value: 1 },
  { label: "停用", value: 0 }
];

/** updateMutation 乐观更新上下文：失败时用于回滚列表缓存 */
interface CategoryUpdateMutationContext {
  previousList: Category[] | undefined;
  /** 列表内仅改 status：已先写入缓存，失败需回滚 */
  isOptimisticStatusToggle: boolean;
}

export function CategoriesPage() {
  const queryClient = useQueryClient();
  const { message } = App.useApp();
  const [form] = Form.useForm<CategoryFormValue>();
  const [editing, setEditing] = useState<Category | null>(null);
  const [open, setOpen] = useState(false);
  const [filter, setFilter] = useState<FilterState>({});
  /** 递增后用于重置筛选区未受控组件的展示（与 filter 清空同步） */
  const [filterResetKey, setFilterResetKey] = useState(0);

  const listQuery = useQuery({
    queryKey: ["categories"],
    queryFn: categoriesApi.list
  });

  const tableData = useMemo(() => {
    const list = listQuery.data ?? [];
    return list.filter((item) => {
      const kw = filter.keyword?.trim().toLowerCase();
      if (kw !== undefined && kw.length > 0 && !item.name.toLowerCase().includes(kw)) {
        return false;
      }
      if (filter.level !== undefined && item.level !== filter.level) {
        return false;
      }
      if (filter.status !== undefined && item.status !== filter.status) {
        return false;
      }
      return true;
    });
  }, [listQuery.data, filter]);

  const createMutation = useMutation({
    mutationFn: categoriesApi.create,
    onSuccess: () => {
      message.success("分类创建成功");
      queryClient.invalidateQueries({ queryKey: ["categories"] });
      closeModal();
    },
    onError: (error: Error) => message.error(error.message)
  });

  const updateMutation = useMutation({
    mutationFn: ({
      id,
      payload
    }: {
      id: string;
      payload: Partial<Category>;
      /** 为 true 时在成功后关闭编辑弹窗（列表内 Switch 改状态不应关弹窗） */
      closeAfter?: boolean;
    }) => categoriesApi.update(id, payload),
    onMutate: async (variables): Promise<CategoryUpdateMutationContext> => {
      const isOptimisticStatusToggle = variables.closeAfter === false && variables.payload.status !== undefined;
      if (!isOptimisticStatusToggle) {
        return { previousList: undefined, isOptimisticStatusToggle: false };
      }
      await queryClient.cancelQueries({ queryKey: ["categories"] });
      const previousList = queryClient.getQueryData<Category[]>(["categories"]);
      const nextStatus = variables.payload.status as CategoryStatus;
      queryClient.setQueryData<Category[]>(["categories"], (old) =>
        (old ?? []).map((c) => (c.id === variables.id ? { ...c, status: nextStatus } : c))
      );
      return { previousList, isOptimisticStatusToggle: true };
    },
    onSuccess: (data, variables, context) => {
      if (context?.isOptimisticStatusToggle) {
        queryClient.setQueryData<Category[]>(["categories"], (old) =>
          (old ?? []).map((c) => (c.id === variables.id ? { ...c, ...data } : c))
        );
      } else {
        void queryClient.invalidateQueries({ queryKey: ["categories"] });
      }
      message.success("分类更新成功");
      if (variables.closeAfter) {
        closeModal();
      }
    },
    onError: (error: Error, _variables, context) => {
      if (context?.isOptimisticStatusToggle) {
        if (context.previousList !== undefined) {
          queryClient.setQueryData(["categories"], context.previousList);
        } else {
          void queryClient.invalidateQueries({ queryKey: ["categories"] });
        }
      }
      message.error(error.message);
    }
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => categoriesApi.remove(id),
    onSuccess: () => {
      message.success("分类删除成功");
      queryClient.invalidateQueries({ queryKey: ["categories"] });
    },
    onError: (error: Error) => message.error(error.message)
  });

  const level1Options = useMemo(
    () => (listQuery.data ?? []).filter((item) => item.level === 1).map((item) => ({ label: item.name, value: item.id })),
    [listQuery.data]
  );

  function closeModal() {
    setOpen(false);
    setEditing(null);
    form.resetFields();
  }

  function onCreate() {
    setEditing(null);
    setOpen(true);
    form.setFieldsValue({
      level: 1,
      statusEnabled: true,
      sortOrder: 1,
      parentId: null
    });
  }

  function onEdit(record: Category) {
    setEditing(record);
    setOpen(true);
    form.setFieldsValue({
      name: record.name,
      level: record.level,
      parentId: record.parentId,
      sortOrder: record.sortOrder,
      statusEnabled: record.status === 1
    });
  }

  /**
   * 列表内切换启用/停用：仅提交 status，不关闭编辑弹窗。
   * @param record 当前行分类
   * @param enabled 开关是否打开（true=启用）
   */
  function onTableStatusChange(record: Category, enabled: boolean) {
    const nextStatus: CategoryStatus = enabled ? 1 : 0;
    if (record.status === nextStatus) {
      return;
    }
    updateMutation.mutate({ id: record.id, payload: { status: nextStatus }, closeAfter: false });
  }

  function onSubmit(values: CategoryFormValue) {
    const payload: Pick<Category, "name" | "level" | "parentId" | "status" | "sortOrder"> = {
      name: values.name,
      level: values.level,
      parentId: values.level === 2 ? values.parentId : null,
      sortOrder: values.sortOrder,
      status: values.statusEnabled ? 1 : 0
    };
    if (editing) {
      updateMutation.mutate({ id: editing.id, payload, closeAfter: true });
      return;
    }
    createMutation.mutate(payload);
  }

  /** 清空筛选条件并重新拉取列表（与 VehiclesPage 一致） */
  function handleRefreshList() {
    setFilter({});
    setFilterResetKey((k) => k + 1);
    void queryClient.invalidateQueries({ queryKey: ["categories"] });
  }

  return (
    <PageCard
      title="分类管理"
      extra={
        <Button type="primary" icon={<PlusOutlined />} onClick={onCreate}>
          新建分类
        </Button>
      }
    >
      <Space key={filterResetKey} wrap className="categories-page-filter-bar" style={{ marginBottom: 16 }}>
        <Input.Search
          allowClear
          placeholder="请输入分类名称"
          style={{ width: 240 }}
          onSearch={(keyword) => setFilter((prev) => ({ ...prev, keyword }))}
        />
        <Select
          allowClear
          placeholder="请选择层级"
          style={{ width: 160 }}
          options={levelFilterOptions}
          onChange={(level) => setFilter((prev) => ({ ...prev, level }))}
        />
        <Select
          allowClear
          placeholder="请选择状态"
          style={{ width: 160 }}
          options={statusFilterOptions}
          onChange={(status) => setFilter((prev) => ({ ...prev, status }))}
        />
        <Button icon={<ReloadOutlined />} loading={listQuery.isFetching} onClick={handleRefreshList}>
          重置
        </Button>
      </Space>

      <Table
        rowKey="id"
        loading={listQuery.isLoading}
        dataSource={tableData}
        columns={[
          { title: "分类名称", dataIndex: "name" },
          { title: "层级", dataIndex: "level", render: (level: number) => `L${level}` },
          { title: "父级分类", dataIndex: "parentName" },
          { title: "排序值", dataIndex: "sortOrder", width: 100 },
          {
            title: "状态",
            dataIndex: "status",
            width: 140,
            render: (_: Category["status"], record: Category) => (
              <Switch
                checked={record.status === 1}
                checkedChildren="启用"
                unCheckedChildren="停用"
                onChange={(checked) => onTableStatusChange(record, checked)}
              />
            )
          },
          {
            title: "操作",
            width: 220,
            render: (_, record: Category) => (
              <Space>
                <Button size="small" onClick={() => onEdit(record)}>
                  编辑
                </Button>
                <Popconfirm title="确定删除该分类吗？" onConfirm={() => deleteMutation.mutate(record.id)}>
                  <Button size="small" danger icon={<DeleteOutlined />}>
                    删除
                  </Button>
                </Popconfirm>
              </Space>
            )
          }
        ]}
      />

      <Modal
        title={editing ? "编辑分类" : "新建分类"}
        open={open}
        onCancel={closeModal}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        destroyOnHidden
      >
        <Form form={form} layout="vertical" onFinish={onSubmit}>
          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item name="name" label="分类名称" rules={[{ required: true, message: "请输入分类名称" }]}>
                <Input placeholder="请输入分类名称" />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item name="level" label="层级" rules={[{ required: true, message: "请选择层级" }]}>
                <Select
                  placeholder="请选择分类层级"
                  options={[
                    { label: "一级分类", value: 1 },
                    { label: "二级分类", value: 2 }
                  ]}
                />
              </Form.Item>
            </Col>
            <Form.Item noStyle shouldUpdate={(prev, curr) => prev.level !== curr.level}>
              {({ getFieldValue }) =>
                getFieldValue("level") === 2 ? (
                  <Col xs={24} md={12}>
                    <Form.Item name="parentId" label="父级分类" rules={[{ required: true, message: "二级分类必须选择父级" }]}>
                      <Select placeholder="请选择父级分类" options={level1Options} />
                    </Form.Item>
                  </Col>
                ) : null
              }
            </Form.Item>
            <Col xs={24} md={12}>
              <Form.Item name="sortOrder" label="排序值" rules={[{ required: true, message: "请输入排序值" }]}>
                <InputNumber style={{ width: "100%" }} min={0} placeholder="请输入排序值" />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item name="statusEnabled" label="是否启用" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </PageCard>
  );
}

