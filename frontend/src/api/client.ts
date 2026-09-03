import type {
  Credential,
  CredentialFilter,
  CredentialInput,
  CredentialList,
  ErrorResponse,
  IdentityProfile,
  IdentityProfileFilter,
  IdentityProfileInput,
  IdentityProfileList,
  Platform,
  PlatformInput,
  PlatformList,
  RuntimeStatus,
  StorageResult,
  TOTPCodeList,
  UpdateInfo,
} from "@/api/types";

const apiBase = import.meta.env.PROD ? "http://127.0.0.1:8787" : "";

export class APIError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code: string,
  ) {
    super(message);
    this.name = "APIError";
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(apiBase + path, {
    ...init,
    headers: init?.body ? { "Content-Type": "application/json", ...init.headers } : init?.headers,
  });
  if (!response.ok) {
    let body: ErrorResponse = {};
    try {
      body = (await response.json()) as ErrorResponse;
    } catch {
      // The status still provides a useful fallback when an intermediary returns non-JSON.
    }
    throw new APIError(
      body.error?.message ?? `请求失败（${response.status}）`,
      response.status,
      body.error?.code ?? "request_failed",
    );
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export function listCredentials(filter: CredentialFilter): Promise<CredentialList> {
  const params = new URLSearchParams();
  if (filter.provider) params.set("provider", filter.provider);
  if (filter.query) params.set("q", filter.query);
  const suffix = params.size > 0 ? `?${params.toString()}` : "";
  return request<CredentialList>(`/api/v1/credentials${suffix}`);
}

export function listTOTPCodes(): Promise<TOTPCodeList> {
  return request<TOTPCodeList>("/api/v1/totp");
}

export function createCredential(input: CredentialInput): Promise<Credential> {
  return request<Credential>("/api/v1/credentials", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateCredential(id: number, input: CredentialInput): Promise<Credential> {
  return request<Credential>(`/api/v1/credentials/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deleteCredential(id: number): Promise<void> {
  return request<void>(`/api/v1/credentials/${id}`, { method: "DELETE" });
}

export function listIdentityProfiles(filter: IdentityProfileFilter): Promise<IdentityProfileList> {
  const params = new URLSearchParams();
  if (filter.country) params.set("country", filter.country);
  if (filter.query) params.set("q", filter.query);
  const suffix = params.size > 0 ? `?${params.toString()}` : "";
  return request<IdentityProfileList>(`/api/v1/identities${suffix}`);
}

export function createIdentityProfile(input: IdentityProfileInput): Promise<IdentityProfile> {
  return request<IdentityProfile>("/api/v1/identities", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateIdentityProfile(id: number, input: IdentityProfileInput): Promise<IdentityProfile> {
  return request<IdentityProfile>(`/api/v1/identities/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function deleteIdentityProfile(id: number): Promise<void> {
  return request<void>(`/api/v1/identities/${id}`, { method: "DELETE" });
}

export function listPlatforms(): Promise<PlatformList> {
  return request<PlatformList>("/api/v1/platforms");
}

export function createPlatform(input: PlatformInput): Promise<Platform> {
  return request<Platform>("/api/v1/platforms", { method: "POST", body: JSON.stringify(input) });
}

export function updatePlatform(id: number, input: PlatformInput): Promise<Platform> {
  return request<Platform>(`/api/v1/platforms/${id}`, { method: "PUT", body: JSON.stringify(input) });
}

export function deletePlatform(id: number): Promise<void> {
  return request<void>(`/api/v1/platforms/${id}`, { method: "DELETE" });
}

export function getSettingsStatus(): Promise<RuntimeStatus> {
  return request<RuntimeStatus>("/api/v1/settings/status");
}

export function migrateStorage(databasePath: string): Promise<StorageResult> {
  return request<StorageResult>("/api/v1/settings/storage", {
    method: "PUT",
    body: JSON.stringify({ database_path: databasePath }),
  });
}

export function checkUpdates(): Promise<UpdateInfo> {
  return request<UpdateInfo>("/api/v1/settings/updates");
}
