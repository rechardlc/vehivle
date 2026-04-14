import type { PagedResult, PriceMode, Vehicle, VehicleDetailImage, VehicleListItem, VehicleStatus } from "../types";
import { http, requestData } from "./client";

export interface VehicleListQuery {
  page: number;
  pageSize: number;
  keyword?: string;
  status?: VehicleStatus;
  categoryId?: string;
  /** 与后端一致：仅支持 createdAt */
  sortField?: "createdAt";
  sortOrder?: "asc" | "desc";
}

/** 与后端 response.ListResult 对齐 */
export interface VehicleListResponse {
  list: Vehicle[];
  page: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
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
function mapToListItem(v: Vehicle & { categoryName?: string }): VehicleListItem {
  return {
    ...v,
    priceMode: backendPriceToUi(v.priceMode as string),
    categoryName: typeof v.categoryName === "string" ? v.categoryName : ""
  };
}

export const vehiclesApi = {
  /**
   * 管理端车型分页列表：查询参数与后端 VehicleListQuery 对齐，响应为 { list, page }。
   */
  async list(query: VehicleListQuery): Promise<PagedResult<VehicleListItem>> {
    const params: Record<string, string | number> = {
      page: query.page,
      pageSize: query.pageSize
    };
    const kw = (query.keyword ?? "").trim();
    if (kw.length > 0) {
      params.keyword = kw;
    }
    if (query.status !== undefined) {
      params.status = query.status;
    }
    if (query.categoryId !== undefined) {
      params.categoryId = query.categoryId;
    }
    if (query.sortField === "createdAt") {
      params.sortField = "createdAt";
      params.sortOrder = query.sortOrder === "asc" ? "asc" : "desc";
    }
    const data = await requestData<VehicleListResponse>(http.get("/admin/vehicles", { params }));
    const list = data.list.map(mapToListItem);
    return {
      list,
      total: data.page.total,
      page: data.page.page,
      pageSize: data.page.pageSize
    };
  },

  async create(
    payload: Pick<Vehicle, "name" | "categoryId" | "coverMediaId" | "priceMode" | "msrpPrice" | "sellingPoints" | "sortOrder"> & {
      detailImages?: VehicleDetailImage[];
    }
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
    if (payload.detailImages !== undefined) {
      await vehiclesApi.saveDetailImages(raw.id, payload.detailImages);
    }
    return mapToListItem(raw);
  },

  async update(id: string, payload: Partial<Vehicle> & { detailImages?: VehicleDetailImage[] }): Promise<VehicleListItem> {
    const body: Record<string, unknown> = {};
    if (payload.name !== undefined) body.name = payload.name;
    if (payload.categoryId !== undefined) body.categoryId = payload.categoryId === "" ? null : payload.categoryId;
    if (payload.coverMediaId !== undefined) body.coverMediaId = payload.coverMediaId;
    if (payload.priceMode !== undefined) body.priceMode = uiPriceToBackend(payload.priceMode);
    if (payload.msrpPrice !== undefined) body.msrpPrice = payload.msrpPrice;
    if (payload.sellingPoints !== undefined) body.sellingPoints = payload.sellingPoints;
    if (payload.sortOrder !== undefined) body.sortOrder = payload.sortOrder;
    const raw = await requestData<Vehicle>(http.put(`/admin/vehicles/${id}`, body));
    if (payload.detailImages !== undefined) {
      await vehiclesApi.saveDetailImages(id, payload.detailImages);
    }
    return mapToListItem(raw);
  },

  detailImages(id: string) {
    return requestData<VehicleDetailImage[]>(http.get(`/admin/vehicles/${id}/detail-images`));
  },

  saveDetailImages(id: string, images: VehicleDetailImage[]) {
    return requestData<VehicleDetailImage[]>(
      http.put(`/admin/vehicles/${id}/detail-images`, {
        images: images.map((image) => ({ mediaId: image.mediaId }))
      })
    );
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
