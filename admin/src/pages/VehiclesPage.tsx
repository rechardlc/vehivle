import {
  CopyOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
  PlusOutlined,
  ReloadOutlined,
  RocketOutlined,
  SearchOutlined
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
import type { TableProps } from "antd";
import dayjs from "dayjs";
import { useEffect, useMemo, useRef, useState } from "react";
import { categoriesApi } from "../api/categories";
import type { VehicleListQuery } from "../api/vehicles";
import { vehiclesApi } from "../api/vehicles";
import { ImageUploader } from "../components/ImageUploader";
import { MultiImageUploader } from "../components/MultiImageUploader";
import { PageCard } from "../components/PageCard";
import type { PriceMode, Vehicle, VehicleDetailImage, VehicleListItem, VehicleStatus } from "../types";

/** 车型列表默认分页（改此处即可调整首屏条数与起始页） */
const DEFAULT_VEHICLE_LIST_PAGE = 1;
const DEFAULT_VEHICLE_LIST_PAGE_SIZE = 10;
/** 表格每页条数可选项，须与下方 pagination.pageSizeOptions 一致 */
const VEHICLE_LIST_PAGE_SIZE_OPTIONS = [10, 20, 50, 100] as const;

interface FilterState {
  keyword?: string;
  status?: VehicleStatus;
  categoryId?: string;
}

type VehicleFormValue = Pick<Vehicle, "name" | "categoryId" | "coverMediaId" | "msrpPrice" | "sellingPoints" | "sortOrder"> & {
  priceMode: PriceMode;
  detailImages: VehicleDetailImage[];
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
  /** 筛选区草稿：修改不触发列表请求 */
  const [draftFilter, setDraftFilter] = useState<FilterState>({});
  /** 已应用的筛选：仅在此变化时更新 queryKey 并请求 */
  const [appliedFilter, setAppliedFilter] = useState<FilterState>({});
  /** 递增后用于重置筛选区未受控组件的展示（与草稿/已应用清空同步） */
  const [filterResetKey, setFilterResetKey] = useState(0);
  /** 与服务端分页对齐；与 DEFAULT_* 初始化一致 */
  const [listPage, setListPage] = useState(DEFAULT_VEHICLE_LIST_PAGE);
  const [listPageSize, setListPageSize] = useState(DEFAULT_VEHICLE_LIST_PAGE_SIZE);
  /** 创建时间列排序：为 null 时不传 sortField/sortOrder；有值时传 createdAt + asc/desc */
  const [createdAtSortOrder, setCreatedAtSortOrder] = useState<"ascend" | "descend" | null>(null);

  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<VehicleListItem | null>(null);
  const [detailImagesLoading, setDetailImagesLoading] = useState(false);

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
      const pagerH =
        pager !== null
          ? pager.offsetHeight +
            parseFloat(getComputedStyle(pager).marginTop || "0") +
            parseFloat(getComputedStyle(pager).marginBottom || "0")
          : 56;
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

  /** 与 VehicleListQuery 对齐 */
  const listParams: VehicleListQuery = useMemo(() => {
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
      ...(appliedFilter.status !== undefined ? { status: appliedFilter.status } : {}),
      ...(appliedFilter.categoryId !== undefined ? { categoryId: appliedFilter.categoryId } : {})
    };
  }, [appliedFilter, listPage, listPageSize, createdAtSortOrder]);

  const vehiclesQuery = useQuery({
    queryKey: ["vehicles", listParams],
    queryFn: () => vehiclesApi.list(listParams)
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
    mutationFn: ({ id, payload }: { id: string; payload: Partial<Vehicle> & { detailImages?: VehicleDetailImage[] } }) =>
      vehiclesApi.update(id, payload),
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

  function closeModal() {
    setEditing(null);
    setModalOpen(false);
    form.resetFields();
  }

  /** 已发布、已删除：弹窗只读查看 */
  const isModalReadOnly =
    editing !== null && (editing.status === "published" || editing.status === "deleted");

  function populateVehicleForm(record: VehicleListItem) {
    form.setFieldsValue({
      name: record.name,
      categoryId: record.categoryId || undefined,
      coverMediaId: record.coverMediaId,
      msrpPrice: record.msrpPrice,
      priceMode: record.priceMode,
      sellingPoints: record.sellingPoints,
      sortOrder: record.sortOrder,
      detailImages: []
    });
  }

  async function loadDetailImages(record: VehicleListItem) {
    setDetailImagesLoading(true);
    try {
      const detailImages = await vehiclesApi.detailImages(record.id);
      form.setFieldValue("detailImages", detailImages);
    } catch (error) {
      message.error((error as Error).message || "详情图加载失败");
    } finally {
      setDetailImagesLoading(false);
    }
  }

  function onCreate() {
    setEditing(null);
    setModalOpen(true);
    form.setFieldsValue({
      name: "",
      categoryId: undefined,
      coverMediaId: "",
      msrpPrice: 0,
      priceMode: "msrp",
      sellingPoints: "",
      sortOrder: 1,
      detailImages: []
    });
  }

  /** 草稿 / 已下架：可编辑 */
  function onEdit(record: VehicleListItem) {
    setEditing(record);
    setModalOpen(true);
    populateVehicleForm(record);
    void loadDetailImages(record);
  }

  /** 已发布 / 已删除：仅查看 */
  function onView(record: VehicleListItem) {
    setEditing(record);
    setModalOpen(true);
    populateVehicleForm(record);
    void loadDetailImages(record);
  }

  function submitForm(values: VehicleFormValue) {
    if (isModalReadOnly) {
      return;
    }
    if (editing) {
      updateMutation.mutate({ id: editing.id, payload: values });
      return;
    }
    createMutation.mutate(values);
  }

  /** 将草稿条件应用为查询并请求列表 */
  function handleQuery() {
    const keyword = draftFilter.keyword?.trim();
    setListPage(DEFAULT_VEHICLE_LIST_PAGE);
    setAppliedFilter({
      ...(keyword !== undefined && keyword.length > 0 ? { keyword } : {}),
      ...(draftFilter.status !== undefined ? { status: draftFilter.status } : {}),
      ...(draftFilter.categoryId !== undefined ? { categoryId: draftFilter.categoryId } : {})
    });
  }

  /**
   * 清空筛选并刷新列表。
   * 有筛选 → 清空后 queryKey 变化，useQuery 自动拉一次；已无筛选 → 仅显式 refetch 主列表一次。
   */
  function handleRefreshList() {
    const keyword = appliedFilter.keyword?.trim();
    const appliedEmpty =
      (keyword === undefined || keyword.length === 0) &&
      appliedFilter.status === undefined &&
      appliedFilter.categoryId === undefined;

    setDraftFilter({});
    setAppliedFilter({});
    setListPage(DEFAULT_VEHICLE_LIST_PAGE);
    setListPageSize(DEFAULT_VEHICLE_LIST_PAGE_SIZE);
    setCreatedAtSortOrder(null);
    setFilterResetKey((k) => k + 1);
    setSelectedIds([]);

    if (appliedEmpty) {
      void vehiclesQuery.refetch();
    }
  }

  /** 表头排序 / 翻页：与列表分页、排序联动 */
  const onTableChange: TableProps<VehicleListItem>["onChange"] = (pagination, _filters, sorter, extra) => {
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
        setListPage(DEFAULT_VEHICLE_LIST_PAGE);
      }
    }
  };

  return (
    <div className="vehicles-page">
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
            <Button type="primary" icon={<PlusOutlined />} onClick={onCreate}>
              新建车型
            </Button>
          </Space>
        }
      >
        <Space key={filterResetKey} wrap className="vehicles-page-filter-bar" style={{ marginBottom: 16 }}>
          <Input
            allowClear
            placeholder="请输入车型名称"
            style={{ width: 240 }}
            value={draftFilter.keyword ?? ""}
            onChange={(e) => {
              const keyword = e.target.value;
              setDraftFilter((prev) => ({ ...prev, keyword: keyword === "" ? undefined : keyword }));
            }}
          />
          <Select
            allowClear
            placeholder="请选择状态"
            style={{ width: 180 }}
            options={statusOptions}
            value={draftFilter.status}
            onChange={(status) => setDraftFilter((prev) => ({ ...prev, status }))}
          />
          <Select
            allowClear
            placeholder="请选择分类"
            style={{ width: 220 }}
            options={categoryOptions}
            value={draftFilter.categoryId}
            onChange={(categoryId) => setDraftFilter((prev) => ({ ...prev, categoryId }))}
          />
          <Button type="primary" icon={<SearchOutlined />} loading={vehiclesQuery.isFetching} onClick={handleQuery}>
            查询
          </Button>
          <Button icon={<ReloadOutlined />} loading={vehiclesQuery.isFetching} onClick={handleRefreshList}>
            重置
          </Button>
        </Space>

        <div ref={tableWrapRef} className="vehicles-page-table-wrap">
          <Table<VehicleListItem>
            rowKey="id"
            loading={vehiclesQuery.isLoading}
            dataSource={vehiclesQuery.data?.list ?? []}
            onChange={onTableChange}
            scroll={tableScrollY !== undefined ? { x: "max-content", y: tableScrollY } : { x: "max-content" }}
            rowSelection={{
              selectedRowKeys: selectedIds,
              onChange: (keys) => setSelectedIds(keys.map(String))
            }}
            pagination={{
              current: listPage,
              pageSize: listPageSize,
              total: vehiclesQuery.data?.total ?? 0,
              showSizeChanger: true,
              pageSizeOptions: [...VEHICLE_LIST_PAGE_SIZE_OPTIONS],
              showTotal: (total) => `共 ${total} 条`
            }}
            columns={[
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
                title: "创建时间",
                key: "createdAt",
                dataIndex: "createdAt",
                width: 180,
                sorter: true,
                sortOrder: createdAtSortOrder ?? undefined,
                sortDirections: ["descend", "ascend"],
                render: (v: Vehicle["createdAt"]) => (v ? dayjs(v).format("YYYY-MM-DD HH:mm:ss") : "—")
              },
              {
                title: "操作",
                key: "actions",
                width: 360,
                fixed: "right",
                render: (_, record) => {
                  if (record.status === "published") {
                    return (
                      <Space wrap>
                        <Button size="small" type="link" icon={<EyeOutlined />} onClick={() => onView(record)}>
                          查看
                        </Button>
                        <Button
                          size="small"
                          icon={<EyeInvisibleOutlined />}
                          onClick={() => publishMutation.mutate({ id: record.id, action: "unpublish" })}
                        >
                          下架
                        </Button>
                        <Button size="small" icon={<CopyOutlined />} onClick={() => duplicateMutation.mutate(record.id)}>
                          复制
                        </Button>
                      </Space>
                    );
                  }
                  if (record.status === "deleted") {
                    return (
                      <Button size="small" type="link" icon={<EyeOutlined />} onClick={() => onView(record)}>
                        查看
                      </Button>
                    );
                  }
                  return (
                    <Space wrap>
                      <Button size="small" icon={<EditOutlined />} onClick={() => onEdit(record)}>
                        编辑
                      </Button>
                      <Button
                        size="small"
                        icon={<RocketOutlined />}
                        onClick={() => publishMutation.mutate({ id: record.id, action: "publish" })}
                      >
                        发布
                      </Button>
                      <Button size="small" icon={<CopyOutlined />} onClick={() => duplicateMutation.mutate(record.id)}>
                        复制
                      </Button>
                      <Popconfirm title="确定删除该车型吗？" onConfirm={() => deleteMutation.mutate(record.id)}>
                        <Button size="small" danger icon={<DeleteOutlined />}>
                          删除
                        </Button>
                      </Popconfirm>
                    </Space>
                  );
                }
              }
            ]}
          />
        </div>

        <Modal
          title={
            editing === null ? "新建车型" : isModalReadOnly ? "查看车型" : "编辑车型"
          }
          open={modalOpen}
          onCancel={closeModal}
          onOk={() => {
            if (isModalReadOnly) {
              closeModal();
              return;
            }
            void form.submit();
          }}
          okText={isModalReadOnly ? "关闭" : "确定"}
          cancelButtonProps={{ style: { display: isModalReadOnly ? "none" : undefined } }}
          width={720}
          confirmLoading={!isModalReadOnly && (createMutation.isPending || updateMutation.isPending || detailImagesLoading)}
          destroyOnHidden
        >
          <Form<VehicleFormValue> form={form} layout="vertical" disabled={isModalReadOnly} onFinish={submitForm}>
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
                  <ImageUploader
                    readOnly={isModalReadOnly}
                    placeholder="请上传车型封面图"
                    previewFromServer={editing?.coverImageUrl}
                  />
                </Form.Item>
              </Col>
              <Col xs={24}>
                <Form.Item
                  name="detailImages"
                  label="详情图集"
                  extra="用于详情页轮播，最多 9 张；单张建议不超过 2MB，拖拽图片可调整顺序。"
                  rules={[
                    {
                      validator: async (_, value: VehicleDetailImage[] | undefined) => {
                        if (value !== undefined && value.length > 9) {
                          throw new Error("详情图最多上传 9 张");
                        }
                      }
                    }
                  ]}
                >
                  <MultiImageUploader readOnly={isModalReadOnly || detailImagesLoading} />
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
    </div>
  );
}
