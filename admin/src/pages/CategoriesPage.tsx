import { DeleteOutlined, PlusOutlined, ReloadOutlined, SearchOutlined } from "@ant-design/icons";
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
import type { CategoryListParams } from "../api/categories";
import { categoriesApi } from "../api/categories";
import { PageCard } from "../components/PageCard";
import type { Category, CategoryStatus } from "../types";

type CategoryFormValue = Pick<Category, "name" | "level" | "parentId" | "sortOrder"> & { statusEnabled: boolean };

/** 列表筛选条件（草稿与已应用共用结构；仅「已应用」驱动列表请求） */
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

/** updateMutation 乐观更新上下文：失败时用于回滚各列表 query 缓存 */
interface CategoryUpdateMutationContext {
  previousEntries: Array<[readonly unknown[], Category[] | undefined]>;
  /** 列表内仅改 status：已先写入缓存，失败需回滚 */
  isOptimisticStatusToggle: boolean;
}

export function CategoriesPage() {
  const queryClient = useQueryClient();
  const { message } = App.useApp();
  const [form] = Form.useForm<CategoryFormValue>();
  const [editing, setEditing] = useState<Category | null>(null);
  const [open, setOpen] = useState(false);
  /** 筛选区草稿：修改不触发列表请求 */
  const [draftFilter, setDraftFilter] = useState<FilterState>({});
  /** 已应用的筛选：仅在此变化时更新 queryKey 并请求后端 */
  const [appliedFilter, setAppliedFilter] = useState<FilterState>({});
  /** 递增后用于重置筛选区未受控组件的展示（与草稿/已应用清空同步） */
  const [filterResetKey, setFilterResetKey] = useState(0);

  /** 与后端 CategoryListParams 对齐；空对象表示全量列表 */
  const listParams: CategoryListParams = useMemo(() => {
    const keyword = appliedFilter.keyword?.trim();
    return {
      ...(keyword !== undefined && keyword.length > 0 ? { keyword } : {}),
      ...(appliedFilter.level !== undefined ? { level: appliedFilter.level } : {}),
      ...(appliedFilter.status !== undefined ? { status: appliedFilter.status } : {})
    };
  }, [appliedFilter]);

  /** 有筛选时主列表可能不含全部一级分类，需单独请求 ?level=1；无筛选时复用主列表数据，避免与「无参全量」重复请求 */
  const hasListFilter = useMemo(() => {
    const keyword = appliedFilter.keyword?.trim();
    return (
      (keyword !== undefined && keyword.length > 0) ||
      appliedFilter.level !== undefined ||
      appliedFilter.status !== undefined
    );
  }, [appliedFilter]);

  const listQuery = useQuery({
    queryKey: ["categories", listParams],
    queryFn: () => categoriesApi.list(listParams)
  });

  /** 仅在有列表筛选时拉取一级分类（父级下拉）；无筛选时父级选项由 listQuery 全量数据推导 */
  const level1ParentsQuery = useQuery({
    queryKey: ["categories", "level1-parents"],
    queryFn: () => categoriesApi.list({ level: 1 }),
    enabled: hasListFilter
  });

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
        return { previousEntries: [], isOptimisticStatusToggle: false };
      }
      await queryClient.cancelQueries({ queryKey: ["categories"] });
      const previousEntries = queryClient.getQueriesData<Category[]>({ queryKey: ["categories"] });
      const nextStatus = variables.payload.status as CategoryStatus;
      queryClient.setQueriesData<Category[]>({ queryKey: ["categories"] }, (old) =>
        (old ?? []).map((c) => (c.id === variables.id ? { ...c, status: nextStatus } : c))
      );
      return { previousEntries, isOptimisticStatusToggle: true };
    },
    onSuccess: (data, variables, context) => {
      if (context?.isOptimisticStatusToggle) {
        queryClient.setQueriesData<Category[]>({ queryKey: ["categories"] }, (old) =>
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
      if (context?.isOptimisticStatusToggle && context.previousEntries !== undefined) {
        for (const [key, data] of context.previousEntries) {
          queryClient.setQueryData(key, data);
        }
      } else if (context?.isOptimisticStatusToggle) {
        void queryClient.invalidateQueries({ queryKey: ["categories"] });
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

  const level1Options = useMemo(() => {
    if (!hasListFilter) {
      return (listQuery.data ?? [])
        .filter((item) => item.level === 1)
        .map((item) => ({ label: item.name, value: item.id }));
    }
    return (level1ParentsQuery.data ?? []).map((item) => ({ label: item.name, value: item.id }));
  }, [hasListFilter, listQuery.data, level1ParentsQuery.data]);

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

  /** 将草稿条件应用为查询并请求列表 */
  function handleQuery() {
    const keyword = draftFilter.keyword?.trim();
    setAppliedFilter({
      ...(keyword !== undefined && keyword.length > 0 ? { keyword } : {}),
      ...(draftFilter.level !== undefined ? { level: draftFilter.level } : {}),
      ...(draftFilter.status !== undefined ? { status: draftFilter.status } : {})
    });
  }

  /**
   * 清空筛选并刷新列表。
   * 不在此调用 `invalidateQueries({ queryKey: ["categories"] })`：会一次性失效主列表、`level1-parents` 等所有前缀匹配的 query，与 setState 触发的 key 变化叠加导致多次请求。
   * 有筛选 → 清空后 queryKey 变化，useQuery 自动拉一次；已无筛选 → 仅显式 refetch 主列表一次。
   */
  function handleRefreshList() {
    const keyword = appliedFilter.keyword?.trim();
    const appliedEmpty =
      (keyword === undefined || keyword.length === 0) &&
      appliedFilter.level === undefined &&
      appliedFilter.status === undefined;

    setDraftFilter({});
    setAppliedFilter({});
    setFilterResetKey((k) => k + 1);

    if (appliedEmpty) {
      void listQuery.refetch();
    }
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
        <Input
          allowClear
          placeholder="请输入分类名称"
          style={{ width: 240 }}
          value={draftFilter.keyword ?? ""}
          onChange={(e) => {
            const keyword = e.target.value;
            setDraftFilter((prev) => ({ ...prev, keyword: keyword === "" ? undefined : keyword }));
          }}
        />
        <Select
          allowClear
          placeholder="请选择层级"
          style={{ width: 160 }}
          options={levelFilterOptions}
          value={draftFilter.level}
          onChange={(level) => setDraftFilter((prev) => ({ ...prev, level }))}
        />
        <Select
          allowClear
          placeholder="请选择状态"
          style={{ width: 160 }}
          options={statusFilterOptions}
          value={draftFilter.status}
          onChange={(status) => setDraftFilter((prev) => ({ ...prev, status }))}
        />
        <Button type="primary" icon={<SearchOutlined />} loading={listQuery.isFetching} onClick={handleQuery}>
          查询
        </Button>
        <Button icon={<ReloadOutlined />} loading={listQuery.isFetching} onClick={handleRefreshList}>
          重置
        </Button>
      </Space>

      <Table
        rowKey="id"
        loading={listQuery.isLoading}
        dataSource={listQuery.data ?? []}
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

