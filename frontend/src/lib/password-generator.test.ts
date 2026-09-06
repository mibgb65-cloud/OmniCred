import { describe, expect, it } from "vitest";

import { generateStrongPassword } from "@/lib/password-generator";

describe("password generator", () => {
  it("generates an 18-character password with all character classes", () => {
    for (let index = 0; index < 20; index++) {
      const password = generateStrongPassword();
      expect(password).toHaveLength(18);
      expect(password).toMatch(/[a-z]/);
      expect(password).toMatch(/[A-Z]/);
      expect(password).toMatch(/[0-9]/);
      expect(password).toMatch(/[!@#$%^&*_=+?-]/);
    }
  });
});
