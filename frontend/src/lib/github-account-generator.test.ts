import { describe, expect, it } from "vitest";

import { generateGithubPassword, generateGithubUsername } from "@/lib/github-account-generator";

describe("GitHub account generator", () => {
  it("combines English names with a short random suffix", () => {
    for (let index = 0; index < 20; index++) {
      const username = generateGithubUsername();
      expect(username).toMatch(/^[A-Z][a-z]+[A-Z][a-z]+[a-z][0-9]$/);
      expect(username).not.toMatch(/^User/i);
      expect(username.length).toBeLessThanOrEqual(39);
    }
  });

  it("generates an 18-character password with all character classes", () => {
    const password = generateGithubPassword();
    expect(password).toHaveLength(18);
    expect(password).toMatch(/[a-z]/);
    expect(password).toMatch(/[A-Z]/);
    expect(password).toMatch(/[0-9]/);
    expect(password).toMatch(/[!@#$%^&*_=+?-]/);
  });
});
