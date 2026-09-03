import {
  Bot,
  Database,
  GitBranch,
  Globe2,
  IdCard,
  KeyRound,
  LayoutGrid,
  Settings2,
  SlidersHorizontal,
  ShieldCheck,
  TriangleAlert,
  type LucideIcon,
} from "lucide-react";

import type { Platform } from "@/api/types";
import { providerLabel } from "@/components/credentials/provider-mark";
import { cn } from "@/lib/utils";

const platformIcons: Record<string, LucideIcon> = {
  github: GitBranch,
  google: Globe2,
  chatgpt: Bot,
  omnicred: ShieldCheck,
};

interface CredentialSidebarProps {
  activePage: "credentials" | "identities" | "settings";
  provider: string;
  platforms: Platform[];
  resultCount: number;
  onProviderChange: (value: string) => void;
  onManagePlatforms: () => void;
  onOpenSettings: () => void;
  onOpenIdentities: () => void;
}

export function CredentialSidebar(props: CredentialSidebarProps) {
  const filters = [
    { value: "", label: "全部账号", icon: LayoutGrid },
    ...props.platforms.map((platform) => ({
      value: platform.name,
      label: providerLabel(platform.name),
      icon: platformIcons[platform.name.toLowerCase()] ?? KeyRound,
    })),
  ];

  return (
    <aside className="shrink-0 border-b border-border bg-card/45 min-[720px]:w-52 min-[720px]:border-b-0 min-[720px]:border-r">
      <div className="sidebar-scrollbar flex gap-2 overflow-x-auto p-3 min-[720px]:h-full min-[720px]:flex-col min-[720px]:gap-5 min-[720px]:overflow-x-hidden min-[720px]:overflow-y-auto min-[720px]:p-4">
        <div className="flex shrink-0 items-center justify-between gap-2 min-[720px]:px-2">
          <p className="hidden text-[11px] font-bold uppercase tracking-[0.16em] text-muted-foreground min-[720px]:block">账号平台</p>
          <button
            type="button"
            onClick={props.onManagePlatforms}
            className="flex min-h-10 cursor-pointer items-center gap-2 rounded-xl border border-transparent px-3 text-sm font-semibold text-muted-foreground transition-colors hover:bg-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring min-[720px]:min-h-7 min-[720px]:px-2 min-[720px]:text-[11px]"
            aria-label="管理账号平台"
          >
            <SlidersHorizontal className="size-3.5" aria-hidden="true" />
            管理
          </button>
        </div>
        <nav className="flex gap-2 min-[720px]:flex-col" aria-label="平台筛选">
          {filters.map((filter) => {
            const active = props.activePage === "credentials" && props.provider === filter.value;
            const Icon = filter.icon;
            return (
              <button
                key={filter.value || "all"}
                type="button"
                aria-pressed={active}
                onClick={() => props.onProviderChange(filter.value)}
                className={cn(
                  "flex min-h-10 shrink-0 cursor-pointer items-center gap-2.5 rounded-xl border px-3 text-sm font-semibold transition-[background-color,color,border-color] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                  "min-[720px]:w-full",
                  active
                    ? "border-primary/20 bg-primary/12 text-primary"
                    : "border-transparent text-muted-foreground hover:bg-accent hover:text-foreground",
                )}
              >
                <Icon className="size-4 shrink-0" aria-hidden="true" />
                <span>{filter.label}</span>
                {active && <span className="ml-auto hidden size-1.5 rounded-full bg-primary min-[720px]:block" aria-hidden="true" />}
              </button>
            );
          })}
        </nav>

        <button
          type="button"
          onClick={props.onOpenIdentities}
          aria-current={props.activePage === "identities" ? "page" : undefined}
          aria-label="打开身份资料"
          className={cn(
            "flex min-h-10 shrink-0 cursor-pointer items-center gap-2.5 rounded-xl border px-3 text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring min-[720px]:w-full",
            props.activePage === "identities"
              ? "border-primary/20 bg-primary/12 text-primary"
              : "border-transparent text-muted-foreground hover:bg-accent hover:text-foreground min-[720px]:border-t-border",
          )}
        >
          <IdCard className="size-4" aria-hidden="true" />
          身份资料
          {props.activePage === "identities" && <span className="ml-auto hidden size-1.5 rounded-full bg-primary min-[720px]:block" aria-hidden="true" />}
        </button>

        <button
          type="button"
          onClick={props.onOpenSettings}
          aria-current={props.activePage === "settings" ? "page" : undefined}
          aria-label="打开设置"
          className={cn(
            "flex min-h-10 shrink-0 cursor-pointer items-center gap-2.5 rounded-xl border px-3 text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring min-[720px]:w-full",
            props.activePage === "settings"
              ? "border-primary/20 bg-primary/12 text-primary"
              : "border-transparent text-muted-foreground hover:bg-accent hover:text-foreground min-[720px]:border-t-border",
          )}
        >
          <Settings2 className="size-4" aria-hidden="true" />
          设置
          {props.activePage === "settings" && <span className="ml-auto hidden size-1.5 rounded-full bg-primary min-[720px]:block" aria-hidden="true" />}
        </button>

        <div className="mt-auto hidden space-y-3 min-[720px]:block">
          <section className="rounded-xl border border-border bg-background/65 p-3" aria-label="当前账号概览">
            <div className="flex items-center gap-2 text-xs font-semibold text-foreground">
              <Database className="size-3.5 text-primary" aria-hidden="true" />
              本地保险库
            </div>
            <dl className="mt-3 grid grid-cols-2 gap-2">
              <div>
                <dt className="text-[11px] text-muted-foreground">当前结果</dt>
                <dd className="mt-0.5 text-lg font-bold tabular-nums">{props.resultCount}</dd>
              </div>
              <div>
                <dt className="text-[11px] text-muted-foreground">平台数量</dt>
                <dd className="mt-0.5 text-lg font-bold tabular-nums">{props.platforms.length}</dd>
              </div>
            </dl>
          </section>

          <div className="flex gap-2.5 rounded-xl border border-amber-500/20 bg-amber-500/8 p-3 text-amber-800 dark:text-amber-200">
            <TriangleAlert className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
            <p className="text-[11px] leading-4">当前版本以明文保存在本机，请妥善保护数据库文件。</p>
          </div>

          <p className="flex items-center gap-2 px-1 text-[11px] text-muted-foreground">
            <KeyRound className="size-3.5" aria-hidden="true" />
            SQLite · 仅本机访问
          </p>
        </div>
      </div>
    </aside>
  );
}
