import type { ParamTemplate } from "../types";
import { http, requestData } from "./client";

export const paramTemplatesApi = {
  list() {
    return requestData<Array<ParamTemplate & { categoryName: string }>>(http.get("/admin/param-templates"));
  },
  create(payload: Omit<ParamTemplate, "id">) {
    return requestData<ParamTemplate>(http.post("/admin/param-templates", payload));
  },
  update(id: string, payload: Partial<ParamTemplate>) {
    return requestData<ParamTemplate>(http.put(`/admin/param-templates/${id}`, payload));
  },
  remove(id: string) {
    return requestData<boolean>(http.delete(`/admin/param-templates/${id}`));
  }
};
