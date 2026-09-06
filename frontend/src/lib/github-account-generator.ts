const lowercase = "abcdefghijkmnopqrstuvwxyz";
const digits = "23456789";

const firstNames = [
  "Liam", "Noah", "Ethan", "Mason", "Lucas", "Owen", "Caleb", "Julian",
  "Adrian", "Nolan", "Miles", "Leo", "Evan", "Ryan", "Aaron", "Isaac",
  "Olivia", "Emma", "Ava", "Mia", "Sophia", "Chloe", "Grace", "Lily",
  "Nora", "Zoe", "Ella", "Lucy", "Maya", "Alice", "Ruby", "Clara",
];

const lastNames = [
  "Parker", "Bennett", "Carter", "Collins", "Foster", "Hayes", "Morgan", "Reed",
  "Brooks", "Cooper", "Ellis", "Grant", "Lewis", "Miller", "Murphy", "Nelson",
  "Perry", "Rogers", "Russell", "Scott", "Stewart", "Turner", "Walker", "Ward",
  "Wright", "Young", "Baker", "Bailey", "Clark", "Davis", "Evans", "Fisher",
];

export function generateGithubUsername() {
  const firstName = firstNames[randomIndex(firstNames.length)];
  const lastName = lastNames[randomIndex(lastNames.length)];
  return firstName + lastName + randomCharacter(lowercase) + randomCharacter(digits);
}

function randomCharacter(characters: string) {
  return characters[randomIndex(characters.length)];
}

function randomIndex(maximum: number) {
  if (!globalThis.crypto?.getRandomValues) {
    throw new Error("当前环境不支持安全随机数生成");
  }
  const range = 0x1_0000_0000;
  const limit = Math.floor(range / maximum) * maximum;
  const values = new Uint32Array(1);
  do {
    globalThis.crypto.getRandomValues(values);
  } while (values[0] >= limit);
  return values[0] % maximum;
}
