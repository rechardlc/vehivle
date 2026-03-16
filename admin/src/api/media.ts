import { http, requestData } from "./client";

export interface UploadPolicyPayload {
  filename: string;
  mimeType: string;
  size: number;
}

export interface UploadPolicyResult {
  uploadUrl: string;
  objectKey: string;
  expiresIn: number;
}

export interface UploadCompletePayload {
  objectKey: string;
  mimeType: string;
  size: number;
}

export const mediaApi = {
  uploadPolicy(payload: UploadPolicyPayload) {
    return requestData<UploadPolicyResult>(http.post("/admin/media/upload-policy", payload));
  },
  complete(payload: UploadCompletePayload) {
    return requestData<{ id: string; storageKey: string; mimeType: string; fileSize: number }>(
      http.post("/admin/media/complete", payload)
    );
  }
};
