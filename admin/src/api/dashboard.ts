import type { DashboardSummary } from "../types";
import { http, requestData } from "./client";

export const dashboardApi = {
  summary() {
    return requestData<DashboardSummary>(http.get("/admin/dashboard"));
  }
};
