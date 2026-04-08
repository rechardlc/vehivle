import { http, requestData } from "./client";

/** 与后端 upload 成功响应 data 字段一致 */
export interface ImageUploadResult {
  id: string;
  url: string;
  storageKey: string;
}

export const mediaApi = {
  /**
   * 上传图片至 MinIO 并写入 media_assets；multipart 字段名为 `file`。
   * 表单应保存返回的 `id` 作为 coverMediaId。
   */
  uploadImage(file: File): Promise<ImageUploadResult> {
    const formData = new FormData();
    formData.append("file", file);
    return requestData<ImageUploadResult>(http.post("/admin/upload/images", formData));
  }
};
