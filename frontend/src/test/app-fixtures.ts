import type { Credential, IdentityProfile } from "@/api/types";

export const saved: Credential = {
  id: 1,
  provider: "github",
  account: "user@example.com",
  username: "octocat",
  password: "test-password-do-not-use",
  totp_secret: "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
  recovery_codes: ["alpha-bravo", "charlie-delta"],
  created_at: "2026-08-21T02:00:00Z",
  updated_at: "2026-08-21T02:00:00Z",
};

export const savedIdentity: IdentityProfile = {
  id: 1,
  country: "Philippines",
  full_name: "Angelo Santos",
  localized_name: "安杰洛·桑托斯",
  first_name: "Angelo",
  middle_name: "Reyes",
  last_name: "Santos",
  gender: "male",
  birth_date: "1998-08-16",
  street_address: "Unit 402, Sunshine Court, Aurora Blvd, Cubao",
  city: "Quezon City",
  region: "Metro Manila",
  postal_code: "1109",
  phone: "+63 (917) 482-9301",
  email: "angelo@example.com",
  password: "identity-password-do-not-use",
  created_at: "2026-09-03T02:00:00Z",
  updated_at: "2026-09-03T02:00:00Z",
};

export const savedPlatform = {
  id: 1,
  name: "github",
  credential_count: 1,
  created_at: "2026-08-21T02:00:00Z",
  updated_at: "2026-08-21T02:00:00Z",
};

export const unusedPlatform = { ...savedPlatform, id: 2, name: "notion", credential_count: 0 };
