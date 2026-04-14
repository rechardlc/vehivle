import { DeleteOutlined, InboxOutlined, UploadOutlined } from "@ant-design/icons";
import { App, Button, Image, Space, Typography, Upload } from "antd";
import type { UploadProps } from "antd";
import { useState } from "react";
import { mediaApi } from "../api/media";
import type { VehicleDetailImage } from "../types";

interface MultiImageUploaderProps {
  value?: VehicleDetailImage[];
  onChange?: (value: VehicleDetailImage[]) => void;
  maxCount?: number;
  maxSizeMB?: number;
  readOnly?: boolean;
}

export function MultiImageUploader({
  value = [],
  onChange,
  maxCount = 9,
  maxSizeMB = 2,
  readOnly = false
}: MultiImageUploaderProps) {
  const { message } = App.useApp();
  const [uploading, setUploading] = useState(false);
  const [dragIndex, setDragIndex] = useState<number | null>(null);

  const emit = (next: VehicleDetailImage[]) => {
    onChange?.(next.map((item, index) => ({ ...item, sortOrder: next.length - index })));
  };

  const beforeUpload: UploadProps["beforeUpload"] = (file) => {
    if (value.length >= maxCount) {
      message.error(`详情图最多上传 ${maxCount} 张`);
      return Upload.LIST_IGNORE;
    }
    if (!file.type.startsWith("image/")) {
      message.error("仅支持上传图片文件");
      return Upload.LIST_IGNORE;
    }
    const isLtMax = file.size / 1024 / 1024 <= maxSizeMB;
    if (!isLtMax) {
      message.error(`单张图片不能超过 ${maxSizeMB}MB`);
      return Upload.LIST_IGNORE;
    }
    return true;
  };

  const customRequest: UploadProps["customRequest"] = async (options) => {
    const file = options.file as File;
    try {
      setUploading(true);
      const result = await mediaApi.uploadImage(file);
      emit([...value, { mediaId: result.id, url: result.url }]);
      options.onSuccess?.({ url: result.url, id: result.id });
      message.success("详情图上传成功");
    } catch (error) {
      const err = error as Error;
      message.error(err.message || "详情图上传失败");
      options.onError?.(err);
    } finally {
      setUploading(false);
    }
  };

  const removeAt = (index: number) => {
    emit(value.filter((_, i) => i !== index));
  };

  const moveItem = (from: number, to: number) => {
    if (from === to || from < 0 || to < 0 || from >= value.length || to >= value.length) return;
    const next = [...value];
    const [picked] = next.splice(from, 1);
    next.splice(to, 0, picked);
    emit(next);
  };

  return (
    <Space direction="vertical" size={10} style={{ width: "100%" }}>
      {!readOnly && (
        <Upload accept="image/*" showUploadList={false} beforeUpload={beforeUpload} customRequest={customRequest}>
          <Button icon={<UploadOutlined />} loading={uploading} disabled={value.length >= maxCount}>
            上传详情图
          </Button>
        </Upload>
      )}

      {value.length > 0 ? (
        <div className="detail-image-grid">
          {value.map((item, index) => (
            <div
              className="detail-image-tile"
              key={`${item.mediaId}-${index}`}
              draggable={!readOnly}
              onDragStart={() => setDragIndex(index)}
              onDragOver={(event) => event.preventDefault()}
              onDrop={() => {
                if (dragIndex !== null) moveItem(dragIndex, index);
                setDragIndex(null);
              }}
            >
              <Image
                src={item.url}
                width="100%"
                height={96}
                style={{ objectFit: "cover", borderRadius: 6 }}
                preview={{ mask: "预览" }}
                fallback="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='144' height='96'%3E%3Crect fill='%23f0f0f0' width='100%25' height='100%25'/%3E%3Ctext x='50%25' y='50%25' dominant-baseline='middle' text-anchor='middle' fill='%23999' font-size='12'%3E加载失败%3C/text%3E%3C/svg%3E"
              />
              <div className="detail-image-meta">
                <Typography.Text type="secondary">#{index + 1}</Typography.Text>
                {!readOnly && (
                  <Button size="small" danger type="text" icon={<DeleteOutlined />} onClick={() => removeAt(index)} />
                )}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="detail-image-empty">
          <InboxOutlined />
          <Typography.Text type="secondary">上传 1-9 张详情图，拖拽图片调整展示顺序</Typography.Text>
        </div>
      )}
    </Space>
  );
}
