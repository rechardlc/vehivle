import type { Category } from "../types";
import { http, requestData } from "./client";

export const categoriesApi = {
  list() {
    return requestData<Array<Category & { parentName: string }>>(http.get("/admin/categories"));
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
