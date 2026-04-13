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
import { paramTemplatesApi } from "../api/paramTemplates";
import { PageCard } from "../components/PageCard";
import type { FieldType, ParamTemplate, ParamTemplateItem } from "../types";

interface TemplateFormValue {
  name: string;
  categoryId: string;
  statusEnabled: boolean;
  items: Array<Omit<ParamTemplateItem, "id"> & { id?: string }>;
}

const fieldTypeOptions: Array<{ label: string; value: FieldType }> = [
  { label: "文本", value: "text" },
  { label: "数值", value: "number" },
  { label: "单选", value: "single" }
];

export function ParamTemplatesPage() {
  const { message } = App.useApp();
  const queryClient = useQueryClient();
  const [form] = Form.useForm<TemplateFormValue>();
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<ParamTemplate | null>(null);

  const templatesQuery = useQuery({
    queryKey: ["param-templates"],
    queryFn: paramTemplatesApi.list
  });

  /** 模板只绑定一级分类（PRD：按一级大类维护参数项；二级为品牌/筛选用） */
  const categoriesQuery = useQuery({
    queryKey: ["categories", "level1"],
    queryFn: () => categoriesApi.list({ level: 1, pageSize: 0 })
  });

  const createMutation = useMutation({
    mutationFn: paramTemplatesApi.create,
    onSuccess: () => {
      message.success("模板创建成功");
      queryClient.invalidateQueries({ queryKey: ["param-templates"] });
      closeModal();
    },
    onError: (error: Error) => message.error(error.message)
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: Partial<ParamTemplate> }) => paramTemplatesApi.update(id, payload),
    onSuccess: () => {
      message.success("模板更新成功");
      queryClient.invalidateQueries({ queryKey: ["param-templates"] });
      closeModal();
    },
    onError: (error: Error) => message.error(error.message)
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => paramTemplatesApi.remove(id),
    onSuccess: () => {
      message.success("模板删除成功");
      queryClient.invalidateQueries({ queryKey: ["param-templates"] });
    },
    onError: (error: Error) => message.error(error.message)
  });

  const categoryOptions = useMemo(
    () =>
      (categoriesQuery.data?.list ?? [])
        .filter((c) => c.level === 1)
        .map((item) => ({ label: item.name, value: item.id })),
    [categoriesQuery.data]
  );

  function closeModal() {
    setOpen(false);
    setEditing(null);
    form.resetFields();
  }

  function openCreateModal() {
    setEditing(null);
    setOpen(true);
    form.setFieldsValue({
      name: "",
      categoryId: undefined,
      statusEnabled: true,
      items: [
        {
          fieldKey: "",
          fieldName: "",
          fieldType: "text",
          unit: "",
          required: true,
          display: true,
          sortOrder: 1
        }
      ]
    });
  }

  function openEditModal(record: ParamTemplate) {
    setEditing(record);
    setOpen(true);
    form.setFieldsValue({
      name: record.name,
      categoryId: record.categoryId,
      statusEnabled: record.status === "enabled",
      items: record.items.map((item) => ({ ...item }))
    });
  }

  function submit(values: TemplateFormValue) {
    const payload: Omit<ParamTemplate, "id"> = {
      name: values.name,
      categoryId: values.categoryId,
      status: values.statusEnabled ? "enabled" : "disabled",
      items: values.items.map((item, index) => ({
        id: item.id ?? "",
        fieldKey: item.fieldKey,
        fieldName: item.fieldName,
        fieldType: item.fieldType,
        unit: item.unit,
        required: item.required,
        display: item.display,
        sortOrder: item.sortOrder ?? index + 1
      }))
    };
    if (editing) {
      updateMutation.mutate({ id: editing.id, payload });
      return;
    }
    createMutation.mutate(payload);
  }

  return (
    <PageCard
      title="参数模板管理"
      extra={
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal}>
          新建模板
        </Button>
      }
    >
      <Table
        rowKey="id"
        loading={templatesQuery.isLoading}
        dataSource={templatesQuery.data ?? []}
        columns={[
          { title: "模板名称", dataIndex: "name" },
          { title: "所属分类", dataIndex: "categoryName" },
          { title: "参数项数量", render: (_, record: ParamTemplate) => record.items.length },
          {
            title: "状态",
            dataIndex: "status",
            render: (status: ParamTemplate["status"]) =>
              status === "enabled" ? <Tag color="green">启用</Tag> : <Tag>停用</Tag>
          },
          {
            title: "操作",
            width: 220,
            render: (_, record: ParamTemplate) => (
              <Space>
                <Button size="small" onClick={() => openEditModal(record)}>
                  编辑
                </Button>
                <Popconfirm title="确定删除该模板吗？" onConfirm={() => deleteMutation.mutate(record.id)}>
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
        title={editing ? "编辑参数模板" : "新建参数模板"}
        width={900}
        open={open}
        onCancel={closeModal}
        onOk={() => form.submit()}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
      >
        <Form form={form} layout="vertical" onFinish={submit}>
          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item name="name" label="模板名称" rules={[{ required: true, message: "请输入模板名称" }]}>
                <Input placeholder="请输入模板名称" />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item
                name="categoryId"
                label="所属一级分类"
                rules={[{ required: true, message: "请选择所属一级分类" }]}
              >
                <Select placeholder="请选择一级分类（车型大类）" options={categoryOptions} />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item name="statusEnabled" label="是否启用" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
          </Row>

          <Form.List name="items">
            {(fields, { add, remove }) => (
              <Space orientation="vertical" style={{ width: "100%" }} size={12}>
                {fields.map((field, index) => (
                  <div className="list-row" key={field.key}>
                    <Form.Item
                      name={[field.name, "fieldKey"]}
                      label={index === 0 ? "字段 Key" : ""}
                      rules={[{ required: true, message: "必填" }]}
                    >
                      <Input placeholder="例如：max_speed" />
                    </Form.Item>
                    <Form.Item
                      name={[field.name, "fieldName"]}
                      label={index === 0 ? "字段名称" : ""}
                      rules={[{ required: true, message: "必填" }]}
                    >
                      <Input placeholder="例如：最高时速" />
                    </Form.Item>
                    <Form.Item
                      name={[field.name, "fieldType"]}
                      label={index === 0 ? "字段类型" : ""}
                      rules={[{ required: true, message: "必填" }]}
                    >
                      <Select placeholder="请选择字段类型" options={fieldTypeOptions} />
                    </Form.Item>
                    <Form.Item name={[field.name, "unit"]} label={index === 0 ? "单位" : ""}>
                      <Input placeholder="例如：km/h" />
                    </Form.Item>
                    <Form.Item name={[field.name, "sortOrder"]} label={index === 0 ? "排序值" : ""}>
                      <InputNumber min={1} style={{ width: "100%" }} placeholder="请输入排序值" />
                    </Form.Item>
                    <Form.Item name={[field.name, "required"]} label={index === 0 ? "必填" : ""} valuePropName="checked">
                      <Switch />
                    </Form.Item>
                    <Form.Item name={[field.name, "display"]} label={index === 0 ? "展示" : ""} valuePropName="checked">
                      <Switch />
                    </Form.Item>
                    <Form.Item label={index === 0 ? "删除" : ""}>
                      <Button danger onClick={() => remove(field.name)}>
                        移除
                      </Button>
                    </Form.Item>
                  </div>
                ))}
                <Button
                  onClick={() =>
                    add({
                      fieldKey: "",
                      fieldName: "",
                      fieldType: "text",
                      unit: "",
                      required: true,
                      display: true,
                      sortOrder: fields.length + 1
                    })
                  }
                >
                  添加参数项
                </Button>
              </Space>
            )}
          </Form.List>
        </Form>
      </Modal>
    </PageCard>
  );
}


