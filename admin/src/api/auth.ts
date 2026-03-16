import type { AdminUser, AuthPayload } from "../types";
import { http, requestData } from "./client";

export const authApi = {
  login(payload: { username: string; password: string }) {
    return requestData<AuthPayload>(http.post("/admin/auth/login", payload));
  },
  refresh(refreshToken: string) {
    return requestData<{ token: string }>(http.post("/admin/auth/refresh", { refreshToken }));
  },
  me() {
    return requestData<Omit<AdminUser, "password">>(http.get("/admin/auth/me"));
  }
};
