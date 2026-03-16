import type { AuthPayload } from "../types";

const AUTH_STORAGE_KEY = "vehivle_admin_auth";

export interface StoredAuthState {
  token: string;
  refreshToken: string;
  user: AuthPayload["user"];
}

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
