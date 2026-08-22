import { Plus, Search, WandSparkles } from "lucide-react";
import type { RefObject } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface CredentialToolbarProps {
  search: string;
  searchInputRef: RefObject<HTMLInputElement | null>;
  onSearchChange: (value: string) => void;
  onGenerateGithub: () => void;
  onCreate: () => void;
}

export function CredentialToolbar(props: CredentialToolbarProps) {
  return (
    <section aria-label="账号搜索和操作" className="flex min-w-0 flex-1 flex-col gap-2.5 sm:flex-row sm:justify-end">
      <div className="relative min-w-0 flex-1 min-[1100px]:max-w-xl">
        <Search className="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
        <Input
          ref={props.searchInputRef}
          type="search"
          value={props.search}
          onChange={(event) => props.onSearchChange(event.target.value)}
          placeholder="搜索登录账号或用户名"
          aria-label="搜索账号或用户名"
          aria-keyshortcuts="Control+K Meta+K Control+F Meta+F"
          className="h-10 pl-10 pr-20"
        />
        <kbd className="pointer-events-none absolute right-3 top-1/2 hidden -translate-y-1/2 items-center rounded-md border border-border bg-card px-1.5 py-0.5 font-sans text-[10px] font-semibold text-muted-foreground min-[900px]:inline-flex">
          Ctrl K
        </kbd>
      </div>
      <Button type="button" variant="outline" onClick={props.onGenerateGithub} aria-label="生成 GitHub 账号" className="h-10 shrink-0" title="生成 GitHub 账号草稿">
        <WandSparkles className="size-4" aria-hidden="true" />
        GitHub 生成
      </Button>
      <Button onClick={props.onCreate} aria-label="新增账号" aria-keyshortcuts="Control+N Meta+N" className="h-10 shrink-0 sm:min-w-30" title="新增账号（Ctrl+N）">
        <Plus className="size-4" aria-hidden="true" />
        新增账号
      </Button>
    </section>
  );
}
