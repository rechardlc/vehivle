import type {
  AxiosAdapter,
  AxiosHeaders,
  AxiosRequestConfig,
  AxiosResponse,
  InternalAxiosRequestConfig
} from "axios";
import type {
  AdminUser,
  AuditLog,
  Category,
  DashboardSummary,
  ParamTemplate,
  PagedResult,
  PriceMode,
  SystemSettings,
  Vehicle,
  VehicleListItem,
  VehicleStatus
} from "../types";
import { addAuditLog, createToken, db, getCategoryName, nextId } from "./db";

const API_PREFIX = "/api/v1";

function toApiResponse<T>(code: string, message: string, data: T) {
  return {
    code,
    message,
    data,
    request_id: `req_${Math.random().toString(36).slice(2, 10)}`,
    timestamp: new Date().toISOString()
  };
}

function resolveUrl(url = ""): { path: string; query: URLSearchParams } {
  const normalized = url.startsWith(API_PREFIX) ? url.slice(API_PREFIX.length) : url;
  const withSlash = normalized.startsWith("/") ? normalized : `/${normalized}`;
  const [pathname, search = ""] = withSlash.split("?");
  return { path: pathname, query: new URLSearchParams(search) };
}

function parseBody<T>(config: AxiosRequestConfig): T {
  if (!config.data) {
    return {} as T;
  }
  if (typeof config.data === "string") {
    try {
      return JSON.parse(config.data) as T;
    } catch {
      return {} as T;
    }
  }
  return config.data as T;
}

function getAuthUser(config: AxiosRequestConfig): AdminUser | null {
  const headers = config.headers as AxiosHeaders | Record<string, unknown> | undefined;
  const authHeader = headers
    ? String((headers as Record<string, unknown>).Authorization ?? (headers as Record<string, unknown>).authorization ?? "")
    : "";

  if (!authHeader.startsWith("Bearer ")) {
    return null;
  }

  const token = authHeader.slice(7);
  const userId = db.tokenToUser.get(token);
  if (!userId) {
    return null;
  }
  return db.adminUsers.find((item) => item.id === userId) ?? null;
}

function withLatency<T>(payload: T): Promise<T> {
  return new Promise((resolve) => {
    setTimeout(() => resolve(payload), 200);
  });
}

function paginate<T>(items: T[], page: number, pageSize: number): PagedResult<T> {
  const start = (page - 1) * pageSize;
  return {
    list: items.slice(start, start + pageSize),
    total: items.length,
    page,
    pageSize
  };
}

function createResponse<T>(
  config: InternalAxiosRequestConfig,
  status: number,
  code: string,
  message: string,
  data: T
): Promise<AxiosResponse> {
  return withLatency({
    config,
    status,
    statusText: status >= 200 && status < 300 ? "OK" : "ERROR",
    headers: {},
    data: toApiResponse(code, message, data)
  });
}

function nonPasswordUser(user: AdminUser): Omit<AdminUser, "password"> {
  const clone = { ...user } as AdminUser & { password?: string };
  delete clone.password;
  return clone as Omit<AdminUser, "password">;
}

function routeDashboard(): DashboardSummary {
  const activeVehicles = db.vehicles.filter((item) => item.status !== "deleted");
  return {
    vehicleCount: activeVehicles.length,
    publishedCount: activeVehicles.filter((item) => item.status === "published").length,
    draftCount: activeVehicles.filter((item) => item.status === "draft").length,
    categoryCount: db.categories.length,
    latestOperationAt: db.auditLogs[0]?.timestamp ?? new Date().toISOString()
  };
}

function vehicleToListItem(vehicle: Vehicle): VehicleListItem {
  return {
    ...vehicle,
    categoryName: getCategoryName(vehicle.categoryId)
  };
}

function parsePagination(query: URLSearchParams): { page: number; pageSize: number } {
  const page = Number(query.get("page") ?? "1");
  const pageSize = Number(query.get("pageSize") ?? "10");
  return {
    page: Number.isNaN(page) ? 1 : page,
    pageSize: Number.isNaN(pageSize) ? 10 : pageSize
  };
}

