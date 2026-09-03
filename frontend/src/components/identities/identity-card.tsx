import { CalendarDays, Check, Copy, Eye, EyeOff, Mail, MapPin, Pencil, Phone, Trash2, UserRound } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import type { IdentityProfile } from "@/api/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

interface IdentityCardProps {
  profile: IdentityProfile;
  onEdit: (profile: IdentityProfile) => void;
  onDelete: (profile: IdentityProfile) => void;
}

const genderLabels: Record<string, string> = { male: "Male（男）", female: "Female（女）", other: "Other（其他）" };

export function IdentityCard({ profile, onEdit, onDelete }: IdentityCardProps) {
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [copied, setCopied] = useState<"email" | "phone" | "password" | "address" | null>(null);
  const nameParts = [profile.first_name, profile.middle_name, profile.last_name].filter(Boolean).join(" · ");
  const address = [profile.street_address, profile.city, profile.region, profile.postal_code].filter(Boolean).join(", ");

  async function copy(value: string, field: NonNullable<typeof copied>) {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(field);
      toast.success(field === "password" ? "密码已复制" : field === "address" ? "地址已复制" : field === "phone" ? "电话已复制" : "邮箱已复制");
      window.setTimeout(() => setCopied(null), 1500);
    } catch {
      toast.error("复制失败，请手动选择内容");
    }
  }

  return (
    <article className="group relative overflow-hidden rounded-xl border border-border bg-card p-4 shadow-sm transition-[border-color,box-shadow] duration-200 hover:border-primary/25 hover:shadow-md hover:shadow-slate-950/5 dark:hover:shadow-black/15">
      <div className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary/45 to-transparent opacity-0 transition-opacity group-hover:opacity-100" />
      <header className="flex items-start gap-3.5">
        <div className="grid size-10 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
          <UserRound className="size-5" aria-hidden="true" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h2 className="overflow-wrap-anywhere font-bold text-foreground">{profile.full_name}</h2>
              {profile.localized_name && <p className="mt-0.5 overflow-wrap-anywhere text-sm text-muted-foreground">{profile.localized_name}</p>}
            </div>
            <Badge variant="outline" className="shrink-0">{profile.country}</Badge>
          </div>
        </div>
      </header>

      <div className="mt-4 space-y-2.5">
        {(nameParts || profile.gender || profile.birth_date) && (
          <div className="rounded-lg border border-border/75 bg-background/55 p-3">
            {nameParts && <p className="overflow-wrap-anywhere text-sm font-semibold text-foreground">{nameParts}</p>}
            <div className="mt-1.5 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
              {profile.gender && <span>{genderLabels[profile.gender] ?? profile.gender}</span>}
              {profile.birth_date && <span className="flex items-center gap-1.5"><CalendarDays className="size-3.5" aria-hidden="true" />{profile.birth_date}</span>}
            </div>
          </div>
        )}

        {address && (
          <div className="flex items-start gap-2.5 rounded-lg border border-border/75 bg-background/55 p-3">
            <MapPin className="mt-0.5 size-4 shrink-0 text-primary" aria-hidden="true" />
            <p className="min-w-0 flex-1 overflow-wrap-anywhere text-sm leading-5 text-foreground">{address}</p>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button type="button" variant="ghost" size="icon-sm" onClick={() => copy(address, "address")} aria-label="复制身份资料地址">
                  {copied === "address" ? <Check className="size-4 text-emerald-500" aria-hidden="true" /> : <Copy className="size-4" aria-hidden="true" />}
                </Button>
              </TooltipTrigger>
              <TooltipContent>复制地址</TooltipContent>
            </Tooltip>
          </div>
        )}

        {profile.phone && (
          <ContactRow icon={Phone} label="电话" value={profile.phone} copied={copied === "phone"} onCopy={() => copy(profile.phone, "phone")} />
        )}
        {profile.email && (
          <ContactRow icon={Mail} label="邮箱" value={profile.email} copied={copied === "email"} onCopy={() => copy(profile.email, "email")} />
        )}
        {profile.password && (
          <div className="rounded-lg border border-border/75 bg-background/55 p-3">
            <div className="flex items-center justify-between gap-3">
              <div className="min-w-0">
                <p className="text-[11px] font-bold uppercase tracking-[0.14em] text-muted-foreground">密码</p>
                <p className="mt-1.5 max-w-full truncate font-mono text-sm font-semibold tracking-[0.1em] text-foreground">
                  {passwordVisible ? profile.password : "••••••••••••"}
                </p>
              </div>
              <div className="flex shrink-0 gap-1">
                <Button type="button" variant="ghost" size="icon-sm" onClick={() => setPasswordVisible((value) => !value)} aria-label={passwordVisible ? "隐藏身份资料密码" : "显示身份资料密码"} aria-pressed={passwordVisible}>
                  {passwordVisible ? <EyeOff className="size-4" aria-hidden="true" /> : <Eye className="size-4" aria-hidden="true" />}
                </Button>
                <Button type="button" variant="ghost" size="icon-sm" onClick={() => copy(profile.password, "password")} aria-label="复制身份资料密码">
                  {copied === "password" ? <Check className="size-4 text-emerald-500" aria-hidden="true" /> : <Copy className="size-4" aria-hidden="true" />}
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>

      <footer className="mt-3.5 flex items-center justify-between gap-3 border-t border-border/70 pt-3.5">
        <span className="text-xs text-muted-foreground">资料 #{profile.id}</span>
        <div className="flex gap-1">
          <Button type="button" variant="ghost" size="icon-sm" onClick={() => onEdit(profile)} aria-label={`编辑 ${profile.full_name}`}>
            <Pencil className="size-4" aria-hidden="true" />
          </Button>
          <Button type="button" variant="ghost" size="icon-sm" onClick={() => onDelete(profile)} aria-label={`删除 ${profile.full_name}`} className="hover:bg-destructive/10 hover:text-destructive">
            <Trash2 className="size-4" aria-hidden="true" />
          </Button>
        </div>
      </footer>
    </article>
  );
}

interface ContactRowProps {
  icon: typeof Phone;
  label: string;
  value: string;
  copied: boolean;
  onCopy: () => void;
}

function ContactRow({ icon: Icon, label, value, copied, onCopy }: ContactRowProps) {
  return (
    <div className="flex items-center gap-2.5 rounded-lg border border-border/75 bg-background/55 p-3">
      <Icon className="size-4 shrink-0 text-primary" aria-hidden="true" />
      <div className="min-w-0 flex-1">
        <p className="text-[11px] font-bold uppercase tracking-[0.14em] text-muted-foreground">{label}</p>
        <p className="mt-1 overflow-wrap-anywhere text-sm font-semibold text-foreground">{value}</p>
      </div>
      <Button type="button" variant="ghost" size="icon-sm" onClick={onCopy} aria-label={`复制${label}`}>
        {copied ? <Check className="size-4 text-emerald-500" aria-hidden="true" /> : <Copy className="size-4" aria-hidden="true" />}
      </Button>
    </div>
  );
}
