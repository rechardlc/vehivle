/// <reference types="vite/client" />

/** 与 admin/.env* 中 VITE_ 变量对齐，供 import.meta.env 类型合并 */
interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_HTTP_TIMEOUT_MS?: string;
}