function filterVehicles(query: URLSearchParams): VehicleListItem[] {
  const keyword = (query.get("keyword") ?? "").trim().toLowerCase();
  const status = query.get("status");
  const categoryId = query.get("categoryId");

  return db.vehicles
    .filter((item) => item.status !== "deleted")
    .filter((item) => !status || item.status === status)
    .filter((item) => !categoryId || item.categoryId === categoryId)
    .filter((item) => !keyword || item.name.toLowerCase().includes(keyword))
    .sort((a, b) => b.updatedAt.localeCompare(a.updatedAt))
    .map(vehicleToListItem);
}

function canSwitchVehicleStatus(current: VehicleStatus, next: VehicleStatus): boolean {
  if (next === "published") {
    return current === "draft" || current === "unpublished";
  }
  if (next === "unpublished") {
    return current === "published";
  }
  return true;
}

function parseIdFromPath(path: string, pattern: RegExp): string | null {
  const match = path.match(pattern);
  if (!match) {
    return null;
  }
  return match[1] ?? null;
}

export const mockAdapter: AxiosAdapter = async (config) => {
  const { path, query } = resolveUrl(config.url);
  const method = (config.method ?? "GET").toUpperCase();

  if (path === "/admin/auth/login" && method === "POST") {
    const payload = parseBody<{ username: string; password: string }>(config);
    const user = db.adminUsers.find(
      (item) => item.username === payload.username && item.password === payload.password && item.status === "active"
    );
    if (!user) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "A00001", "Invalid username or password.", null);
    }
    const token = createToken(user.id);
    const refreshToken = createToken(user.id);
    user.lastLoginAt = new Date().toISOString();
    addAuditLog(user.id, "LOGIN", "auth", user.id, `User ${user.username} logged in.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", {
      token,
      refreshToken,
      user: nonPasswordUser(user)
    });
  }

  if (path === "/admin/auth/refresh" && method === "POST") {
    const payload = parseBody<{ refreshToken: string }>(config);
    const userId = db.tokenToUser.get(payload.refreshToken);
    if (!userId) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "A00001", "Refresh token invalid.", null);
    }
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", {
      token: createToken(userId)
    });
  }

  const authRequired = path.startsWith("/admin");
  const user = authRequired ? getAuthUser(config) : null;
  if (authRequired && !user) {
    return createResponse(config as InternalAxiosRequestConfig, 401, "A00001", "Token missing or invalid.", null);
  }

  if (path === "/admin/auth/me" && method === "GET" && user) {
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", nonPasswordUser(user));
  }

  if (path === "/admin/dashboard" && method === "GET") {
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", routeDashboard());
  }

  if (path === "/admin/vehicles" && method === "GET") {
    const { page, pageSize } = parsePagination(query);
    const list = filterVehicles(query);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", paginate(list, page, pageSize));
  }

  if (path === "/admin/vehicles" && method === "POST" && user) {
    const payload = parseBody<
      Pick<Vehicle, "categoryId" | "name" | "coverUrl" | "priceMode" | "msrpPrice" | "sellingPoints" | "sortOrder">
    >(config);
    if (!payload.name || !payload.categoryId) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "B00001", "name and categoryId are required.", null);
    }
    const now = new Date().toISOString();
    const vehicle: Vehicle = {
      id: nextId("v"),
      categoryId: payload.categoryId,
      name: payload.name,
      coverUrl: payload.coverUrl ?? "",
      priceMode: payload.priceMode ?? "msrp",
      msrpPrice: payload.msrpPrice ?? 0,
      status: "draft",
      sellingPoints: payload.sellingPoints ?? "",
      sortOrder: payload.sortOrder ?? 0,
      createdAt: now,
      updatedAt: now
    };
    db.vehicles.unshift(vehicle);
    addAuditLog(user.id, "CREATE", "vehicle", vehicle.id, `Created vehicle ${vehicle.name}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", vehicleToListItem(vehicle));
  }

  if (path === "/admin/vehicles/batch-status" && method === "POST" && user) {
    const payload = parseBody<{ ids: string[]; status: VehicleStatus }>(config);
    if (!Array.isArray(payload.ids) || payload.ids.length === 0) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "B00001", "ids are required.", null);
    }
    db.vehicles = db.vehicles.map((item) => {
      if (!payload.ids.includes(item.id)) {
        return item;
      }
      if (!canSwitchVehicleStatus(item.status, payload.status)) {
        return item;
      }
      return { ...item, status: payload.status, updatedAt: new Date().toISOString() };
    });
    addAuditLog(user.id, "BATCH_STATUS", "vehicle", payload.ids.join(","), `Batch status to ${payload.status}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", true);
  }

  if (path === "/admin/categories" && method === "GET") {
    const list = db.categories
      .map((item) => ({
        ...item,
        parentName: item.parentId ? getCategoryName(item.parentId) : "-"
      }))
      .sort((a, b) => a.level - b.level || a.sortOrder - b.sortOrder);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", list);
  }

  if (path === "/admin/categories" && method === "POST" && user) {
    const payload = parseBody<Pick<Category, "name" | "level" | "parentId" | "status" | "sortOrder">>(config);
    if (!payload.name || !payload.level) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "B00001", "name and level are required.", null);
    }
    const category: Category = {
      id: nextId("c"),
      name: payload.name,
      level: payload.level,
      parentId: payload.level === 2 ? payload.parentId ?? null : null,
      status: payload.status ?? "enabled",
      sortOrder: payload.sortOrder ?? 0
    };
    db.categories.push(category);
    addAuditLog(user.id, "CREATE", "category", category.id, `Created category ${category.name}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", category);
  }

  if (path === "/admin/param-templates" && method === "GET") {
    const list = db.paramTemplates.map((item) => ({
      ...item,
      categoryName: getCategoryName(item.categoryId)
    }));
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", list);
  }

  if (path === "/admin/param-templates" && method === "POST" && user) {
    const payload = parseBody<Omit<ParamTemplate, "id">>(config);
    if (!payload.name || !payload.categoryId) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "B00001", "name and categoryId are required.", null);
    }
    const template: ParamTemplate = {
      id: nextId("pt"),
      name: payload.name,
      categoryId: payload.categoryId,
      status: payload.status ?? "enabled",
      items: (payload.items ?? []).map((item, index) => ({
        ...item,
        id: item.id || nextId("pti"),
        sortOrder: item.sortOrder ?? index + 1
      }))
    };
    db.paramTemplates.unshift(template);
    addAuditLog(user.id, "CREATE", "param_template", template.id, `Created template ${template.name}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", template);
  }

  if (path === "/admin/system-settings" && method === "GET") {
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", db.systemSettings);
  }

  if (path === "/admin/system-settings" && method === "PUT" && user) {
    const payload = parseBody<SystemSettings>(config);
    db.systemSettings = {
      ...db.systemSettings,
      ...payload,
      defaultPriceMode: (payload.defaultPriceMode as PriceMode) ?? db.systemSettings.defaultPriceMode,
      updatedAt: new Date().toISOString()
    };
    addAuditLog(user.id, "UPDATE", "system_settings", db.systemSettings.id, "Updated global system settings.");
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", db.systemSettings);
  }

  if (path === "/admin/audit-logs" && method === "GET") {
    const { page, pageSize } = parsePagination(query);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", paginate<AuditLog>(db.auditLogs, page, pageSize));
  }

  if (path === "/admin/media/upload-policy" && method === "POST") {
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", {
      uploadUrl: "https://mock-oss.example.com/upload",
      objectKey: `${Date.now()}-mock.png`,
      expiresIn: 3600
    });
  }

  if (path === "/admin/media/complete" && method === "POST") {
    const payload = parseBody<{ objectKey: string; mimeType: string; size: number }>(config);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", {
      id: nextId("m"),
      storageKey: payload.objectKey,
      mimeType: payload.mimeType,
      fileSize: payload.size
    });
  }

  const vehicleId = parseIdFromPath(path, /^\/admin\/vehicles\/([^/]+)$/);
  if (vehicleId && method === "PUT" && user) {
    const payload = parseBody<Partial<Vehicle>>(config);
    const target = db.vehicles.find((item) => item.id === vehicleId);
    if (!target) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "B00001", "vehicle not found.", null);
    }
    Object.assign(target, payload, { updatedAt: new Date().toISOString() });
    addAuditLog(user.id, "UPDATE", "vehicle", target.id, `Updated vehicle ${target.name}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", vehicleToListItem(target));
  }

  if (vehicleId && method === "DELETE" && user) {
    const target = db.vehicles.find((item) => item.id === vehicleId);
    if (!target) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "B00001", "vehicle not found.", null);
    }
    target.status = "deleted";
    target.updatedAt = new Date().toISOString();
    addAuditLog(user.id, "DELETE", "vehicle", target.id, `Deleted vehicle ${target.name}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", true);
  }

  const publishId = parseIdFromPath(path, /^\/admin\/vehicles\/([^/]+)\/publish$/);
  if (publishId && method === "POST" && user) {
    const target = db.vehicles.find((item) => item.id === publishId);
    if (!target) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "B00001", "vehicle not found.", null);
    }
    if (!canSwitchVehicleStatus(target.status, "published")) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "C00002", "status transition invalid.", null);
    }
    target.status = "published";
    target.updatedAt = new Date().toISOString();
    addAuditLog(user.id, "PUBLISH", "vehicle", target.id, `Published vehicle ${target.name}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", true);
  }

  const unpublishId = parseIdFromPath(path, /^\/admin\/vehicles\/([^/]+)\/unpublish$/);
  if (unpublishId && method === "POST" && user) {
    const target = db.vehicles.find((item) => item.id === unpublishId);
    if (!target) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "B00001", "vehicle not found.", null);
    }
    if (!canSwitchVehicleStatus(target.status, "unpublished")) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "C00002", "status transition invalid.", null);
    }
    target.status = "unpublished";
    target.updatedAt = new Date().toISOString();
    addAuditLog(user.id, "UNPUBLISH", "vehicle", target.id, `Unpublished vehicle ${target.name}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", true);
  }

  const duplicateId = parseIdFromPath(path, /^\/admin\/vehicles\/([^/]+)\/duplicate$/);
  if (duplicateId && method === "POST" && user) {
    const source = db.vehicles.find((item) => item.id === duplicateId);
    if (!source) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "B00001", "vehicle not found.", null);
    }
    const cloned: Vehicle = {
      ...source,
      id: nextId("v"),
      name: `${source.name} Copy`,
      status: "draft",
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
    db.vehicles.unshift(cloned);
    addAuditLog(user.id, "DUPLICATE", "vehicle", source.id, `Duplicated vehicle ${source.name}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", vehicleToListItem(cloned));
  }

  const categoryId = parseIdFromPath(path, /^\/admin\/categories\/([^/]+)$/);
  if (categoryId && method === "PUT" && user) {
    const payload = parseBody<Partial<Category>>(config);
    const target = db.categories.find((item) => item.id === categoryId);
    if (!target) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "B00001", "category not found.", null);
    }
    Object.assign(target, payload);
    if (target.level === 1) {
      target.parentId = null;
    }
    addAuditLog(user.id, "UPDATE", "category", target.id, `Updated category ${target.name}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", target);
  }

  if (categoryId && method === "DELETE" && user) {
    const childCount = db.categories.filter((item) => item.parentId === categoryId).length;
    const vehicleCount = db.vehicles.filter((item) => item.categoryId === categoryId && item.status !== "deleted").length;
    if (childCount > 0 || vehicleCount > 0) {
      return createResponse(
        config as InternalAxiosRequestConfig,
        200,
        "C00002",
        "category has child categories or vehicles, delete denied.",
        null
      );
    }
    db.categories = db.categories.filter((item) => item.id !== categoryId);
    addAuditLog(user.id, "DELETE", "category", categoryId, `Deleted category ${categoryId}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", true);
  }

  const templateId = parseIdFromPath(path, /^\/admin\/param-templates\/([^/]+)$/);
  if (templateId && method === "PUT" && user) {
    const payload = parseBody<Partial<ParamTemplate>>(config);
    const target = db.paramTemplates.find((item) => item.id === templateId);
    if (!target) {
      return createResponse(config as InternalAxiosRequestConfig, 200, "B00001", "param template not found.", null);
    }
    target.name = payload.name ?? target.name;
    target.categoryId = payload.categoryId ?? target.categoryId;
    target.status = payload.status ?? target.status;
    if (payload.items) {
      target.items = payload.items.map((item, index) => ({
        ...item,
        id: item.id || nextId("pti"),
        sortOrder: item.sortOrder ?? index + 1
      }));
    }
    addAuditLog(user.id, "UPDATE", "param_template", target.id, `Updated template ${target.name}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", target);
  }

  if (templateId && method === "DELETE" && user) {
    db.paramTemplates = db.paramTemplates.filter((item) => item.id !== templateId);
    addAuditLog(user.id, "DELETE", "param_template", templateId, `Deleted template ${templateId}.`);
    return createResponse(config as InternalAxiosRequestConfig, 200, "000000", "success", true);
  }

  return createResponse(config as InternalAxiosRequestConfig, 404, "B00001", `Unmocked route: ${method} ${path}`, null);
};
