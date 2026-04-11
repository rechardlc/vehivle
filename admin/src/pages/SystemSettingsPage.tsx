import { useMutation, useQuery } from "@tanstack/react-query";
import { App, Button, Col, Form, Input, Row, Select, Skeleton } from "antd";
import { useEffect } from "react";
import { settingsApi } from "../api/settings";
import { ImageUploader } from "../components/ImageUploader";
import { PageCard } from "../components/PageCard";
import type { SystemSettings } from "../types";

export function SystemSettingsPage() {
  const [form] = Form.useForm<SystemSettings>();
  const { message } = App.useApp();

  const settingsQuery = useQuery({
    queryKey: ["system-settings"],
    queryFn: settingsApi.get
  });

  const isNew = !settingsQuery.data;

  const saveMutation = useMutation({
    mutationFn: (values: Partial<SystemSettings>) =>
      isNew ? settingsApi.create(values) : settingsApi.update(values),
    onSuccess: () => {
      message.success("系统设置已保存");
      settingsQuery.refetch();
    },
    onError: (error: Error) => message.error(error.message)
  });

  useEffect(() => {
    if (!settingsQuery.data) {
      form.setFieldsValue({ defaultPriceMode: "negotiable" });
      return;
    }
    form.setFieldsValue(settingsQuery.data);
  }, [settingsQuery.data, form]);

  return (
    <PageCard title="系统设置" subtitle="基础信息建议双列编辑，文本与图片类字段保留整行">
      {settingsQuery.isLoading ? (
        <Skeleton active />
      ) : (
        <Form form={form} layout="vertical" onFinish={(values) => saveMutation.mutate(values)}>
          <Row gutter={16}>
            <Col xs={24} md={12}>
              <Form.Item name="companyName" label="公司名称" rules={[{ required: true, message: "请输入公司名称" }]}>
                <Input placeholder="请输入公司名称" />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item name="defaultPriceMode" label="默认价格模式" rules={[{ required: true, message: "请选择默认价格模式" }]}>
                <Select
                  placeholder="请选择默认价格模式"
                  options={[
                    { label: "显示零售价", value: "msrp" },
                    { label: "电话询价", value: "negotiable" }
                  ]}
                />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item name="customerServicePhone" label="客服电话">
                <Input placeholder="请输入客服电话" />
              </Form.Item>
            </Col>
            <Col xs={24} md={12}>
              <Form.Item name="customerServiceWechat" label="客服微信">
                <Input placeholder="请输入客服微信" />
              </Form.Item>
            </Col>
            <Col xs={24}>
              <Form.Item name="defaultShareTitle" label="默认分享标题">
                <Input placeholder="请输入默认分享标题" />
              </Form.Item>
            </Col>
            <Col xs={24}>
              <Form.Item name="defaultShareImage" label="默认分享图片">
                <ImageUploader
                  placeholder="请上传默认分享图片"
                  previewFromServer={settingsQuery.data?.defaultShareImageUrl}
                />
              </Form.Item>
            </Col>
            <Col xs={24}>
              <Form.Item name="disclaimerText" label="免责声明">
                <Input.TextArea rows={4} placeholder="请输入免责声明" />
              </Form.Item>
            </Col>
            <Col xs={24}>
              <Button className="pressable" type="primary" htmlType="submit" loading={saveMutation.isPending}>
                {isNew ? "创建配置" : "保存设置"}
              </Button>
            </Col>
          </Row>
        </Form>
      )}
    </PageCard>
  );
}
