import { zodResolver } from "@hookform/resolvers/zod";
import { ChevronDown, Eye, EyeOff, FileInput, KeyRound, ListChecks, LoaderCircle, ScanText, Trash2 } from "lucide-react";
import { useEffect, useState, type FormEvent } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

import type { Credential, CredentialInput, Platform } from "@/api/types";
import { providerLabel } from "@/components/credentials/provider-mark";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { parseCredentialEntries, type ParsedCredentialText } from "@/lib/credential-text-parser";
import { cn } from "@/lib/utils";

const schema = z.object({
  provider: z.string().trim().min(1, "请选择账号平台").max(100, "平台名称不能超过 100 个字符"),
  account: z.string().trim().min(1, "请输入登录账号").max(4096, "登录账号过长"),
  username: z.string().trim().max(4096, "用户名过长"),
  password: z.string().min(1, "请输入密码").max(16384, "密码过长"),
  totp_secret: z.string().max(512, "2FA 密钥过长").refine(
    (value) => !value.trim() || /^[A-Z2-7]+=*$/i.test(value.replace(/[\s-]/g, "")),
    "请输入有效的 Base32 2FA 密钥",
  ),
  recovery_codes: z.string().max(25700, "恢复码内容过长").refine(
    (value) => parseRecoveryCodes(value).length <= 100 && parseRecoveryCodes(value).every((code) => code.length <= 256),
    "最多保存 100 个恢复码，每个不超过 256 个字符",
  ).refine(
    (value) => new Set(parseRecoveryCodes(value)).size === parseRecoveryCodes(value).length,
    "恢复码不能重复",
  ),
});

type FormValues = z.infer<typeof schema>;

const emptyValues: FormValues = { provider: "", account: "", username: "", password: "", totp_secret: "", recovery_codes: "" };

interface CredentialFormDialogProps {
  open: boolean;
  credential: Credential | null;
  initialProvider: string;
  platforms: Platform[];
  pending: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (input: CredentialInput) => Promise<void>;
  onSubmitBatch: (inputs: CredentialInput[]) => Promise<void>;
}

type BatchRow = ParsedCredentialText & { account: string; password: string };

function FieldError({ id, message }: { id: string; message?: string }) {
  if (!message) return null;
  return <p id={id} className="text-xs font-medium text-destructive" role="alert">{message}</p>;
}

