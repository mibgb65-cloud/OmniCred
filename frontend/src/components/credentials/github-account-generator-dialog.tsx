import { Copy, Eye, EyeOff, LoaderCircle, RefreshCw, ShieldCheck, WandSparkles } from "lucide-react";
import { useEffect, useState, type Dispatch, type FormEvent, type ReactNode, type SetStateAction } from "react";
import { toast } from "sonner";
import { z } from "zod";

import type { CredentialInput } from "@/api/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { generateGithubUsername } from "@/lib/github-account-generator";
import { generateStrongPassword } from "@/lib/password-generator";

const emailSchema = z.string().trim().min(1, "请输入邮箱").email("请输入有效邮箱").max(254, "邮箱地址过长");
const usernamePattern = /^(?!.*--)[A-Za-z0-9](?:[A-Za-z0-9-]{0,37}[A-Za-z0-9])?$/;

interface GithubAccountGeneratorDialogProps {
  open: boolean;
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (input: CredentialInput) => Promise<void>;
}

export function GithubAccountGeneratorDialog(props: GithubAccountGeneratorDialogProps) {
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [generated, setGenerated] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (props.open) return;
    setEmail("");
    setUsername("");
    setPassword("");
    setGenerated(false);
    setShowPassword(false);
    setErrors({});
  }, [props.open]);

  function createDraft(event: FormEvent) {
    event.preventDefault();
    const parsed = emailSchema.safeParse(email);
    if (!parsed.success) {
      setErrors({ email: parsed.error.issues[0]?.message ?? "请输入有效邮箱" });
      return;
    }
    setEmail(parsed.data);
    setUsername(generateGithubUsername());
    setPassword(generateStrongPassword());
    setGenerated(true);
    setShowPassword(false);
    setErrors({});
  }

  function regenerateUsername() {
    setUsername(generateGithubUsername());
    setErrors((current) => ({ ...current, username: "" }));
  }

  async function confirmDraft(event: FormEvent) {
    event.preventDefault();
    const nextErrors: Record<string, string> = {};
    const parsedEmail = emailSchema.safeParse(email);
    if (!parsedEmail.success) nextErrors.email = parsedEmail.error.issues[0]?.message ?? "请输入有效邮箱";
    if (!usernamePattern.test(username)) nextErrors.username = "用户名需为 1–39 位字母、数字或单个连字符";
    if (password.length < 15) nextErrors.password = "密码至少需要 15 个字符";
    if (Object.keys(nextErrors).length > 0) {
      setErrors(nextErrors);
      return;
    }
    setErrors({});
    await props.onConfirm({ provider: "github", account: parsedEmail.success ? parsedEmail.data : email.trim(), username, password, totp_secret: "", recovery_codes: [] });
  }

  function regenerateAll() {
    regenerateUsername();
    setPassword(generateStrongPassword());
    setShowPassword(false);
  }

  async function copy(value: string, label: string) {
    try {
      await navigator.clipboard.writeText(value);
      toast.success(`${label}已复制`);
    } catch {
      toast.error("复制失败，请手动选择内容");
    }
  }

  return (
    <Dialog open={props.open} onOpenChange={(open) => !props.pending && props.onOpenChange(open)}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <div className="mb-1 flex items-center gap-2">
            <div className="grid size-9 place-items-center rounded-xl bg-slate-900 text-white dark:bg-white dark:text-slate-950">
              <WandSparkles className="size-4.5" aria-hidden="true" />
            </div>
            <Badge variant="outline">本地生成</Badge>
          </div>
          <DialogTitle>生成 GitHub 账号</DialogTitle>
          <DialogDescription>输入邮箱后，从内置英文姓名表组合用户名并生成强密码。草稿只有在你确认后才会写入账号池。</DialogDescription>
        </DialogHeader>

        <form className="space-y-4" onSubmit={generated ? confirmDraft : createDraft} noValidate>
          <Field label="邮箱" id="github-email" error={errors.email}>
            <Input id="github-email" type="email" value={email} onChange={(event) => { setEmail(event.target.value); clearError("email", setErrors); }} placeholder="avery.parker27@example.com" autoComplete="email" aria-invalid={Boolean(errors.email)} aria-describedby={errors.email ? "github-email-error" : undefined} />
          </Field>

          {!generated && (
            <Button type="submit" className="w-full">
              <WandSparkles className="size-4" aria-hidden="true" />
              生成待确认信息
            </Button>
          )}

          {generated && (
            <>
              <div className="flex items-center justify-between gap-3 rounded-xl border border-primary/20 bg-primary/6 px-3.5 py-3">
                <p className="flex items-center gap-2 text-xs font-semibold text-primary"><ShieldCheck className="size-4" aria-hidden="true" />待确认草稿，尚未保存</p>
                <Button type="button" variant="ghost" size="sm" onClick={regenerateAll}><RefreshCw className="size-3.5" aria-hidden="true" />重新生成</Button>
              </div>

              <Field label="用户名" id="github-username" error={errors.username} helper="英文名 + 姓氏 + 随机字母数字，需在 GitHub 注册时确认是否可用。">
                <div className="flex gap-2">
                  <Input id="github-username" value={username} onChange={(event) => { setUsername(event.target.value); clearError("username", setErrors); }} autoComplete="off" aria-invalid={Boolean(errors.username)} aria-describedby={errors.username ? "github-username-error" : "github-username-helper"} />
                  <Button type="button" variant="outline" size="icon" onClick={regenerateUsername} aria-label="重新生成用户名"><RefreshCw className="size-4" aria-hidden="true" /></Button>
                  <Button type="button" variant="outline" size="icon" onClick={() => copy(username, "用户名")} aria-label="复制用户名"><Copy className="size-4" aria-hidden="true" /></Button>
                </div>
              </Field>

              <Field label="密码" id="github-password" error={errors.password} helper="18 位随机强密码，仅保存在当前草稿中。">
                <div className="flex gap-2">
                  <Input id="github-password" type={showPassword ? "text" : "password"} value={password} onChange={(event) => { setPassword(event.target.value); clearError("password", setErrors); }} autoComplete="new-password" className="font-mono" aria-invalid={Boolean(errors.password)} aria-describedby={errors.password ? "github-password-error" : "github-password-helper"} />
                  <Button type="button" variant="outline" size="icon" onClick={() => setShowPassword((value) => !value)} aria-label={showPassword ? "隐藏生成密码" : "显示生成密码"} aria-pressed={showPassword}>{showPassword ? <EyeOff className="size-4" aria-hidden="true" /> : <Eye className="size-4" aria-hidden="true" />}</Button>
                  <Button type="button" variant="outline" size="icon" onClick={() => copy(password, "密码")} aria-label="复制生成密码"><Copy className="size-4" aria-hidden="true" /></Button>
                </div>
              </Field>

              <DialogFooter className="pt-1">
                <Button type="button" variant="outline" onClick={() => props.onOpenChange(false)} disabled={props.pending}>放弃草稿</Button>
                <Button type="submit" disabled={props.pending}>
                  {props.pending && <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />}
                  确认并加入账号池
                </Button>
              </DialogFooter>
            </>
          )}
        </form>
      </DialogContent>
    </Dialog>
  );
}

function Field(props: { label: string; id: string; error?: string; helper?: string; children: ReactNode }) {
  return (
    <div className="space-y-2">
      <Label htmlFor={props.id}>{props.label}</Label>
      {props.children}
      {props.error
        ? <p id={`${props.id}-error`} className="text-xs font-medium text-destructive" role="alert">{props.error}</p>
        : props.helper && <p id={`${props.id}-helper`} className="text-xs text-muted-foreground">{props.helper}</p>}
    </div>
  );
}

function clearError(field: string, setErrors: Dispatch<SetStateAction<Record<string, string>>>) {
  setErrors((current) => ({ ...current, [field]: "" }));
}
