import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import App from "@/App";
import type { Credential } from "@/api/types";
import { TooltipProvider } from "@/components/ui/tooltip";

const saved: Credential = {
  id: 1,
  provider: "github",
  account: "user@example.com",
  username: "octocat",
  password: "test-password-do-not-use",
  totp_secret: "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
  created_at: "2026-08-21T02:00:00Z",
  updated_at: "2026-08-21T02:00:00Z",
};

const savedPlatform = {
  id: 1,
  name: "github",
  credential_count: 1,
  created_at: "2026-08-21T02:00:00Z",
  updated_at: "2026-08-21T02:00:00Z",
};

const unusedPlatform = { ...savedPlatform, id: 2, name: "notion", credential_count: 0 };

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function renderApp() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <TooltipProvider><App /></TooltipProvider>
    </QueryClientProvider>,
  );
}

describe("App", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      if (String(input).endsWith("/api/v1/totp")) return jsonResponse({
        items: [{ credential_id: 1, code: "287082" }],
        seconds_remaining: 17,
        period: 30,
        generated_at: "2026-08-22T02:00:00Z",
      });
      if (String(input).endsWith("/api/v1/settings/status")) return jsonResponse({
        version: "0.1.0",
        database_path: "C:\\Users\\test\\AppData\\Roaming\\OmniCred\\omnicred.db",
        config_path: "C:\\Users\\test\\AppData\\Roaming\\OmniCred\\config.json",
        api_address: "127.0.0.1:8787",
        repository_url: "https://github.com/mibgb65-cloud/OmniCred",
        started_at: "2026-08-22T02:00:00Z",
        uptime_seconds: 125,
        database_ok: true,
        credential_count: 1,
        uninstall_available: true,
      });
      if (String(input).endsWith("/api/v1/settings/updates")) return jsonResponse({
        current_version: "0.1.0",
        latest_version: "v0.2.0",
        update_available: true,
        release_url: "https://github.com/mibgb65-cloud/OmniCred/releases/tag/v0.2.0",
        checked_at: "2026-08-22T02:05:00Z",
        status: "ok",
      });
      if (String(input).includes("/api/v1/platforms")) {
        if (init?.method === "POST") {
          const body = JSON.parse(String(init.body)) as { name: string };
          return jsonResponse({ ...savedPlatform, id: 3, name: body.name, credential_count: 0 }, 201);
        }
        if (init?.method === "PUT") {
          const body = JSON.parse(String(init.body)) as { name: string };
          return jsonResponse({ ...savedPlatform, name: body.name });
        }
        if (init?.method === "DELETE") return new Response(null, { status: 204 });
        return jsonResponse({ items: [savedPlatform, unusedPlatform] });
      }
      if (init?.method === "POST") return jsonResponse({ ...saved, id: 2 }, 201);
      return jsonResponse({ items: [saved] });
    }));
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    localStorage.clear();
  });

  it("loads credentials and reveals a password on request", async () => {
    const user = userEvent.setup();
    renderApp();

    expect(await screen.findByText("user@example.com")).toBeInTheDocument();
    expect(screen.queryByText(saved.password)).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "显示密码" }));
    expect(screen.getByText(saved.password)).toBeInTheDocument();
    expect(await screen.findByText("287 082")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "复制 2FA 验证码" }));
    expect(await navigator.clipboard.readText()).toBe("287082");
  });

  it("validates and submits the create form", async () => {
    const user = userEvent.setup();
    renderApp();
    await screen.findByText("user@example.com");

    await user.click(screen.getByRole("button", { name: "新增账号" }));
    const dialog = screen.getByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "保存账号" }));
    expect(await screen.findByText("请选择账号平台")).toBeInTheDocument();

    expect(within(dialog).queryByRole("textbox", { name: /平台/ })).not.toBeInTheDocument();
    await user.click(within(dialog).getByRole("combobox", { name: /平台/ }));
    await user.click(screen.getByRole("option", { name: "notion" }));
    await user.click(within(dialog).getByRole("button", { name: "从文本快速解析" }));
    await user.type(within(dialog).getByLabelText("账号文本"), "邮箱：new@example.com\n密码：new-password\n用户名：person\n2FA密钥：GEZD GNBV GY3T QOJQ");
    await user.click(within(dialog).getByRole("button", { name: "解析并填入" }));
    expect(within(dialog).getByLabelText(/登录账号/)).toHaveValue("new@example.com");
    expect(within(dialog).getByLabelText(/^用户名/)).toHaveValue("person");
    expect(within(dialog).getByLabelText(/^密码/)).toHaveValue("new-password");
    expect(within(dialog).getByLabelText(/2FA 密钥/, { selector: "input" })).toHaveValue("GEZD GNBV GY3T QOJQ");
    await user.click(within(dialog).getByRole("button", { name: "保存账号" }));

    await waitFor(() => {
      const request = vi.mocked(fetch).mock.calls.find(([input, init]) => String(input).endsWith("/api/v1/credentials") && init?.method === "POST");
      expect(JSON.parse(String(request?.[1]?.body))).toMatchObject({ provider: "notion", account: "new@example.com", username: "person", password: "new-password", totp_secret: "GEZD GNBV GY3T QOJQ" });
    });
  });

  it("prefills the selected platform when creating an account and keeps it editable", async () => {
    const user = userEvent.setup();
    renderApp();
    await screen.findByText("user@example.com");

    await user.click(screen.getByRole("button", { name: "GitHub" }));
    await user.click(screen.getByRole("button", { name: "新增账号" }));

    const dialog = screen.getByRole("dialog", { name: "新增账号" });
    const platform = within(dialog).getByRole("combobox", { name: /平台/ });
    expect(platform).toHaveTextContent("GitHub");

    await user.click(platform);
    await user.click(screen.getByRole("option", { name: "notion" }));
    expect(platform).toHaveTextContent("notion");
  });

  it("supports desktop keyboard shortcuts", async () => {
    const user = userEvent.setup();
    renderApp();
    await screen.findByText("user@example.com");

    await user.keyboard("{Control>}k{/Control}");
    expect(screen.getByRole("searchbox", { name: "搜索账号或用户名" })).toHaveFocus();

    await user.keyboard("{Control>}n{/Control}");
    expect(screen.getByRole("dialog", { name: "新增账号" })).toBeInTheDocument();
  });

  it("renders desktop window controls", async () => {
    renderApp();
    await screen.findByText("user@example.com");

    expect(screen.getByRole("button", { name: "最小化窗口" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "最大化窗口" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "关闭窗口" })).toBeInTheDocument();
  });

  it("shows runtime settings and checks GitHub releases", async () => {
    const user = userEvent.setup();
    renderApp();
    await screen.findByText("user@example.com");

    await user.click(screen.getByRole("button", { name: "打开设置" }));
    expect(await screen.findByRole("heading", { name: "设置" })).toBeInTheDocument();
    expect(await screen.findByText("v0.1.0")).toBeInTheDocument();
    expect(screen.getByText("127.0.0.1:8787")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "检查更新" }));
    expect(await screen.findByText("发现 v0.2.0")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "卸载应用" }));
    expect(screen.getByRole("alertdialog", { name: "确认卸载 OmniCred？" })).toBeInTheDocument();
  });

  it("keeps the current list visible while a platform filter refreshes", async () => {
    const user = userEvent.setup();
    let finishFilter!: (response: Response) => void;
    const filteredResponse = new Promise<Response>((resolve) => {
      finishFilter = resolve;
    });
    vi.stubGlobal("fetch", vi.fn(async (input: RequestInfo | URL) => {
      if (String(input).includes("/api/v1/platforms")) return jsonResponse({ items: [savedPlatform, unusedPlatform] });
      if (String(input).includes("provider=github")) return filteredResponse;
      return jsonResponse({ items: [saved] });
    }));
    renderApp();
    await screen.findByText("user@example.com");

    await user.click(screen.getByRole("button", { name: "GitHub" }));
    expect(screen.getByText("user@example.com")).toBeInTheDocument();
    expect(screen.queryByLabelText("正在加载账号")).not.toBeInTheDocument();

    finishFilter(jsonResponse({ items: [saved] }));
    await waitFor(() => {
      expect(vi.mocked(fetch).mock.calls.some(([input]) => String(input).includes("provider=github"))).toBe(true);
    });
  });

  it("opens platform management and adds a platform", async () => {
    const user = userEvent.setup();
    renderApp();
    await screen.findByText("user@example.com");

    await user.click(screen.getByRole("button", { name: "管理账号平台" }));
    const dialog = screen.getByRole("dialog", { name: "管理账号平台" });
    expect(within(dialog).getByRole("button", { name: "删除 GitHub" })).toBeDisabled();
    await user.type(within(dialog).getByLabelText("新增平台"), "notion");
    await user.click(within(dialog).getByRole("button", { name: "添加" }));

    await waitFor(() => {
      const request = vi.mocked(fetch).mock.calls.find(([input, init]) => String(input).endsWith("/api/v1/platforms") && init?.method === "POST");
      expect(request).toBeDefined();
      expect(request?.[1]?.body).toBe(JSON.stringify({ name: "notion" }));
    });
  });

  it("renames a platform and deletes an unused platform", async () => {
    const user = userEvent.setup();
    renderApp();
    await screen.findByText("user@example.com");

    await user.click(screen.getByRole("button", { name: "管理账号平台" }));
    const dialog = screen.getByRole("dialog", { name: "管理账号平台" });
    await user.click(within(dialog).getByRole("button", { name: "修改 GitHub" }));
    const editInput = within(dialog).getByLabelText("修改 GitHub");
    await user.clear(editInput);
    await user.type(editInput, "gitlab");
    await user.click(within(dialog).getByRole("button", { name: "保存 GitHub" }));
    await waitFor(() => {
      expect(vi.mocked(fetch).mock.calls.some(([input, init]) => String(input).endsWith("/api/v1/platforms/1") && init?.method === "PUT")).toBe(true);
    });

    await user.click(within(dialog).getByRole("button", { name: "删除 notion" }));
    const confirmation = screen.getByRole("alertdialog", { name: "删除“notion”平台？" });
    await user.click(within(confirmation).getByRole("button", { name: "确认删除" }));
    await waitFor(() => {
      expect(vi.mocked(fetch).mock.calls.some(([input, init]) => String(input).endsWith("/api/v1/platforms/2") && init?.method === "DELETE")).toBe(true);
    });
  });

  it("keeps a generated GitHub account as a draft until confirmation", async () => {
    const user = userEvent.setup();
    renderApp();
    await screen.findByText("user@example.com");

    await user.click(screen.getByRole("button", { name: "生成 GitHub 账号" }));
    const dialog = screen.getByRole("dialog", { name: "生成 GitHub 账号" });
    await user.type(within(dialog).getByLabelText("邮箱"), "avery.parker27@example.com");
    await user.click(within(dialog).getByRole("button", { name: "生成待确认信息" }));

    const generatedUsername = String((within(dialog).getByLabelText("用户名") as HTMLInputElement).value);
    expect(generatedUsername).toMatch(/^[A-Z][a-z]+[A-Z][a-z]+[a-z][0-9]$/);
    expect(generatedUsername).not.toMatch(/^User/i);
    const password = within(dialog).getByLabelText("密码").getAttribute("value") ?? "";
    expect(password).toHaveLength(18);
    expect(vi.mocked(fetch).mock.calls.some(([input, init]) => String(input).includes("/api/v1/credentials") && init?.method === "POST")).toBe(false);

    await user.click(within(dialog).getByRole("button", { name: "确认并加入账号池" }));
    await waitFor(() => {
      const request = vi.mocked(fetch).mock.calls.find(([input, init]) => String(input).endsWith("/api/v1/credentials") && init?.method === "POST");
      expect(request).toBeDefined();
      expect(JSON.parse(String(request?.[1]?.body))).toMatchObject({
        provider: "github",
        account: "avery.parker27@example.com",
        username: generatedUsername,
      });
    });
  });

  it("previews and imports multiple four-dash credential rows", async () => {
    const user = userEvent.setup();
    renderApp();
    await screen.findByText("user@example.com");

    await user.click(screen.getByRole("button", { name: "新增账号" }));
    const dialog = screen.getByRole("dialog", { name: "新增账号" });
    await user.click(within(dialog).getByRole("combobox", { name: /平台/ }));
    await user.click(screen.getByRole("option", { name: "GitHub" }));
    await user.click(within(dialog).getByRole("button", { name: "从文本快速解析" }));
    await user.type(within(dialog).getByLabelText("账号文本"), [
      "邮箱----密码----用户名",
      "avery.parker27@example.com----T7#qL2$vN9!mX4@p----AveryParker27",
      "riley.stone83@example.com----M4@zK8#pR2$qW7!n----RileyStone83",
    ].join("\n"));
    await user.click(within(dialog).getByRole("button", { name: "解析并填入" }));

    expect(within(dialog).getByText("已解析 2 条账号")).toBeInTheDocument();
    expect(within(dialog).getByText("avery.parker27@example.com")).toBeInTheDocument();
    expect(within(dialog).getByText("riley.stone83@example.com")).toBeInTheDocument();
    expect(vi.mocked(fetch).mock.calls.some(([input, init]) => String(input).includes("/api/v1/credentials") && init?.method === "POST")).toBe(false);

    await user.click(within(dialog).getByRole("button", { name: "导入 2 个账号" }));
    await waitFor(() => {
      const requests = vi.mocked(fetch).mock.calls.filter(([input, init]) => String(input).endsWith("/api/v1/credentials") && init?.method === "POST");
      expect(requests).toHaveLength(2);
      expect(requests.map((request) => JSON.parse(String(request[1]?.body)).account)).toEqual([
        "avery.parker27@example.com",
        "riley.stone83@example.com",
      ]);
    });
  });
});
