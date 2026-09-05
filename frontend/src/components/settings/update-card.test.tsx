import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { RuntimeStatus, UpdateState } from "@/api/types";
import { UpdateCard } from "@/components/settings/update-card";

const info: RuntimeStatus = {
  version: "0.3.2", database_path: "C:\\data\\omnicred.db", config_path: "C:\\data\\config.json",
  api_address: "127.0.0.1:8787", repository_url: "https://github.com/mibgb65-cloud/OmniCred",
  started_at: "2026-09-06T00:00:00Z", uptime_seconds: 1, database_ok: true, credential_count: 0, uninstall_available: true,
};

function setup(downloadAvailable = true) {
  let state: UpdateState = { phase: "idle", downloaded: 0, total: 0 };
  const bridge = {
    UpdateStatus: vi.fn(async () => state),
    DownloadUpdate: vi.fn(async () => (state = { phase: "downloading", downloaded: 25, total: 100, version: "v0.4.0" })),
    CancelUpdate: vi.fn(async () => (state = { phase: "idle", downloaded: 0, total: 0 })),
    RestartToUpdate: vi.fn(async () => {}),
  };
  vi.stubGlobal("go", { desktop: { App: bridge } });
  vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({
    current_version: "0.3.2", latest_version: "v0.4.0", update_available: true,
    download_available: downloadAvailable, status: "ok",
    unavailable_reason: downloadAvailable ? undefined : "此版本未提供应用内更新清单，请从发布页安装",
  }), { status: 200 })));
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  const view = render(<QueryClientProvider client={client}><UpdateCard info={info} /></QueryClientProvider>);
  return {
    bridge, client, view,
    async setState(value: UpdateState) {
      state = value;
      await act(async () => {
        await client.cancelQueries({ queryKey: ["app-update-state"] });
        client.setQueryData(["app-update-state"], value);
        await new Promise((resolve) => setTimeout(resolve, 0));
      });
    },
  };
}

afterEach(() => { cleanup(); vi.unstubAllGlobals(); document.documentElement.classList.remove("dark"); });

describe("应用内更新", () => {
  it.each(["light", "dark"])("%s 主题：下载校验完成后由用户触发安装", async (theme) => {
    if (theme === "dark") document.documentElement.classList.add("dark");
    const user = userEvent.setup();
    const { bridge, setState } = setup();
    await user.click(screen.getByRole("button", { name: "检查更新" }));
    await user.click(await screen.findByRole("button", { name: "下载更新" }));
    expect(await screen.findByRole("progressbar", { name: "安装包下载进度" })).toHaveAttribute("aria-valuenow", "25");
    expect(screen.getByRole("button", { name: "检查更新" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: "重启并更新" })).not.toBeInTheDocument();
    await setState({ phase: "verifying", downloaded: 100, total: 100 });
    expect(screen.getByText("下载完成，正在校验安装包…")).toBeInTheDocument();
    expect(bridge.RestartToUpdate).not.toHaveBeenCalled();
    await setState({ phase: "ready", version: "v0.4.0", downloaded: 100, total: 100 });
    expect(screen.getByText(/已下载并通过 SHA-256 校验/)).toBeInTheDocument();
    expect(bridge.RestartToUpdate).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "重启并更新" }));
    await waitFor(() => expect(bridge.RestartToUpdate).toHaveBeenCalledOnce());
    expect(screen.getByRole("button", { name: "重启并更新" })).toBeDisabled();
  });

  it("取消下载后可以重新下载", async () => {
    const user = userEvent.setup();
    const { bridge } = setup();
    await user.click(screen.getByRole("button", { name: "检查更新" }));
    await user.click(await screen.findByRole("button", { name: "下载更新" }));
    await user.click(await screen.findByRole("button", { name: "取消下载" }));
    expect(await screen.findByRole("button", { name: "下载更新" })).toBeEnabled();
    expect(bridge.CancelUpdate).toHaveBeenCalledOnce();
  });

  it("校验失败时显示错误和重试入口，不提供安装入口", async () => {
    const user = userEvent.setup();
    const { setState } = setup();
    await user.click(screen.getByRole("button", { name: "检查更新" }));
    await screen.findByRole("button", { name: "下载更新" });
    await setState({ phase: "error", downloaded: 100, total: 100, error: "安装包 SHA-256 校验失败" });
    expect(screen.getByRole("alert")).toHaveTextContent("SHA-256 校验失败");
    expect(screen.getByRole("button", { name: "重新下载" })).toBeEnabled();
    expect(screen.queryByRole("button", { name: "重启并更新" })).not.toBeInTheDocument();
  });

  it("取消 Windows 授权后保留重试入口并展示 Go 返回的字符串错误", async () => {
    const user = userEvent.setup();
    const { bridge, setState } = setup();
    bridge.RestartToUpdate.mockRejectedValueOnce("已取消管理员授权，应用尚未退出");
    await setState({ phase: "ready", version: "v0.4.0", downloaded: 100, total: 100 });
    await user.click(screen.getByRole("button", { name: "重启并更新" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("已取消管理员授权");
    expect(screen.getByRole("button", { name: "重启并更新" })).toBeEnabled();
    await user.click(screen.getByRole("button", { name: "重启并更新" }));
    expect(bridge.RestartToUpdate).toHaveBeenCalledTimes(2);
  });

  it("旧发布缺少清单时说明原因", async () => {
    const user = userEvent.setup();
    setup(false);
    await user.click(screen.getByRole("button", { name: "检查更新" }));
    expect(await screen.findByText(/此版本未提供应用内更新清单/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "下载更新" })).not.toBeInTheDocument();
  });

  it("离开设置后再进入仍然显示已有下载状态", async () => {
    const user = userEvent.setup();
    const { bridge, view, client } = setup();
    await user.click(screen.getByRole("button", { name: "检查更新" }));
    await user.click(await screen.findByRole("button", { name: "下载更新" }));
    await screen.findByRole("progressbar");
    view.unmount();
    render(<QueryClientProvider client={client}><UpdateCard info={info} /></QueryClientProvider>);
    expect(await screen.findByRole("progressbar")).toHaveAttribute("aria-valuenow", "25");
    expect(bridge.DownloadUpdate).toHaveBeenCalledOnce();
  });
});
