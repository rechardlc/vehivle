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

export type Role = "super_admin" | "operator";
export type UserStatus = "active" | "inactive";

export interface AdminUser {
  id: string;
  username: string;
  password: string;
  role: Role;
  status: UserStatus;
  lastLoginAt: string;
}

export interface AuthPayload {
  token: string;
  refreshToken: string;
  user: Omit<AdminUser, "password">;
}

/** 分类启用状态：1=启用 0=禁用（与后端 JSON 数字一致） */
export type CategoryStatus = 0 | 1;
export interface Category {
  id: string;
  parentId: string | null;
  level: 1 | 2;
  name: string;
  status: CategoryStatus;
  sortOrder: number;
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
  /** 后端封面媒体资源 ID（关联 media_assets） */
  coverMediaId: string;
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
