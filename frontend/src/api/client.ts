import { message } from "antd";
import type { ErrorEnvelope, SuccessEnvelope } from "./schema";

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly requestId?: string;

  constructor(status: number, envelope: ErrorEnvelope) {
    super(envelope.error.message);
    this.name = "ApiError";
    this.status = status;
    this.code = envelope.error.code;
    this.requestId = envelope.error.request_id;
  }
}

export async function apiRequest<T>(
  path: string,
  init: RequestInit = {}
): Promise<T> {
  const headers = new Headers(init.headers);
  const isFormData = typeof FormData !== "undefined" && init.body instanceof FormData;
  if (init.body !== undefined && !isFormData && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`/api/v1${path}`, {
    ...init,
    headers,
    credentials: "same-origin"
  });
  if (!response.ok) {
    const envelope = (await response.json()) as ErrorEnvelope;
    throw new ApiError(response.status, envelope);
  }
  const method = (init.method ?? "GET").toUpperCase();
  if (response.status === 204) {
    showSuccessMessage(path, method);
    return undefined as T;
  }
  const envelope = (await response.json()) as SuccessEnvelope<T>;
  showSuccessMessage(path, method);
  return envelope.data;
}

function showSuccessMessage(path: string, method: string) {
  if (method === "GET" || path === "/public/visits") return;
  void message.success("操作成功");
}
