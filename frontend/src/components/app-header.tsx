import { HardDrive, ShieldCheck } from "lucide-react";

import { ThemeToggle } from "@/components/theme-toggle";
import { Badge } from "@/components/ui/badge";

export function AppHeader() {
  return (
    <header className="flex h-16 shrink-0 items-center border-b border-border bg-card/72 backdrop-blur-xl">
      <div className="flex w-full items-center justify-between gap-4 px-4 sm:px-5">
        <div className="flex min-w-0 items-center gap-3">
          <div className="grid size-9 shrink-0 place-items-center rounded-xl bg-primary text-primary-foreground shadow-sm shadow-primary/20">
            <ShieldCheck className="size-5" aria-hidden="true" />
          </div>
          <div className="min-w-0">
            <p className="truncate text-base font-bold tracking-tight">OmniCred</p>
            <p className="truncate text-[11px] text-muted-foreground">本地账号凭据工作台</p>
          </div>
        </div>
        <div className="flex items-center gap-2.5">
          <Badge variant="success" className="hidden sm:inline-flex">
            <HardDrive className="size-3" aria-hidden="true" />
            数据仅存本机
          </Badge>
          <span className="hidden h-5 w-px bg-border sm:block" aria-hidden="true" />
          <ThemeToggle />
        </div>
      </div>
    </header>
  );
}
