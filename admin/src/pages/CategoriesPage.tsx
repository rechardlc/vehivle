import { DeleteOutlined, PlusOutlined } from "@ant-design/icons";
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
  Table,
  Tag
} from "antd";
import { useMemo, useState } from "react";
import { categoriesApi } from "../api/categories";
import { PageCard } from "../components/PageCard";
import type { Category } from "../types";

type CategoryFormValue = Pick<Category, "name" | "level" | "parentId" | "sortOrder"> & { statusEnabled: boolean };

export function CategoriesPage() {
  const queryClient = useQueryClient();
  const { message } = App.useApp();
  const [form] = Form.useForm<CategoryFormValue>();
  const [editing, setEditing] = useState<Category | null>(null);
  const [open, setOpen] = useState(false);

  const listQuery = useQuery({
    queryKey: ["categories"],
    queryFn: categoriesApi.list
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
    mutationFn: ({ id, payload }: { id: string; payload: Partial<Category> }) => categoriesApi.update(id, payload),
    onSuccess: () => {
      message.success("分类更新成功");
      queryClient.invalidateQueries({ queryKey: ["categories"] });
      closeModal();
    },
    onError: (error: Error) => message.error(error.message)
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
      statusEnabled: record.status === "enabled"
    });
  }

  function onSubmit(values: CategoryFormValue) {
    const payload: Pick<Category, "name" | "level" | "parentId" | "status" | "sortOrder"> = {
      name: values.name,
      level: values.level,
      parentId: values.level === 2 ? values.parentId : null,
      sortOrder: values.sortOrder,
      status: values.statusEnabled ? "enabled" : "disabled"
    };
    if (editing) {
      updateMutation.mutate({ id: editing.id, payload });
      return;
    }
    createMutation.mutate(payload);
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
            width: 120,
            render: (status: Category["status"]) =>
              status === "enabled" ? <Tag color="green">启用</Tag> : <Tag color="default">停用</Tag>
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
        destroyOnClose
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

