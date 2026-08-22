import { Check, Copy, Eye, EyeOff, Pencil, ShieldCheck, Trash2, UserRound } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import type { Credential } from "@/api/types";
import { ProviderMark, providerLabel } from "@/components/credentials/provider-mark";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

interface CredentialCardProps {
  item: Credential;
  totpCode?: string;
  totpPeriod: number;
  totpSecondsRemaining: number;
  onEdit: (item: Credential) => void;
  onDelete: (item: Credential) => void;
}

function formatUpdatedAt(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

export function CredentialCard({ item, totpCode, totpPeriod, totpSecondsRemaining, onEdit, onDelete }: CredentialCardProps) {
  const [visible, setVisible] = useState(false);
  const [copied, setCopied] = useState<"account" | "password" | "totp" | null>(null);

  async function copy(value: string, field: "account" | "password" | "totp") {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(field);
      toast.success(field === "password" ? "密码已复制" : field === "totp" ? "2FA 验证码已复制" : "账号已复制");
      window.setTimeout(() => setCopied(null), 1500);
    } catch {
      toast.error("复制失败，请手动选择内容");
    }
  }

  const formattedTOTP = totpCode ? `${totpCode.slice(0, 3)} ${totpCode.slice(3)}` : "--- ---";
  const totpProgress = Math.max(0, Math.min(1, totpSecondsRemaining / totpPeriod));

  return (
    <article data-credential-card className="group relative overflow-hidden rounded-xl border border-border bg-card p-4 shadow-sm transition-[border-color,box-shadow] duration-200 hover:border-primary/25 hover:shadow-md hover:shadow-slate-950/5 dark:hover:shadow-black/15">
      <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary/45 to-transparent opacity-0 transition-opacity group-hover:opacity-100" />
      <div className="flex items-start gap-3.5">
        <ProviderMark provider={item.provider} className="size-10 rounded-lg" />
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h3 className="truncate font-bold text-foreground">{providerLabel(item.provider)}</h3>
              <p className="mt-0.5 truncate text-sm text-muted-foreground">{item.username || "未设置用户名"}</p>
            </div>
            <Badge variant="outline" className="shrink-0 font-normal">
              #{item.id}
            </Badge>
          </div>
        </div>
      </div>

      <div className="mt-4 space-y-2.5">
        <div className="rounded-lg border border-border/75 bg-background/55 p-3">
          <div className="flex items-center justify-between gap-3">
            <div className="min-w-0">
              <p className="text-[11px] font-bold uppercase tracking-[0.16em] text-muted-foreground">登录账号</p>
              <p className="mt-1.5 overflow-wrap-anywhere text-sm font-semibold text-foreground">{item.account}</p>
            </div>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button type="button" variant="ghost" size="icon-sm" onClick={() => copy(item.account, "account")} aria-label="复制登录账号">
                  {copied === "account" ? <Check className="size-4 text-emerald-500" aria-hidden="true" /> : <Copy className="size-4" aria-hidden="true" />}
                </Button>
              </TooltipTrigger>
              <TooltipContent>复制账号</TooltipContent>
            </Tooltip>
          </div>
        </div>

        <div className="rounded-lg border border-border/75 bg-background/55 p-3">
          <div className="flex items-center justify-between gap-3">
            <div className="min-w-0">
              <p className="text-[11px] font-bold uppercase tracking-[0.16em] text-muted-foreground">密码</p>
              <p className="mt-1.5 max-w-full truncate font-mono text-sm font-semibold tracking-[0.12em] text-foreground">
                {visible ? item.password : "••••••••••••"}
              </p>
            </div>
            <div className="flex shrink-0 gap-1">
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button type="button" variant="ghost" size="icon-sm" onClick={() => setVisible((value) => !value)} aria-label={visible ? "隐藏密码" : "显示密码"} aria-pressed={visible}>
                    {visible ? <EyeOff className="size-4" aria-hidden="true" /> : <Eye className="size-4" aria-hidden="true" />}
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{visible ? "隐藏密码" : "显示密码"}</TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button type="button" variant="ghost" size="icon-sm" onClick={() => copy(item.password, "password")} aria-label="复制密码">
                    {copied === "password" ? <Check className="size-4 text-emerald-500" aria-hidden="true" /> : <Copy className="size-4" aria-hidden="true" />}
                  </Button>
                </TooltipTrigger>
                <TooltipContent>复制密码</TooltipContent>
              </Tooltip>
            </div>
          </div>
        </div>

        {item.totp_secret && (
          <div className="relative overflow-hidden rounded-lg border border-primary/25 bg-primary/6 p-3">
            <div className="flex items-center justify-between gap-3">
              <div className="min-w-0">
                <div className="flex items-center gap-1.5 text-[11px] font-bold uppercase tracking-[0.14em] text-primary">
                  <ShieldCheck className="size-3.5" aria-hidden="true" />
                  2FA 验证码
                </div>
                <p className="mt-1.5 font-mono text-lg font-bold tracking-[0.18em] text-foreground">{formattedTOTP}</p>
              </div>
              <div className="flex shrink-0 items-center gap-1.5">
                <span className="min-w-8 text-right text-xs font-semibold tabular-nums text-muted-foreground">
                  {totpSecondsRemaining > 0 ? `${totpSecondsRemaining}s` : "同步"}
                </span>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button type="button" variant="ghost" size="icon-sm" onClick={() => totpCode && copy(totpCode, "totp")} disabled={!totpCode} aria-label="复制 2FA 验证码">
                      {copied === "totp" ? <Check className="size-4 text-emerald-500" aria-hidden="true" /> : <Copy className="size-4" aria-hidden="true" />}
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>复制验证码</TooltipContent>
                </Tooltip>
              </div>
            </div>
            <span className="absolute inset-x-0 bottom-0 h-0.5 origin-left bg-primary transition-transform duration-200 motion-reduce:transition-none" style={{ transform: `scaleX(${totpProgress})` }} aria-hidden="true" />
          </div>
        )}
      </div>

      <footer className="mt-3.5 flex items-center justify-between gap-3 border-t border-border/70 pt-3.5">
        <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
          <UserRound className="size-3.5" aria-hidden="true" />
          更新于 {formatUpdatedAt(item.updated_at)}
        </span>
        <div className="flex gap-1">
          <Button type="button" variant="ghost" size="icon-sm" onClick={() => onEdit(item)} aria-label={`编辑 ${item.account}`}>
            <Pencil className="size-4" aria-hidden="true" />
          </Button>
          <Button type="button" variant="ghost" size="icon-sm" onClick={() => onDelete(item)} aria-label={`删除 ${item.account}`} className="hover:bg-destructive/10 hover:text-destructive">
            <Trash2 className="size-4" aria-hidden="true" />
          </Button>
        </div>
      </footer>
    </article>
  );
}
