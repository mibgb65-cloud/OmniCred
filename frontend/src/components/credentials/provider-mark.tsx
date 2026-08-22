import { Bot, GitBranch, Globe2, KeyRound, ShieldCheck, type LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

const providers: Record<string, { label: string; icon: LucideIcon; className: string }> = {
  github: { label: "GitHub", icon: GitBranch, className: "bg-slate-900 text-white dark:bg-white dark:text-slate-950" },
  google: { label: "Google", icon: Globe2, className: "bg-blue-500/12 text-blue-700 dark:text-blue-300" },
  chatgpt: { label: "ChatGPT", icon: Bot, className: "bg-emerald-500/12 text-emerald-700 dark:text-emerald-300" },
  omnicred: { label: "OmniCred", icon: ShieldCheck, className: "bg-primary/12 text-primary" },
};

export function providerLabel(provider: string) {
  return providers[provider.toLowerCase()]?.label ?? provider;
}

export function ProviderMark({ provider, className }: { provider: string; className?: string }) {
  const meta = providers[provider.toLowerCase()] ?? {
    label: provider,
    icon: KeyRound,
    className: "bg-secondary text-secondary-foreground",
  };
  const Icon = meta.icon;
  return (
    <div
      className={cn("grid size-11 shrink-0 place-items-center rounded-xl", meta.className, className)}
      role="img"
      aria-label={`${meta.label} 平台`}
    >
      <Icon className="size-5" aria-hidden="true" />
    </div>
  );
}
