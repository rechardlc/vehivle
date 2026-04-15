import type { AdminUser } from "../types";

const AUTH_USER_KEY = "vehivle_admin_user";
const AUTH_ACCESS_TOKEN_KEY = "vehivle_admin_access_token";

/**
 * 从 localStorage 读取缓存的用户信息（仅 UI 展示用，不代表 Access Token 有效）。
 * 真实登录状态需通过 /auth/me 校验。
 */
export function getStoredUser(): AdminUser | null {
  const raw = localStorage.getItem(AUTH_USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as AdminUser;
  } catch {
    return null;
  }
}

/**
 * 登录成功或 /auth/me 校验通过后，缓存用户信息供 UI 即时展示。
 */
export function setStoredUser(user: AdminUser): void {
  localStorage.setItem(AUTH_USER_KEY, JSON.stringify(user));
}

export function getStoredAccessToken(): string | null {
  return localStorage.getItem(AUTH_ACCESS_TOKEN_KEY);
}

export function setStoredAccessToken(accessToken: string): void {
  localStorage.setItem(AUTH_ACCESS_TOKEN_KEY, accessToken);
}

/**
 * 清除本地缓存的用户信息与 Access Token（配合服务端 Refresh Token Cookie 清除，完成登出）。
 */
export function clearStoredUser(): void {
  localStorage.removeItem(AUTH_USER_KEY);
  localStorage.removeItem(AUTH_ACCESS_TOKEN_KEY);
}
