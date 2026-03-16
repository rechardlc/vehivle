import { UploadOutlined } from "@ant-design/icons";
import { App, Button, Image, Space, Typography, Upload } from "antd";
import type { UploadProps } from "antd";
import { useState } from "react";
import { mediaApi } from "../api/media";

interface ImageUploaderProps {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  maxSizeMB?: number;
}

function fileToDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result ?? ""));
    reader.onerror = () => reject(new Error("读取图片失败"));
    reader.readAsDataURL(file);
  });
}

export function ImageUploader({
  value,
  onChange,
  placeholder = "请上传图片",
  maxSizeMB = 5
}: ImageUploaderProps) {
  const { message } = App.useApp();
  const [uploading, setUploading] = useState(false);

  const beforeUpload: UploadProps["beforeUpload"] = (file) => {
    if (!file.type.startsWith("image/")) {
      message.error("仅支持上传图片文件");
      return Upload.LIST_IGNORE;
    }
    const isLtMax = file.size / 1024 / 1024 <= maxSizeMB;
    if (!isLtMax) {
      message.error(`图片大小不能超过 ${maxSizeMB}MB`);
      return Upload.LIST_IGNORE;
    }
    return true;
  };

  const customRequest: UploadProps["customRequest"] = async (options) => {
    const file = options.file as File;
    try {
      setUploading(true);
      const policy = await mediaApi.uploadPolicy({
        filename: file.name,
        mimeType: file.type || "image/png",
        size: file.size
      });
      await mediaApi.complete({
        objectKey: policy.objectKey,
        mimeType: file.type || "image/png",
        size: file.size
      });
      const dataUrl = await fileToDataUrl(file);
      onChange?.(dataUrl);
      options.onSuccess?.({ url: dataUrl });
      message.success("图片上传成功");
    } catch (error) {
      const err = error as Error;
      message.error(err.message || "图片上传失败");
      options.onError?.(err);
    } finally {
      setUploading(false);
    }
  };

  return (
    <Space direction="vertical" size={8}>
      <Upload accept="image/*" showUploadList={false} beforeUpload={beforeUpload} customRequest={customRequest}>
        <Button icon={<UploadOutlined />} loading={uploading}>
          {value ? "重新上传" : "上传图片"}
        </Button>
      </Upload>
      {value ? (
        <Image
          src={value}
          width={160}
          height={120}
          style={{ objectFit: "cover", borderRadius: 6, border: "1px solid #f0f0f0" }}
          preview={{ mask: "预览" }}
        />
      ) : (
        <Typography.Text type="secondary">{placeholder}</Typography.Text>
      )}
    </Space>
  );
}
