import type {
  AdminUser,
  AuditLog,
  Category,
  ParamTemplate,
  SystemSettings,
  Vehicle
} from "../types";

let idCounter = 1000;
export function nextId(prefix = ""): string {
  idCounter += 1;
  return `${prefix}${idCounter}`;
}

function nowIso(): string {
  return new Date().toISOString();
}

export const db: {
  adminUsers: AdminUser[];
  categories: Category[];
  paramTemplates: ParamTemplate[];
  vehicles: Vehicle[];
  systemSettings: SystemSettings;
  auditLogs: AuditLog[];
  tokenToUser: Map<string, string>;
} = {
  adminUsers: [
    {
      id: "u1",
      username: "admin",
      password: "admin123",
      role: "super_admin",
      status: "active",
      lastLoginAt: nowIso()
    },
    {
      id: "u2",
      username: "operator",
      password: "operator123",
      role: "operator",
      status: "active",
      lastLoginAt: nowIso()
    }
  ],
  categories: [
    { id: "c1", parentId: null, level: 1, name: "Electric Bike", status: "enabled", sortOrder: 1 },
    { id: "c2", parentId: null, level: 1, name: "Electric Tricycle", status: "enabled", sortOrder: 2 },
    { id: "c3", parentId: "c1", level: 2, name: "City Commuter", status: "enabled", sortOrder: 1 },
    { id: "c4", parentId: "c2", level: 2, name: "Cargo Utility", status: "enabled", sortOrder: 1 }
  ],
  paramTemplates: [
    {
      id: "pt1",
      name: "Bike Standard Template",
      categoryId: "c1",
      status: "enabled",
      items: [
        {
          id: "pti1",
          fieldKey: "motor_power",
          fieldName: "Motor Power",
          fieldType: "number",
          unit: "W",
          required: true,
          display: true,
          sortOrder: 1
        },
        {
          id: "pti2",
          fieldKey: "battery_type",
          fieldName: "Battery Type",
          fieldType: "single",
          required: true,
          display: true,
          sortOrder: 2
        }
      ]
    }
  ],
  vehicles: [
    {
      id: "v1",
      categoryId: "c3",
      name: "Urban Swift 60",
      coverUrl: "https://dummyimage.com/600x400/ddd/333&text=Urban+Swift+60",
      priceMode: "msrp",
      msrpPrice: 3999,
      status: "published",
      sellingPoints: "Long range, compact body, quick charging.",
      sortOrder: 1,
      createdAt: nowIso(),
      updatedAt: nowIso()
    },
    {
      id: "v2",
      categoryId: "c4",
      name: "Cargo Pro 200",
      coverUrl: "https://dummyimage.com/600x400/ccc/333&text=Cargo+Pro+200",
      priceMode: "negotiable",
      msrpPrice: 6899,
      status: "draft",
      sellingPoints: "Heavy load support, reinforced frame.",
      sortOrder: 2,
      createdAt: nowIso(),
      updatedAt: nowIso()
    }
  ],
  systemSettings: {
    id: "s1",
    companyName: "Vehivle Channel Inc.",
    customerServicePhone: "400-800-9000",
    customerServiceWechat: "vehivle_service",
    defaultPriceMode: "msrp",
    disclaimerText: "Vehicle specs are for reference only, subject to final delivery.",
    defaultShareTitle: "Vehivle Digital Showroom",
    defaultShareImage: "https://dummyimage.com/600x400/1976d2/ffffff&text=Vehivle",
    updatedAt: nowIso()
  },
  auditLogs: [
    {
      id: "log1",
      adminUserId: "u1",
      action: "LOGIN",
      targetType: "auth",
      targetId: "u1",
      detail: "Admin user logged in.",
      timestamp: nowIso()
    }
  ],
  tokenToUser: new Map<string, string>()
};

export function getCategoryName(categoryId: string): string {
  return db.categories.find((item) => item.id === categoryId)?.name ?? "Unknown";
}

export function addAuditLog(
  adminUserId: string,
  action: string,
  targetType: string,
  targetId: string,
  detail: string
): void {
  db.auditLogs.unshift({
    id: nextId("log"),
    adminUserId,
    action,
    targetType,
    targetId,
    detail,
    timestamp: nowIso()
  });
}

export function createToken(userId: string): string {
  const token = `mock-token-${userId}-${Date.now()}`;
  db.tokenToUser.set(token, userId);
  return token;
}
