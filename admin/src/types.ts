export interface ApiResponse<T> {
  code: string;
  message: string;
  data: T;
  requestId: string;
  timestamp: string;
}

export interface PagedResult<T> {
  list: T[];
  total: number;
  page: number;
  pageSize: number;
}

export type Role = "super_admin" | "editor";

export interface AdminUser {
  id: string;
  username: string;
  role: Role;
}

/**
 * 登录成功响应体（Token 通过 httpOnly Cookie 传输，不在 body 中返回）。
 * expiresIn: Access Token 有效期（秒），前端可用于主动续签倒计时。
 */
export interface LoginResult {
  expiresIn: number;
}

/**
 * /auth/me 返回的当前用户信息。
 */
export type AuthMeResult = AdminUser;

/** 分类启用状态：1=启用 0=禁用（与后端 JSON 数字一致） */
export type CategoryStatus = 0 | 1;
export interface Category {
  id: string;
  parentId: string | null;
  level: 1 | 2;
  name: string;
  status: CategoryStatus;
  sortOrder: number;
  /** ISO 8601，列表接口返回 */
  createdAt: string;
  /** 列表接口由后端填充 */
  parentName?: string;
}

export type FieldType = "text" | "number" | "single";
export interface ParamTemplateItem {
  id: string;
  fieldKey: string;
  fieldName: string;
  fieldType: FieldType;
  unit?: string;
  required: boolean;
  display: boolean;
  sortOrder: number;
}

export interface ParamTemplate {
  id: string;
  name: string;
  categoryId: string;
  status: "enabled" | "disabled";
  items: ParamTemplateItem[];
}

export type VehicleStatus = "draft" | "published" | "unpublished" | "deleted";
export type PriceMode = "msrp" | "negotiable";

export interface Vehicle {
  id: string;
  categoryId: string;
  name: string;
  /** 封面图：media_assets 表主键 UUID（上传接口返回的 id） */
  coverMediaId: string;
  /** 服务端根据 storage_key 拼接的公网地址，仅展示用 */
  coverImageUrl?: string;
  priceMode: PriceMode;
  msrpPrice: number;
  status: VehicleStatus;
  sellingPoints: string;
  sortOrder: number;
  updatedAt: string;
  createdAt: string;
}

export interface VehicleListItem extends Vehicle {
  categoryName: string;
}

export interface SystemSettings {
  id: string;
  companyName: string;
  customerServicePhone: string;
  customerServiceWechat: string;
  defaultPriceMode: PriceMode;
  disclaimerText: string;
  defaultShareTitle: string;
  defaultShareImage: string;
  updatedAt: string;
}

export interface AuditLog {
  id: string;
  adminUserId: string;
  action: string;
  targetType: string;
  targetId: string;
  detail: string;
  timestamp: string;
}

export interface DashboardSummary {
  vehicleCount: number;
  publishedCount: number;
  draftCount: number;
  categoryCount: number;
  latestOperationAt: string;
}
