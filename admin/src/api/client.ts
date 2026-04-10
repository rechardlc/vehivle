import axios, { type AxiosError, type AxiosResponse } from "axios";
import type { ApiResponse } from "../types";
import { clearStoredUser } from "../state/auth";

const DEFAULT_DEV_TIMEOUT_MS = 120_000;
const DEFAULT_PROD_TIMEOUT_MS = 8_000;

/**
 * 解析 VITE_HTTP_TIMEOUT_MS（毫秒）。空或非法时按环境回退：开发 2min、生产 8s。
 */
function resolveHttpTimeoutMs(): number {
  const raw = import.meta.env.VITE_HTTP_TIMEOUT_MS?.trim();
  const fallback = import.meta.env.DEV ? DEFAULT_DEV_TIMEOUT_MS : DEFAULT_PROD_TIMEOUT_MS;
  if (raw === undefined || raw === "") return fallback;
  const n = Number.parseInt(raw, 10);
  if (!Number.isFinite(n) || n <= 0) return fallback;
  return n;
}

/**
 * API 根路径，默认 /api/v1（与 .env.development / .env.production 一致）。
 */
function resolveApiBaseUrl(): string {
  const raw = import.meta.env.VITE_API_BASE_URL?.trim();
  if (raw === undefined || raw === "") return "/api/v1";
  return raw;
}

export const http = axios.create({
  baseURL: resolveApiBaseUrl(),
  timeout: resolveHttpTimeoutMs(),
  withCredentials: true
});

/* ---------- 请求拦截器 ---------- */

// httpOnly Cookie 方案下不需要手动注入 Authorization Header，浏览器自动携带 Cookie。

/* ---------- 响应拦截器：AT 过期自动续签 ---------- */

let isRefreshing = false;
let pendingQueue: Array<{ resolve: () => void; reject: (err: unknown) => void }> = [];

/**
 * 将等待续签的请求排队，续签成功后依次重发。
 */
function enqueueRetry(): Promise<void> {
  return new Promise((resolve, reject) => {
    pendingQueue.push({ resolve, reject });
  });
}

function flushQueue(error?: unknown): void {
  for (const item of pendingQueue) {
    if (error) {
      item.reject(error);
    } else {
      item.resolve();
    }
  }
  pendingQueue = [];
}

http.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiResponse<null>>) => {
    const originalRequest = error.config;
    const data = error.response?.data;

    const isAuthExpired =
      data?.code === "A00001" &&
      typeof data?.message === "string" &&
      data.message.includes("过期");

    const isRefreshRequest = originalRequest?.url?.includes("/admin/auth/refresh");

    if (!isAuthExpired || isRefreshRequest || !originalRequest) {
      return Promise.reject(error);
    }

    // AT 过期，尝试用 RT（Cookie 自动携带）续签
    if (isRefreshing) {
      await enqueueRetry();
      return http(originalRequest);
    }

    isRefreshing = true;
    try {
      await http.post("/admin/auth/refresh");
      flushQueue();
      return http(originalRequest);
    } catch (refreshError) {
      flushQueue(refreshError);
      clearStoredUser();
      window.location.href = "/login";
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  }
);

/* ---------- 统一响应解包 ---------- */

function parseAxiosError(error: AxiosError<ApiResponse<null>>): Error {
  if (error.response?.data?.message) {
    return new Error(error.response.data.message);
  }
  return new Error(error.message || "Request failed");
}

/**
 * 统一解包：提取 ApiResponse<T>.data，非 000000 码视为业务错误。
 */
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
