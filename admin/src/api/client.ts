import axios, { AxiosError, type AxiosResponse } from "axios";
import type { ApiResponse } from "../types";
import { getAuthState } from "../state/auth";

const DEFAULT_DEV_TIMEOUT_MS = 120_000;
const DEFAULT_PROD_TIMEOUT_MS = 8_000;

/**
 * 解析 VITE_HTTP_TIMEOUT_MS（毫秒）。空或非法时按环境回退：开发 2min、生产 8s。
 */
function resolveHttpTimeoutMs(): number {
  const raw = import.meta.env.VITE_HTTP_TIMEOUT_MS?.trim();
  const fallback = import.meta.env.DEV ? DEFAULT_DEV_TIMEOUT_MS : DEFAULT_PROD_TIMEOUT_MS;
  if (raw === undefined || raw === "") {
    return fallback;
  }
  const n = Number.parseInt(raw, 10);
  if (!Number.isFinite(n) || n <= 0) {
    return fallback;
  }
  return n;
}

/**
 * API 根路径，默认 /api/v1（与 .env.development / .env.production 一致）。
 */
function resolveApiBaseUrl(): string {
  const raw = import.meta.env.VITE_API_BASE_URL?.trim();
  if (raw === undefined || raw === "") {
    return "/api/v1";
  }
  return raw;
}

export const http = axios.create({
  baseURL: resolveApiBaseUrl(),
  timeout: resolveHttpTimeoutMs()
});

http.interceptors.request.use((config) => {
  const auth = getAuthState();
  if (auth?.token) {
    config.headers = config.headers ?? {};
    config.headers.Authorization = `Bearer ${auth.token}`;
  }
  return config;
});

function parseAxiosError(error: AxiosError<ApiResponse<null>>): Error {
  if (error.response?.data?.message) {
    return new Error(error.response.data.message);
  }
  return new Error(error.message || "Request failed");
}

export async function requestData<T>(request: Promise<AxiosResponse<ApiResponse<T>>>): Promise<T> {
  try {
    const response = await request;
    const payload = response.data;
    if (payload.code !== "000000") {
      throw new Error(payload.message || "Business error");
    }
    return payload.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      throw parseAxiosError(error as AxiosError<ApiResponse<null>>);
    }
    throw error;
  }
}
