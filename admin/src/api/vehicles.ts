import type { PagedResult, Vehicle, VehicleListItem, VehicleStatus } from "../types";
import { http, requestData } from "./client";

export interface VehicleListQuery {
  page: number;
  pageSize: number;
  keyword?: string;
  status?: VehicleStatus;
  categoryId?: string;
}

export const vehiclesApi = {
  list(query: VehicleListQuery) {
    return requestData<PagedResult<VehicleListItem>>(http.get("/admin/vehicles", { params: query }));
  },
  create(
    payload: Pick<Vehicle, "name" | "categoryId" | "coverUrl" | "priceMode" | "msrpPrice" | "sellingPoints" | "sortOrder">
  ) {
    return requestData<VehicleListItem>(http.post("/admin/vehicles", payload));
  },
  update(id: string, payload: Partial<Vehicle>) {
    return requestData<VehicleListItem>(http.put(`/admin/vehicles/${id}`, payload));
  },
  publish(id: string) {
    return requestData<boolean>(http.post(`/admin/vehicles/${id}/publish`));
  },
  unpublish(id: string) {
    return requestData<boolean>(http.post(`/admin/vehicles/${id}/unpublish`));
  },
  duplicate(id: string) {
    return requestData<VehicleListItem>(http.post(`/admin/vehicles/${id}/duplicate`));
  },
  batchStatus(ids: string[], status: VehicleStatus) {
    return requestData<boolean>(http.post("/admin/vehicles/batch-status", { ids, status }));
  },
  remove(id: string) {
    return requestData<boolean>(http.delete(`/admin/vehicles/${id}`));
  }
};
