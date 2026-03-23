import type { AuthPayload } from "../types";

const AUTH_STORAGE_KEY = "vehivle_admin_auth";

export interface StoredAuthState {
  token: string;
  refreshToken: string;
  user: AuthPayload["user"];
}

/**
 * TODO(登录): 启用 AdminLayout 未登录跳转 /login 后删除此占位与 AdminLayout 中的 `?? LOGIN_BYPASS_GUEST`。
 * 开发期无登录态时用于顶栏展示，避免路由拦截阻断后台。
 */
export const LOGIN_BYPASS_GUEST: StoredAuthState = {
  token: "",
  refreshToken: "",
  user: {
    id: "guest",
    username: "访客",
    role: "operator",
    status: "active",
    lastLoginAt: ""
  }
};

export function getAuthState(): StoredAuthState | null {
  const raw = localStorage.getItem(AUTH_STORAGE_KEY);
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as StoredAuthState;
  } catch {
    return null;
  }
}

export function setAuthState(payload: StoredAuthState): void {
  localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(payload));
}

export function clearAuthState(): void {
  localStorage.removeItem(AUTH_STORAGE_KEY);
}
