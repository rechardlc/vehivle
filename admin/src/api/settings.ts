import type { PriceMode, SystemSettings } from "../types";
import { http, requestData } from "./client";
import { isEmptyObject } from "../utils";

function backendPriceToUi(mode: string): PriceMode {
  if (mode === "show_price") return "msrp";
  if (mode === "phone_inquiry") return "negotiable";
  return "negotiable";
}

function uiPriceToBackend(mode: PriceMode): string {
  return mode === "msrp" ? "show_price" : "phone_inquiry";
}

function mapFromBackend(raw: SystemSettings): SystemSettings {
  return { ...raw, defaultPriceMode: backendPriceToUi(raw.defaultPriceMode as string) };
}

function mapToBackend(payload: Partial<SystemSettings>): Record<string, unknown> {
  const out: Record<string, unknown> = { ...payload };
  if (payload.defaultPriceMode) {
    out.defaultPriceMode = uiPriceToBackend(payload.defaultPriceMode);
  }
  return out;
}

export const settingsApi = {
  async get() {
    const data = await requestData<SystemSettings | null>(http.get("/admin/system-settings"));
    const isEmpty = isEmptyObject(data as unknown as Record<string, unknown>);
    return isEmpty ? null : mapFromBackend(data as SystemSettings);
  },
  async create(payload: Partial<SystemSettings>) {
    const data = await requestData<SystemSettings>(http.post("/admin/system-settings", mapToBackend(payload)));
    return mapFromBackend(data);
  },
  async update(payload: Partial<SystemSettings>) {
    const data = await requestData<SystemSettings>(http.put("/admin/system-settings", mapToBackend(payload)));
    return mapFromBackend(data);
  }
};
