import type { AdminUser } from "../types";

const AUTH_USER_KEY = "vehivle_admin_user";

/**
 * 从 localStorage 读取缓存的用户信息（仅 UI 展示用，不代表 Cookie 有效）。
 * 真实登录状态以 httpOnly Cookie 为准，需通过 /auth/me 校验。
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

/**
 * 清除本地缓存的用户信息（配合服务端 Cookie 清除，完成登出）。
 */
export function clearStoredUser(): void {
  localStorage.removeItem(AUTH_USER_KEY);
}
