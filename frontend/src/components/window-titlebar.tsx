import { Copy, Minus, ShieldCheck, Square, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import {
  closeWindow,
  isDesktopRuntime,
  isWindowMaximised,
  minimiseWindow,
  toggleMaximiseWindow,
} from "@/lib/desktop-runtime";

export function WindowTitlebar() {
  const [maximised, setMaximised] = useState(false);
  const resizeTimer = useRef<number | null>(null);

  useEffect(() => {
    if (!isDesktopRuntime()) return;

    const syncState = () => {
      void isWindowMaximised().then(setMaximised);
    };
    const handleResize = () => {
      if (resizeTimer.current !== null) window.clearTimeout(resizeTimer.current);
      resizeTimer.current = window.setTimeout(syncState, 80);
    };

    syncState();
    window.addEventListener("resize", handleResize);
    return () => {
      window.removeEventListener("resize", handleResize);
      if (resizeTimer.current !== null) window.clearTimeout(resizeTimer.current);
    };
  }, []);

  function toggleMaximise() {
    toggleMaximiseWindow();
    window.setTimeout(() => void isWindowMaximised().then(setMaximised), 80);
  }

  return (
    <div className="flex h-9 shrink-0 border-b border-border bg-card text-card-foreground" aria-label="窗口标题栏">
      <div className="window-drag flex min-w-0 flex-1 items-center gap-2 px-3" onDoubleClick={toggleMaximise}>
        <span className="grid size-5 shrink-0 place-items-center rounded-md bg-primary text-primary-foreground">
          <ShieldCheck className="size-3.5" aria-hidden="true" />
        </span>
        <span className="truncate text-xs font-semibold text-muted-foreground">OmniCred</span>
      </div>

      <div className="window-no-drag flex shrink-0" role="group" aria-label="窗口操作">
        <WindowControl label="最小化窗口" onClick={minimiseWindow}>
          <Minus className="size-4" aria-hidden="true" />
        </WindowControl>
        <WindowControl label={maximised ? "还原窗口" : "最大化窗口"} onClick={toggleMaximise}>
          {maximised ? <Copy className="size-3.5" aria-hidden="true" /> : <Square className="size-3.5" aria-hidden="true" />}
        </WindowControl>
        <WindowControl label="关闭窗口" onClick={closeWindow} close>
          <X className="size-4" aria-hidden="true" />
        </WindowControl>
      </div>
    </div>
  );
}

function WindowControl(props: { label: string; onClick: () => void; close?: boolean; children: React.ReactNode }) {
  return (
    <button
      type="button"
      aria-label={props.label}
      title={props.label}
      onClick={props.onClick}
      className={
        props.close
          ? "grid h-9 w-12 cursor-pointer place-items-center text-muted-foreground transition-colors hover:bg-destructive hover:text-white focus-visible:z-10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
          : "grid h-9 w-12 cursor-pointer place-items-center text-muted-foreground transition-colors hover:bg-accent hover:text-foreground focus-visible:z-10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
      }
    >
      {props.children}
    </button>
  );
}