export function CredentialFormDialog(props: CredentialFormDialogProps) {
  const [showPassword, setShowPassword] = useState(false);
  const [showTOTPSecret, setShowTOTPSecret] = useState(false);
  const [recoveryCodesOpen, setRecoveryCodesOpen] = useState(false);
  const [parserOpen, setParserOpen] = useState(false);
  const [rawText, setRawText] = useState("");
  const [parserMessage, setParserMessage] = useState("");
  const [parserError, setParserError] = useState("");
  const [batchRows, setBatchRows] = useState<BatchRow[]>([]);
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: emptyValues });
  const editing = props.credential !== null;

  useEffect(() => {
    if (!props.open) return;
    const item = props.credential;
    form.reset(item ? {
      provider: item.provider,
      account: item.account,
      username: item.username,
      password: item.password,
      totp_secret: item.totp_secret,
      recovery_codes: item.recovery_codes.join("\n"),
    } : { ...emptyValues, provider: props.initialProvider });
    setShowPassword(false);
    setShowTOTPSecret(false);
    setRecoveryCodesOpen(false);
    setParserOpen(false);
    setRawText("");
    setParserMessage("");
    setParserError("");
    setBatchRows([]);
  }, [form, props.credential, props.initialProvider, props.open]);

  const busy = props.pending || form.formState.isSubmitting;

  function parseText() {
    const entries = parseCredentialEntries(rawText);
    if (entries.length > 1) {
      const complete = entries.filter((entry): entry is BatchRow => Boolean(entry.account && entry.password));
      if (complete.length !== entries.length) {
        setParserError("批量导入的每一行都必须包含邮箱和密码。");
        return;
      }
      setBatchRows(complete);
      setParserMessage(`${complete.length} 条待导入`);
      setParserError("");
      setRawText("");
      setParserOpen(false);
      form.clearErrors(["account", "username", "password"]);
      return;
    }
    const fields = Object.entries(entries[0] ?? {}) as Array<[keyof FormValues, string]>;
    if (fields.length === 0) {
      setParserError("没有识别到邮箱、账号、用户名或密码，请检查标签格式。");
      return;
    }
    for (const [field, value] of fields) {
      form.setValue(field, value, { shouldDirty: true, shouldTouch: true, shouldValidate: true });
    }
    setParserMessage(`已解析并填入 ${fields.length} 项`);
    setParserError("");
    setRawText("");
    setParserOpen(false);
  }

  function submitForm(event: FormEvent<HTMLFormElement>) {
    if (batchRows.length === 0) {
      void form.handleSubmit(
        (values) => props.onSubmit({ ...values, recovery_codes: parseRecoveryCodes(values.recovery_codes) }),
        (errors) => { if (errors.recovery_codes) setRecoveryCodesOpen(true); },
      )(event);
      return;
    }
    event.preventDefault();
    const provider = form.getValues("provider").trim();
    if (!provider) {
      form.setError("provider", { message: "请选择账号平台" });
      form.setFocus("provider");
      return;
    }
    void props.onSubmitBatch(batchRows.map((row) => ({
      provider,
      account: row.account,
      username: row.username ?? "",
      password: row.password,
      totp_secret: row.totp_secret ?? "",
      recovery_codes: [],
    })));
  }

  function removeBatchRow(index: number) {
    setBatchRows((rows) => rows.filter((_, rowIndex) => rowIndex !== index));
  }

  return (
    <Dialog open={props.open} onOpenChange={(open) => !busy && props.onOpenChange(open)}>
      <DialogContent className="grid max-h-[90dvh] grid-rows-[auto_minmax(0,1fr)] gap-0 overflow-hidden p-0 sm:p-0">
        <DialogHeader className="shrink-0 border-b border-border bg-card px-6 py-5 pr-14 sm:px-7 sm:pr-14">
          <DialogTitle>{editing ? "编辑账号" : "新增账号"}</DialogTitle>
          <DialogDescription>
            {editing ? "修改后将直接覆盖本地记录。" : "信息只会写入当前电脑上的 OmniCred 数据库。"}
          </DialogDescription>
        </DialogHeader>

        <form className="flex min-h-0 flex-col" onSubmit={submitForm} noValidate>
          <div className="scrollbar-hidden min-h-0 flex-1 space-y-4 overflow-y-auto overscroll-contain px-6 py-5 sm:px-7">
          {!editing && (
            <section className="overflow-hidden rounded-xl border border-border bg-muted/28" aria-label="账号文本解析">
              <button
                type="button"
                className="flex w-full cursor-pointer items-center gap-3 px-3.5 py-3 text-left transition-colors hover:bg-accent/60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
                aria-label="从文本快速解析"
                aria-expanded={parserOpen}
                aria-controls="credential-parser-panel"
                onClick={() => { setParserOpen((value) => !value); setParserError(""); }}
              >
                <span className="grid size-9 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary"><FileInput className="size-4" aria-hidden="true" /></span>
                <span className="min-w-0 flex-1">
                  <span className="block text-sm font-bold">从文本快速解析</span>
                  <span className="mt-0.5 block text-xs text-muted-foreground">支持字段标签或“账号----密码[----用户名]”多行格式</span>
                </span>
                {parserMessage && <Badge variant="success" className="shrink-0">{parserMessage}</Badge>}
                <ChevronDown className={cn("size-4 shrink-0 text-muted-foreground transition-transform", parserOpen && "rotate-180")} aria-hidden="true" />
              </button>
              {parserOpen && (
                <div id="credential-parser-panel" className="space-y-3 border-t border-border p-3.5 animate-in fade-in slide-in-from-top-1 duration-200">
                  <Label htmlFor="credential-text">账号文本</Label>
                  <Textarea
                    id="credential-text"
                    value={rawText}
                    onChange={(event) => { setRawText(event.target.value); setParserError(""); }}
                    placeholder={"邮箱：avery.parker27@example.com\n密码：T7#qL2$vN9!mX4@p\n用户名：AveryParker27\n\n或每行：账号----密码（用户名可选）"}
                    rows={4}
                    className="resize-none"
                    aria-invalid={Boolean(parserError)}
                    aria-describedby={parserError ? "credential-parser-error" : "credential-parser-helper"}
                  />
                  <div className="flex items-start justify-between gap-3">
                    <p id={parserError ? "credential-parser-error" : "credential-parser-helper"} className={cn("text-xs leading-5 text-muted-foreground", parserError && "font-medium text-destructive")} role={parserError ? "alert" : undefined}>
                      {parserError || "兼容中文/英文冒号和四横线分隔；密码特殊字符会原样保留。"}
                    </p>
                    <Button type="button" size="sm" onClick={parseText} disabled={!rawText.trim()} className="shrink-0"><ScanText className="size-3.5" aria-hidden="true" />解析并填入</Button>
                  </div>
                </div>
              )}
            </section>
          )}

          <Controller
            name="provider"
            control={form.control}
            render={({ field, fieldState }) => (
              <div className="space-y-2">
                <Label htmlFor="provider">平台 <span className="text-destructive">*</span></Label>
                <Select name={field.name} value={field.value} onValueChange={field.onChange} disabled={busy}>
                  <SelectTrigger ref={field.ref} id="provider" onBlur={field.onBlur} aria-invalid={fieldState.invalid} aria-describedby={fieldState.error ? "provider-error" : "provider-helper"}>
                    <SelectValue placeholder="选择账号平台" />
                  </SelectTrigger>
                  <SelectContent>
                    {props.platforms.map((platform) => <SelectItem key={platform.id} value={platform.name}>{providerLabel(platform.name)}</SelectItem>)}
                  </SelectContent>
                </Select>
                {!fieldState.error && <p id="provider-helper" className="text-xs text-muted-foreground">如需其他平台，请先通过左侧“管理”添加。</p>}
                <FieldError id="provider-error" message={fieldState.error?.message} />
              </div>
            )}
          />

          {batchRows.length > 0 && (
            <section className="overflow-hidden rounded-xl border border-primary/20 bg-primary/5" aria-label="批量导入预览">
              <div className="flex items-center justify-between gap-3 border-b border-primary/15 px-3.5 py-3">
                <div className="flex items-center gap-2 text-sm font-bold"><ListChecks className="size-4 text-primary" aria-hidden="true" />已解析 {batchRows.length} 条账号</div>
                <Button type="button" variant="ghost" size="sm" onClick={() => { setBatchRows([]); setParserMessage(""); }}>清除批量内容</Button>
              </div>
              <ul className="desktop-scrollbar max-h-48 divide-y divide-border/70 overflow-y-auto">
                {batchRows.map((row, index) => (
                  <li key={`${row.account}-${index}`} className="flex items-center gap-3 px-3.5 py-3">
                    <span className="grid size-7 shrink-0 place-items-center rounded-lg bg-primary/10 text-xs font-bold text-primary">{index + 1}</span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-semibold">{row.account}</p>
                      <p className="mt-0.5 truncate text-xs text-muted-foreground">{row.username || "未设置用户名"} · 密码已隐藏</p>
                    </div>
                    <Button type="button" variant="ghost" size="icon-sm" onClick={() => removeBatchRow(index)} aria-label={`移除 ${row.account}`} className="hover:bg-destructive/10 hover:text-destructive"><Trash2 className="size-4" aria-hidden="true" /></Button>
                  </li>
                ))}
              </ul>
            </section>
          )}

          {batchRows.length === 0 && <Controller
            name="account"
            control={form.control}
            render={({ field, fieldState }) => (
              <div className="space-y-2">
                <Label htmlFor="account">登录账号 <span className="text-destructive">*</span></Label>
                <Input {...field} id="account" placeholder="邮箱、手机号或登录 ID" autoComplete="off" aria-invalid={fieldState.invalid} aria-describedby={fieldState.error ? "account-error" : undefined} />
                <FieldError id="account-error" message={fieldState.error?.message} />
              </div>
            )}
          />}

          {batchRows.length === 0 && <Controller
            name="username"
            control={form.control}
            render={({ field, fieldState }) => (
              <div className="space-y-2">
                <Label htmlFor="username">用户名 <span className="font-normal text-muted-foreground">（可选）</span></Label>
                <Input {...field} id="username" placeholder="昵称或公开用户名" autoComplete="off" aria-invalid={fieldState.invalid} aria-describedby={fieldState.error ? "username-error" : undefined} />
                <FieldError id="username-error" message={fieldState.error?.message} />
              </div>
            )}
          />}

          {batchRows.length === 0 && <Controller
            name="password"
            control={form.control}
            render={({ field, fieldState }) => (
              <div className="space-y-2">
                <Label htmlFor="password">密码 <span className="text-destructive">*</span></Label>
                <div className="relative">
                  <Input {...field} id="password" type={showPassword ? "text" : "password"} placeholder="输入密码" autoComplete="new-password" className="pr-12" aria-invalid={fieldState.invalid} aria-describedby={fieldState.error ? "password-error" : undefined} />
                  <button type="button" onClick={() => setShowPassword((value) => !value)} className="absolute right-1 top-1 grid size-9 cursor-pointer place-items-center rounded-lg text-muted-foreground hover:bg-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" aria-label={showPassword ? "隐藏密码" : "显示密码"} aria-pressed={showPassword}>
                    {showPassword ? <EyeOff className="size-4" aria-hidden="true" /> : <Eye className="size-4" aria-hidden="true" />}
                  </button>
                </div>
                <FieldError id="password-error" message={fieldState.error?.message} />
              </div>
            )}
          />}

          {batchRows.length === 0 && <Controller
            name="totp_secret"
            control={form.control}
            render={({ field, fieldState }) => (
              <div className="space-y-2">
                <Label htmlFor="totp-secret">2FA 密钥 <span className="font-normal text-muted-foreground">（TOTP，可选）</span></Label>
                <div className="relative">
                  <Input {...field} id="totp-secret" type={showTOTPSecret ? "text" : "password"} placeholder="粘贴 Base32 setup key" autoComplete="off" className="pr-12 font-mono" aria-invalid={fieldState.invalid} aria-describedby={fieldState.error ? "totp-secret-error" : "totp-secret-helper"} />
                  <button type="button" onClick={() => setShowTOTPSecret((value) => !value)} className="absolute right-1 top-1 grid size-9 cursor-pointer place-items-center rounded-lg text-muted-foreground hover:bg-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" aria-label={showTOTPSecret ? "隐藏 2FA 密钥" : "显示 2FA 密钥"} aria-pressed={showTOTPSecret}>
                    {showTOTPSecret ? <EyeOff className="size-4" aria-hidden="true" /> : <Eye className="size-4" aria-hidden="true" />}
                  </button>
                </div>
                {!fieldState.error && <p id="totp-secret-helper" className="text-xs leading-5 text-muted-foreground">支持 GitHub 等平台常用的 SHA-1、6 位、30 秒 Base32 密钥。密钥会和密码一样明文保存在本地数据库。</p>}
                <FieldError id="totp-secret-error" message={fieldState.error?.message} />
              </div>
            )}
          />}

          {batchRows.length === 0 && <Controller
            name="recovery_codes"
            control={form.control}
            render={({ field, fieldState }) => {
              const count = parseRecoveryCodes(field.value).length;
              return (
                <section className="overflow-hidden rounded-xl border border-border bg-muted/28" aria-label="恢复码设置">
                  <button
                    type="button"
                    className="flex w-full cursor-pointer items-center gap-3 px-3.5 py-3 text-left transition-colors hover:bg-accent/60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
                    aria-label="管理恢复码"
                    aria-expanded={recoveryCodesOpen}
                    aria-controls="recovery-codes-panel"
                    onClick={() => setRecoveryCodesOpen((value) => !value)}
                  >
                    <span className="grid size-9 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary"><KeyRound className="size-4" aria-hidden="true" /></span>
                    <span className="min-w-0 flex-1">
                      <span className="block text-sm font-bold">恢复码 <span className="font-normal text-muted-foreground">（可选）</span></span>
                      <span className="mt-0.5 block text-xs text-muted-foreground">一行一个，仅在失去其他 2FA 方式时使用</span>
                    </span>
                    {count > 0 && <Badge variant="outline" className="shrink-0">{count} 个</Badge>}
                    <ChevronDown className={cn("size-4 shrink-0 text-muted-foreground transition-transform", recoveryCodesOpen && "rotate-180")} aria-hidden="true" />
                  </button>
                  {recoveryCodesOpen && (
                    <div id="recovery-codes-panel" className="space-y-2 border-t border-border p-3.5 animate-in fade-in slide-in-from-top-1 duration-200">
                      <Label htmlFor="recovery-codes">恢复码列表</Label>
                      <Textarea {...field} id="recovery-codes" rows={5} placeholder={"alpha-bravo\ncharlie-delta"} className="resize-none font-mono" aria-invalid={fieldState.invalid} aria-describedby={fieldState.error ? "recovery-codes-error" : "recovery-codes-helper"} />
                      {!fieldState.error && <p id="recovery-codes-helper" className="text-xs leading-5 text-muted-foreground">恢复码通常只能使用一次。使用或重新生成后，请手动更新这里的列表；内容会明文保存在本地数据库。</p>}
                      <FieldError id="recovery-codes-error" message={fieldState.error?.message} />
                    </div>
                  )}
                  {!recoveryCodesOpen && <FieldError id="recovery-codes-error" message={fieldState.error?.message} />}
                </section>
              );
            }}
          />}

          </div>

          <DialogFooter className="shrink-0 border-t border-border bg-card/95 px-6 py-4 sm:px-7">
            <Button type="button" variant="outline" onClick={() => props.onOpenChange(false)} disabled={busy}>取消</Button>
            <Button type="submit" disabled={busy}>
              {busy && <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />}
              {batchRows.length > 0 ? `导入 ${batchRows.length} 个账号` : editing ? "保存修改" : "保存账号"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function parseRecoveryCodes(value: string) {
  return value.split(/\r?\n/).map((code) => code.trim()).filter(Boolean);
}
