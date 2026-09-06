const lowercase = "abcdefghijkmnopqrstuvwxyz";
const uppercase = "ABCDEFGHJKLMNPQRSTUVWXYZ";
const digits = "23456789";
const symbols = "!@#$%^&*_=+?-";

const passwordLength = 18;

/**
 * 生成适合大多数账号平台的强密码，并确保每类字符至少出现一次。
 * 易混淆字符（如 0/O、1/l/I）被排除，方便需要手动输入时辨认。
 */
export function generateStrongPassword() {
  const characters = lowercase + uppercase + digits + symbols;
  const password = [
    randomCharacter(lowercase),
    randomCharacter(uppercase),
    randomCharacter(digits),
    randomCharacter(symbols),
  ];

  while (password.length < passwordLength) password.push(randomCharacter(characters));

  for (let index = password.length - 1; index > 0; index--) {
    const target = randomIndex(index + 1);
    [password[index], password[target]] = [password[target], password[index]];
  }

  return password.join("");
}

function randomCharacter(characters: string) {
  return characters[randomIndex(characters.length)];
}

function randomIndex(maximum: number) {
  if (!globalThis.crypto?.getRandomValues) {
    throw new Error("当前环境不支持安全随机数生成");
  }

  // 拒绝超出整除区间的随机值，避免取模导致字符分布偏斜。
  const range = 0x1_0000_0000;
  const limit = Math.floor(range / maximum) * maximum;
  const values = new Uint32Array(1);
  do {
    globalThis.crypto.getRandomValues(values);
  } while (values[0] >= limit);

  return values[0] % maximum;
}
