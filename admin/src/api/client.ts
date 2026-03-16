import axios, { AxiosError, type AxiosResponse } from "axios";
import { mockAdapter } from "../mock/adapter";
import type { ApiResponse } from "../types";
import { getAuthState } from "../state/auth";

export const http = axios.create({
  baseURL: "/api/v1",
  timeout: 8000,
  adapter: mockAdapter
});

http.interceptors.request.use((config) => {
  const auth = getAuthState();
  if (auth?.token) {
    config.headers = config.headers ?? {};
    config.headers.Authorization = `Bearer ${auth.token}`;
  }
  return config;
});

function parseAxiosError(error: AxiosError<ApiResponse<null>>): Error {
  if (error.response?.data?.message) {
    return new Error(error.response.data.message);
  }
  return new Error(error.message || "Request failed");
}

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
