import type { AuditLog, PagedResult } from "../types";
import { http, requestData } from "./client";

export const auditLogsApi = {
  list(page: number, pageSize: number) {
    return requestData<PagedResult<AuditLog>>(http.get("/admin/audit-logs", { params: { page, pageSize } }));
  }
};
