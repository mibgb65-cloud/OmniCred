export interface Credential {
  id: number;
  provider: string;
  account: string;
  username: string;
  password: string;
  totp_secret: string;
  recovery_codes: string[];
  created_at: string;
  updated_at: string;
}

export interface CredentialInput {
  provider: string;
  account: string;
  username: string;
  password: string;
  totp_secret: string;
  recovery_codes: string[];
}

export interface IdentityProfile {
  id: number;
  country: string;
  full_name: string;
  localized_name: string;
  first_name: string;
  middle_name: string;
  last_name: string;
  gender: string;
  birth_date: string;
  street_address: string;
  city: string;
  region: string;
  postal_code: string;
  phone: string;
  email: string;
  password: string;
  created_at: string;
  updated_at: string;
}

export type IdentityProfileInput = Omit<IdentityProfile, "id" | "created_at" | "updated_at">;

export interface IdentityProfileFilter {
  country?: string;
  query?: string;
}

export interface IdentityProfileList {
  items: IdentityProfile[];
}

export interface CredentialFilter {
  provider?: string;
  query?: string;
}

export interface CredentialList {
  items: Credential[];
}

export interface TOTPCodeList {
  items: Array<{ credential_id: number; code: string }>;
  seconds_remaining: number;
  period: number;
  generated_at: string;
}

export interface Platform {
  id: number;
  name: string;
  credential_count: number;
  created_at: string;
  updated_at: string;
}

export interface PlatformInput {
  name: string;
}

export interface PlatformList {
  items: Platform[];
}

export interface RuntimeStatus {
  version: string;
  database_path: string;
  pending_database_path?: string;
  config_path: string;
  api_address: string;
  repository_url: string;
  started_at: string;
  uptime_seconds: number;
  database_ok: boolean;
  credential_count: number;
  uninstall_available: boolean;
}

export interface StorageResult {
  database_path: string;
  restart_required: boolean;
}

export interface UpdateInfo {
  current_version: string;
  latest_version: string;
  update_available: boolean;
  release_url: string;
  published_at?: string;
  checked_at: string;
  status: "ok" | "no_releases";
}

export interface ErrorResponse {
  error?: {
    code?: string;
    message?: string;
  };
}
