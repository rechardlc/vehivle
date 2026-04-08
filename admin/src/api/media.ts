import { http, requestData } from "./client";

export const mediaApi = {
  /**
   * 直传 MinIO：multipart 字段名为 `file`，与后端 `FormFile("file")` 一致。
   * 成功响应 `data` 为对象键字符串，可作为 `coverMediaId`、`defaultShareImage` 等字段存储。
   */
  uploadImage(file: File) {
    const formData = new FormData();
    formData.append("file", file);
    return requestData<string>(http.post("/admin/upload/images", formData));
  }
};
