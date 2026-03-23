import type { PagedResult, PriceMode, Vehicle, VehicleListItem, VehicleStatus } from "../types";
import { http, requestData } from "./client";

export interface VehicleListQuery {
  page: number;
  pageSize: number;
  keyword?: string;
  status?: VehicleStatus;
  categoryId?: string;
}

/**
 * 后端 PriceMode 枚举值与管理端 UI 枚举的映射。
 * 字段名已统一为 camelCase，这里只处理枚举值差异。
 */
function backendPriceToUi(mode: string): PriceMode {
  if (mode === "show_price") return "msrp";
  if (mode === "phone_inquiry") return "negotiable";
  return "negotiable";
}

function uiPriceToBackend(mode: PriceMode): string {
  return mode === "msrp" ? "show_price" : "phone_inquiry";
}

/** 将后端 Vehicle JSON（camelCase）转为管理端 VehicleListItem */
function mapToListItem(v: Vehicle): VehicleListItem {
  return {
    ...v,
    priceMode: backendPriceToUi(v.priceMode as string),
    categoryName: ""
  };
}

export const vehiclesApi = {
  async list(query: VehicleListQuery): Promise<PagedResult<VehicleListItem>> {
    const res = await requestData<{ items: Vehicle[] }>(http.get("/admin/vehicles"));
    let rows = res.items.map(mapToListItem);
    const kw = (query.keyword ?? "").trim().toLowerCase();
    if (kw.length > 0) {
      rows = rows.filter((r) => r.name.toLowerCase().includes(kw));
    }
    if (query.status) {
      rows = rows.filter((r) => r.status === query.status);
    }
    if (query.categoryId) {
      rows = rows.filter((r) => r.categoryId === query.categoryId);
    }
    const total = rows.length;
    const start = (query.page - 1) * query.pageSize;
    const list = rows.slice(start, start + query.pageSize);
    return { list, total, page: query.page, pageSize: query.pageSize };
  },

  async create(
    payload: Pick<Vehicle, "name" | "categoryId" | "coverMediaId" | "priceMode" | "msrpPrice" | "sellingPoints" | "sortOrder">
  ): Promise<VehicleListItem> {
    const body = {
      name: payload.name,
      categoryId: payload.categoryId || null,
      coverMediaId: payload.coverMediaId,
      priceMode: uiPriceToBackend(payload.priceMode),
      msrpPrice: payload.msrpPrice,
      sellingPoints: payload.sellingPoints ?? "",
      sortOrder: payload.sortOrder
    };
    const raw = await requestData<Vehicle>(http.post("/admin/vehicles", body));
    return mapToListItem(raw);
  },

  async update(id: string, payload: Partial<Vehicle>): Promise<VehicleListItem> {
    const body: Record<string, unknown> = {};
    if (payload.name !== undefined) body.name = payload.name;
    if (payload.categoryId !== undefined) body.categoryId = payload.categoryId === "" ? null : payload.categoryId;
    if (payload.coverMediaId !== undefined) body.coverMediaId = payload.coverMediaId;
    if (payload.priceMode !== undefined) body.priceMode = uiPriceToBackend(payload.priceMode);
    if (payload.msrpPrice !== undefined) body.msrpPrice = payload.msrpPrice;
    if (payload.sellingPoints !== undefined) body.sellingPoints = payload.sellingPoints;
    if (payload.sortOrder !== undefined) body.sortOrder = payload.sortOrder;
    const raw = await requestData<Vehicle>(http.put(`/admin/vehicles/${id}`, body));
    return mapToListItem(raw);
  },

  publish(id: string) {
    return requestData<boolean>(http.post(`/admin/vehicles/${id}/publish`));
  },

  unpublish(id: string) {
    return requestData<boolean>(http.post(`/admin/vehicles/${id}/unpublish`));
  },

  async duplicate(id: string): Promise<VehicleListItem> {
    const raw = await requestData<Vehicle>(http.post(`/admin/vehicles/${id}/duplicate`));
    return mapToListItem(raw);
  },

  batchStatus(ids: string[], status: VehicleStatus) {
    return requestData<boolean>(http.post("/admin/vehicles/batch-status", { ids, status }));
  },

  remove(id: string) {
    return requestData<boolean>(http.delete(`/admin/vehicles/${id}`));
  }
};
