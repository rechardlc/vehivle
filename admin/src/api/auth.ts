import type { AuthMeResult, LoginResult } from "../types";
import { http, requestData } from "./client";

export const authApi = {
  /**
   * 登录：POST /admin/auth/login
   * 服务端通过 Set-Cookie 写入 refresh_token（httpOnly）。
   * 响应 body 返回 accessToken，后续请求通过 Authorization: Bearer Token 携带。
   */
  login(payload: { username: string; password: string }) {
    return requestData<LoginResult>(http.post("/admin/auth/login", payload));
  },

  /**
   * 续签：POST /admin/auth/refresh
   * 无需传 body，浏览器自动携带 refresh_token Cookie。
   * 响应 body 返回新的 accessToken，前端更新本地缓存后继续请求。
   */
  refresh() {
    return requestData<LoginResult>(http.post("/admin/auth/refresh"));
  },

  /**
   * 获取当前用户信息：GET /admin/auth/me
   * 通过 Authorization: Bearer Token 认证，返回用户 id、username、role。
   */
  me() {
    return requestData<AuthMeResult>(http.get("/admin/auth/me"));
  },

  /**
   * 登出：POST /admin/auth/logout
   * 服务端清除 refresh_token Cookie（Set-Cookie: Max-Age=0），前端同步清理 accessToken。
   */
  logout() {
    return requestData<null>(http.post("/admin/auth/logout"));
  }
};
