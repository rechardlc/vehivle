import type { SystemSettings } from "../types";
import { http, requestData } from "./client";

export const settingsApi = {
  get() {
    return requestData<SystemSettings>(http.get("/admin/system-settings"));
  },
  update(payload: Partial<SystemSettings>) {
    return requestData<SystemSettings>(http.put("/admin/system-settings", payload));
  }
};
