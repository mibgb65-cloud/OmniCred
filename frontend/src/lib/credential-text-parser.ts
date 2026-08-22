export interface ParsedCredentialText {
  account?: string;
  password?: string;
  username?: string;
  totp_secret?: string;
}

const labels: Record<string, keyof ParsedCredentialText> = {
  邮箱: "account",
  登录邮箱: "account",
  账号: "account",
  账户: "account",
  登录账号: "account",
  email: "account",
  mail: "account",
  account: "account",
  login: "account",
  密码: "password",
  password: "password",
  pass: "password",
  pwd: "password",
  用户名: "username",
  昵称: "username",
  username: "username",
  user: "username",
  "2fa": "totp_secret",
  "2fa密钥": "totp_secret",
  totp: "totp_secret",
  totp密钥: "totp_secret",
  验证码密钥: "totp_secret",
};

export function parseCredentialText(text: string): ParsedCredentialText {
  const result: ParsedCredentialText = {};
  for (const line of text.split(/\r?\n/)) {
    const separator = firstSeparator(line);
    if (separator < 0) continue;
    const label = normalizeLabel(line.slice(0, separator));
    const field = labels[label];
    const value = line.slice(separator + 1).trim();
    if (field && value) result[field] = value;
  }
  return result;
}

export function parseCredentialEntries(text: string): ParsedCredentialText[] {
  const rows = text.split(/\r?\n/).map(parseDelimitedRow).filter((row): row is ParsedCredentialText => row !== null);
  if (rows.length > 0) return rows;
  const labeled = parseCredentialText(text);
  return Object.keys(labeled).length > 0 ? [labeled] : [];
}

function parseDelimitedRow(line: string): ParsedCredentialText | null {
  const first = line.indexOf("----");
  const last = line.lastIndexOf("----");
  if (first < 0 || last <= first) return null;

  const account = line.slice(0, first).trim();
  const password = line.slice(first + 4, last).trim();
  const username = line.slice(last + 4).trim();
  if (labels[normalizeLabel(account)] === "account" && labels[normalizeLabel(password)] === "password") return null;
  if (!account || !password) return null;
  return { account, password, ...(username ? { username } : {}) };
}

function firstSeparator(line: string) {
  const positions = [line.indexOf(":"), line.indexOf("：")].filter((position) => position >= 0);
  return positions.length > 0 ? Math.min(...positions) : -1;
}

function normalizeLabel(value: string) {
  return value.trim().toLowerCase().replace(/[\s_-]+/g, "");
}
