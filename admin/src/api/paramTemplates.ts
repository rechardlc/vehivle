import type {
  ParamTemplate,
  ParamTemplateBackendFieldType,
  ParamTemplateListParams,
  ParamTemplateListResponse,
  ParamTemplatePayload,
  ParamTemplateStatus,
  ParamTemplateUi,
  ParamTemplateUiFieldType,
  ParamTemplateUiItem,
  ParamTemplateUiStatus
} from "../types";
import { http, requestData } from "./client";

const DEFAULT_LIST_PARAMS: Required<ParamTemplateListParams> = {
  page: 1,
  pageSize: 100
};

function backendStatusToUi(status: ParamTemplateStatus): ParamTemplateUiStatus {
  return status === 1 ? "enabled" : "disabled";
}

function uiStatusToBackend(status: ParamTemplateUiStatus): ParamTemplateStatus {
  return status === "enabled" ? 1 : 0;
}

function backendFieldTypeToUi(fieldType: ParamTemplateBackendFieldType): ParamTemplateUiFieldType {
  return fieldType === "single_select" ? "single" : fieldType;
}

function uiFieldTypeToBackend(fieldType: ParamTemplateUiFieldType): ParamTemplateBackendFieldType {
  return fieldType === "single" ? "single_select" : fieldType;
}

function flagToBoolean(value: 0 | 1 | undefined): boolean {
  return value === 1;
}

function booleanToFlag(value: boolean): 0 | 1 {
  return value ? 1 : 0;
}

function normalizeOptionalString(value: string | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed === "" ? undefined : trimmed;
}

function mapItemToUi(item: ParamTemplate["items"][number]): ParamTemplateUiItem {
  return {
    id: item.id,
    templateId: item.templateId,
    fieldKey: item.fieldKey,
    fieldName: item.fieldName,
    fieldType: backendFieldTypeToUi(item.fieldType),
    unit: item.unit ?? "",
    required: flagToBoolean(item.required),
    display: flagToBoolean(item.display),
    sortOrder: item.sortOrder
  };
}

function mapTemplateToUi(template: ParamTemplate & { itemNum?: number; categoryName?: string }): ParamTemplateUi {
  return {
    id: template.id,
    name: template.name,
    categoryId: template.categoryId,
    status: backendStatusToUi(template.status),
    items: (template.items ?? []).map(mapItemToUi),
    itemNum: template.itemNum ?? template.items?.length ?? 0,
    categoryName: template.categoryName,
    createdAt: template.createdAt,
    updatedAt: template.updatedAt
  };
}

function mapItemToPayload(item: ParamTemplateUiItem): ParamTemplatePayload["items"][number] {
  return {
    id: normalizeOptionalString(item.id),
    templateId: normalizeOptionalString(item.templateId),
    fieldKey: item.fieldKey,
    fieldName: item.fieldName,
    fieldType: uiFieldTypeToBackend(item.fieldType),
    unit: normalizeOptionalString(item.unit) ?? null,
    required: booleanToFlag(item.required),
    display: booleanToFlag(item.display),
    sortOrder: item.sortOrder
  };
}

function mapPayloadToBackend(payload: Omit<ParamTemplateUi, "id" | "itemNum">): ParamTemplatePayload {
  return {
    name: payload.name,
    categoryId: payload.categoryId,
    status: uiStatusToBackend(payload.status),
    items: payload.items.map(mapItemToPayload)
  };
}

export const paramTemplatesApi = {
  async list(params?: ParamTemplateListParams): Promise<ParamTemplateUi[]> {
    const query = {
      page: params?.page ?? DEFAULT_LIST_PARAMS.page,
      pageSize: params?.pageSize ?? DEFAULT_LIST_PARAMS.pageSize
    };
    const data = await requestData<ParamTemplateListResponse>(http.get("/admin/param-templates/list", { params: query }));
    return (data.list ?? []).map(mapTemplateToUi);
  },
  async getItemsById(id: string): Promise<ParamTemplateUi> {
    const data = await requestData<ParamTemplate>(http.get(`/admin/param-templates/getItemsById/${id}`));
    return mapTemplateToUi(data);
  },
  async create(payload: Omit<ParamTemplateUi, "id" | "itemNum">): Promise<ParamTemplateUi> {
    const data = await requestData<ParamTemplate>(http.post("/admin/param-templates", mapPayloadToBackend(payload)));
    return mapTemplateToUi(data);
  },
  async update(id: string, payload: Omit<ParamTemplateUi, "id" | "itemNum">): Promise<ParamTemplateUi> {
    const data = await requestData<ParamTemplate>(http.put(`/admin/param-templates/${id}`, mapPayloadToBackend(payload)));
    return mapTemplateToUi(data);
  },
  remove(id: string) {
    return requestData<string>(http.delete(`/admin/param-templates/${id}`));
  }
};
