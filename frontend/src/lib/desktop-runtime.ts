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
