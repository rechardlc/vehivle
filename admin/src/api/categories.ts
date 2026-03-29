import type { Category, CategoryStatus } from "../types";
import { http, requestData } from "./client";

/** 与后端 response.PageResult 对齐 */
export interface CategoryListPage {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

/** 与后端 response.ListResult[Category] 对齐：data 为 { list, page } */
export interface CategoryListResponse {
  list: Category[];
  page: CategoryListPage;
}

/** GET /admin/categories 查询参数，与后端 CategoryListQuery 对齐 */
export interface CategoryListParams {
  keyword?: string;
  level?: Category["level"];
  status?: CategoryStatus;
  /** 页码，从 1 起；不传则后端默认 */
  page?: number;
  /** 每页条数；不传为 0，后端表示不分页（返回全部） */
  pageSize?: number;
  /** 排序字段，仅支持 createdAt */
  sortField?: "createdAt";
  /** asc | desc，与 sortField 同时传；createdAt 时缺省为 desc */
  sortOrder?: "asc" | "desc";
}

export const categoriesApi = {
  /**
   * 分类列表（GET）。无参或空对象等价于全量列表（pageSize=0 不分页）。
   * 筛选参数建议在页面内由用户点击「查询」后再传入，避免输入/下拉即触发请求。
   * 若页面同时需要「全量列表」与「?level=1」，应在业务层用 `enabled` 或从全量结果推导一级分类，避免重复请求。
   */
  list(params?: CategoryListParams) {
    return requestData<CategoryListResponse>(http.get("/admin/categories", { params: params ?? {} }));
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
