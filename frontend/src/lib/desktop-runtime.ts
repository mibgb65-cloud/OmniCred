import type { UpdateState } from "@/api/types";

interface DesktopRuntime {
  BrowserOpenURL: (url: string) => void;
  Quit: () => void;
  WindowIsMaximised: () => Promise<boolean>;
  WindowMinimise: () => void;
  WindowSetDarkTheme: () => void;
  WindowSetLightTheme: () => void;
  WindowToggleMaximise: () => void;
}

interface DesktopAppBridge {
  ChooseDatabasePath: () => Promise<string>;
  Uninstall: () => Promise<void>;
  UpdateStatus: () => Promise<UpdateState>;
  DownloadUpdate: () => Promise<UpdateState>;
  CancelUpdate: () => Promise<UpdateState>;
  RestartToUpdate: () => Promise<void>;
}

function runtime(): DesktopRuntime | undefined {
  return (window as Window & { runtime?: DesktopRuntime }).runtime;
}

export function isDesktopRuntime() {
  return runtime() !== undefined;
}

export function minimiseWindow() {
  runtime()?.WindowMinimise();
}

export function toggleMaximiseWindow() {
  runtime()?.WindowToggleMaximise();
}

export async function isWindowMaximised() {
  return (await runtime()?.WindowIsMaximised()) ?? false;
}

export function closeWindow() {
  runtime()?.Quit();
}

export function setWindowTheme(theme: "light" | "dark") {
  if (theme === "dark") runtime()?.WindowSetDarkTheme();
  else runtime()?.WindowSetLightTheme();
}

export function openExternal(url: string) {
  runtime()?.BrowserOpenURL(url);
}

export async function chooseDatabasePath() {
  const bridge = (window as Window & { go?: { desktop?: { App?: DesktopAppBridge } } }).go?.desktop?.App;
  return (await bridge?.ChooseDatabasePath()) ?? "";
}

export async function uninstallApplication() {
  const bridge = (window as Window & { go?: { desktop?: { App?: DesktopAppBridge } } }).go?.desktop?.App;
  await bridge?.Uninstall();
}

function updateBridge() {
  return (window as Window & { go?: { desktop?: { App?: DesktopAppBridge } } }).go?.desktop?.App;
}

export function hasUpdateBridge() {
  return typeof updateBridge()?.UpdateStatus === "function";
}

async function invokeUpdate<T>(call: (bridge: DesktopAppBridge) => Promise<T>): Promise<T> {
  const bridge = updateBridge();
  if (!bridge) throw new Error("应用内更新仅在桌面应用中可用");
  try {
    return await call(bridge);
  } catch (error) {
    // Wails 的 Go error 可能以字符串拒绝 Promise，统一为可展示的 Error。
    throw error instanceof Error ? error : new Error(typeof error === "string" ? error : "更新操作失败，请重试");
  }
}

export const getUpdateState = () => invokeUpdate((bridge) => bridge.UpdateStatus());
export const downloadUpdate = () => invokeUpdate((bridge) => bridge.DownloadUpdate());
export const cancelUpdate = () => invokeUpdate((bridge) => bridge.CancelUpdate());
export const restartToUpdate = () => invokeUpdate((bridge) => bridge.RestartToUpdate());
