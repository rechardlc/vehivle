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
import type { TableProps } from "antd";
import dayjs from "dayjs";
import { useEffect, useMemo, useRef, useState } from "react";
import type { CategoryListParams, CategoryListResponse } from "../api/categories";
import { categoriesApi } from "../api/categories";
import { PageCard } from "../components/PageCard";
import type { Category, CategoryStatus } from "../types";

/** 分类列表默认分页（改此处即可调整首屏条数与起始页） */
const DEFAULT_CATEGORY_LIST_PAGE = 1;
const DEFAULT_CATEGORY_LIST_PAGE_SIZE = 10;
/** 表格每页条数可选项，须与下方 pagination.pageSizeOptions 一致 */
const CATEGORY_LIST_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const;

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
  previousEntries: Array<[readonly unknown[], CategoryListResponse | undefined]>;
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
  /** 与服务端分页对齐；与 DEFAULT_* 初始化一致 */
  const [listPage, setListPage] = useState(DEFAULT_CATEGORY_LIST_PAGE);
  const [listPageSize, setListPageSize] = useState(DEFAULT_CATEGORY_LIST_PAGE_SIZE);
  /** 创建时间列排序：为 null 时不传 sortField/sortOrder，走后端默认序；有值时传 createdAt + asc/desc */
  const [createdAtSortOrder, setCreatedAtSortOrder] = useState<"ascend" | "descend" | null>(null);

  const tableWrapRef = useRef<HTMLDivElement>(null);
  const [tableScrollY, setTableScrollY] = useState<number | undefined>(undefined);

  /**
   * 监听表格容器高度，动态计算 Ant Design Table 的 scroll.y。
   * scroll.y = 容器高度 - 表头高度 - 分页器高度（含 margin），仅让 tbody 滚动。
   */
  useEffect(() => {
    const wrap = tableWrapRef.current;
    if (wrap === null) return undefined;

    const recalc = (): void => {
      const wrapH = wrap.getBoundingClientRect().height;
      const thead = wrap.querySelector<HTMLElement>(".ant-table-header, .ant-table-thead");
      const pager = wrap.querySelector<HTMLElement>(".ant-table-pagination");
      const theadH = thead?.offsetHeight ?? 55;
      const pagerH = pager !== null ? pager.offsetHeight + parseFloat(getComputedStyle(pager).marginTop || "0") + parseFloat(getComputedStyle(pager).marginBottom || "0") : 56;
      const next = Math.max(100, Math.floor(wrapH - theadH - pagerH));
      setTableScrollY((prev) => (prev === next ? prev : next));
    };

    const frame = requestAnimationFrame(recalc);
    const ro = new ResizeObserver(recalc);
    ro.observe(wrap);
    return () => {
      cancelAnimationFrame(frame);
      ro.disconnect();
    };
  }, []);


  /** 与后端 CategoryListParams 对齐（默认不传排序参数） */
  const listParams: CategoryListParams = useMemo(() => {
    const keyword = appliedFilter.keyword?.trim();
    return {
      page: listPage,
      pageSize: listPageSize,
      ...(createdAtSortOrder != null
        ? {
            sortField: "createdAt" as const,
            sortOrder: createdAtSortOrder === "ascend" ? ("asc" as const) : ("desc" as const)
          }
        : {}),
      ...(keyword !== undefined && keyword.length > 0 ? { keyword } : {}),
      ...(appliedFilter.level !== undefined ? { level: appliedFilter.level } : {}),
      ...(appliedFilter.status !== undefined ? { status: appliedFilter.status } : {})
    };
  }, [appliedFilter, listPage, listPageSize, createdAtSortOrder]);

  const listQuery = useQuery({
    queryKey: ["categories", listParams],
    queryFn: () => categoriesApi.list(listParams)
  });

  /** 父级下拉：一级分类全量（pageSize=0 不分页）；仅弹窗打开时才请求 */
  const level1ForParentQuery = useQuery({
    queryKey: ["categories", "level1-parent-options"],
    queryFn: () =>
      categoriesApi.list({ level: 1, page: DEFAULT_CATEGORY_LIST_PAGE, pageSize: 0 }),
    enabled: open
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
      const previousEntries = queryClient.getQueriesData<CategoryListResponse>({ queryKey: ["categories"] });
      const nextStatus = variables.payload.status as CategoryStatus;
      queryClient.setQueriesData<CategoryListResponse>({ queryKey: ["categories"] }, (old) => {
        if (!old) return old;
        return {
          ...old,
          list: old.list.map((c) => (c.id === variables.id ? { ...c, status: nextStatus } : c))
        };
      });
      return { previousEntries, isOptimisticStatusToggle: true };
    },
    onSuccess: (data, variables, context) => {
      if (context?.isOptimisticStatusToggle) {
        queryClient.setQueriesData<CategoryListResponse>({ queryKey: ["categories"] }, (old) => {
          if (!old) return old;
          return {
            ...old,
            list: old.list.map((c) => (c.id === variables.id ? { ...c, ...data } : c))
          };
        });
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

  const level1Options = useMemo(
    () => (level1ForParentQuery.data?.list ?? []).map((item) => ({ label: item.name, value: item.id })),
    [level1ForParentQuery.data]
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

  /** 表头排序 / 翻页：与服务端分页、排序联动 */
  const onTableChange: TableProps<Category>["onChange"] = (pagination, _filters, sorter, extra) => {
    if (extra?.action === "paginate") {
      if (pagination.current != null) {
        setListPage(pagination.current);
      }
      if (pagination.pageSize != null) {
        setListPageSize(pagination.pageSize);
      }
      return;
    }
    if (extra?.action === "sort") {
      const s = Array.isArray(sorter) ? sorter[0] : sorter;
      const colKey = s && "columnKey" in s ? s.columnKey : undefined;
      const field = s && "field" in s ? s.field : undefined;
      const isCreatedAt = colKey === "createdAt" || field === "createdAt";
      if (s && isCreatedAt) {
        setCreatedAtSortOrder(s.order ?? null);
        setListPage(DEFAULT_CATEGORY_LIST_PAGE);
      }
    }
  };

  /** 将草稿条件应用为查询并请求列表 */
  function handleQuery() {
    const keyword = draftFilter.keyword?.trim();
    setListPage(DEFAULT_CATEGORY_LIST_PAGE);
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
    setListPage(DEFAULT_CATEGORY_LIST_PAGE);
    setListPageSize(DEFAULT_CATEGORY_LIST_PAGE_SIZE);
    setCreatedAtSortOrder(null);
    setFilterResetKey((k) => k + 1);

    if (appliedEmpty) {
      void listQuery.refetch();
    }
  }

  return (
    <div className="categories-page">
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

        <div ref={tableWrapRef} className="categories-page-table-wrap">
          <Table<Category>
            rowKey="id"
            loading={listQuery.isLoading}
            dataSource={listQuery.data?.list ?? []}
            onChange={onTableChange}
            scroll={tableScrollY !== undefined ? { y: tableScrollY } : undefined}
            pagination={{
              current: listPage,
              pageSize: listPageSize,
              total: listQuery.data?.page.total ?? 0,
              showSizeChanger: true,
              pageSizeOptions: [...CATEGORY_LIST_PAGE_SIZE_OPTIONS],
              showTotal: (total) => `共 ${total} 条`
            }}
            columns={[
              { title: "分类名称", dataIndex: "name" },
              { title: "层级", dataIndex: "level", render: (level: number) => `L${level}` },
              { title: "父级分类", dataIndex: "parentName" },
              { title: "排序值", dataIndex: "sortOrder", width: 100 },
              {
                title: "创建时间",
                key: "createdAt",
                dataIndex: "createdAt",
                width: 180,
                sorter: true,
                sortOrder: createdAtSortOrder ?? undefined,
                sortDirections: ["descend", "ascend"],
                render: (v: Category["createdAt"]) => (v ? dayjs(v).format("YYYY-MM-DD HH:mm:ss") : "—")
              },
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
        </div>

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
    </div>
  );
}

