import type { Category, CategoryStatus } from "../types";
import { http, requestData } from "./client";

/** GET /admin/categories 查询参数，与后端 CategoryListQuery 对齐 */
export interface CategoryListParams {
  keyword?: string;
  level?: Category["level"];
  status?: CategoryStatus;
}

export const categoriesApi = {
  /**
   * 分类列表（GET）。无参或空对象等价于全量列表。
   * 筛选参数建议在页面内由用户点击「查询」后再传入，避免输入/下拉即触发请求。
   * 若页面同时需要「全量列表」与「?level=1」，应在业务层用 `enabled` 或从全量结果推导一级分类，避免重复请求。
   */
  list(params?: CategoryListParams) {
    return requestData<Category[]>(http.get("/admin/categories", { params: params ?? {} }));
  },
  create(payload: Pick<Category, "name" | "level" | "parentId" | "status" | "sortOrder">) {
    return requestData<Category>(http.post("/admin/categories", payload));
  },
  update(id: string, payload: Partial<Category>) {
    return requestData<Category>(http.put(`/admin/categories/${id}`, payload));
  },
  remove(id: string) {
    return requestData<boolean>(http.delete(`/admin/categories/${id}`));
  }
};
