import type { AuthMeResult, LoginResult } from "../types";
import { http, requestData } from "./client";

export const authApi = {
  /**
   * 登录：POST /admin/auth/login
   * 服务端通过 Set-Cookie 写入 access_token + refresh_token（httpOnly）。
   * 响应 body 仅返回 expiresIn，Token 不在 body 中。
   */
  login(payload: { username: string; password: string }) {
    return requestData<LoginResult>(http.post("/admin/auth/login", payload));
  },

  /**
   * 续签：POST /admin/auth/refresh
   * 无需传 body，浏览器自动携带 refresh_token Cookie。
   * 服务端写入新的 access_token Cookie。
   */
  refresh() {
    return requestData<LoginResult>(http.post("/admin/auth/refresh"));
  },

  /**
   * 获取当前用户信息：GET /admin/auth/me
   * 通过 access_token Cookie 认证，返回用户 id、username、role。
   */
  me() {
    return requestData<AuthMeResult>(http.get("/admin/auth/me"));
  },

  /**
   * 登出：POST /admin/auth/logout
   * 服务端清除 access_token + refresh_token Cookie（Set-Cookie: Max-Age=0）。
   */
  logout() {
    return requestData<null>(http.post("/admin/auth/logout"));
  }
};
