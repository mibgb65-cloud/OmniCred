import { describe, expect, it } from "vitest";

import { parseCredentialEntries, parseCredentialText } from "@/lib/credential-text-parser";

describe("credential text parser", () => {
  it("parses Chinese labels and preserves password symbols", () => {
    expect(parseCredentialText(`
      邮箱：avery.parker27@example.com
      密码：T7#qL2$vN9!mX4@p:tail
      用户名：AveryParker27
    `)).toEqual({
      account: "avery.parker27@example.com",
      password: "T7#qL2$vN9!mX4@p:tail",
      username: "AveryParker27",
    });
  });

  it("accepts common English and account label aliases", () => {
    expect(parseCredentialText("account: user@example.com\npassword: secret\nusername: octocat")).toEqual({
      account: "user@example.com",
      password: "secret",
      username: "octocat",
    });
  });

  it("parses multiple four-dash rows and ignores a pasted header", () => {
    expect(parseCredentialEntries(`
      邮箱----密码----用户名
      avery.parker27@example.com----T7#qL2$vN9!mX4@p----AveryParker27
      riley.stone83@example.com----M4@zK8#pR2$qW7!n----RileyStone83
    `)).toEqual([
      { account: "avery.parker27@example.com", password: "T7#qL2$vN9!mX4@p", username: "AveryParker27" },
      { account: "riley.stone83@example.com", password: "M4@zK8#pR2$qW7!n", username: "RileyStone83" },
    ]);
  });
});
