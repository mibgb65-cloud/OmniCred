import { Check, LoaderCircle, Pencil, Plus, RefreshCw, Trash2, X } from "lucide-react";
import { useEffect, useState, type FormEvent } from "react";
import { toast } from "sonner";

import { APIError } from "@/api/client";
import type { Platform } from "@/api/types";
import { ProviderMark, providerLabel } from "@/components/credentials/provider-mark";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { usePlatformMutations } from "@/hooks/use-platforms";

interface PlatformManagerDialogProps {
  open: boolean;
  platforms: Platform[];
  loading: boolean;
  error: Error | null;
  onOpenChange: (open: boolean) => void;
  onRetry: () => void;
  onRenamed: (previousName: string, nextName: string) => void;
  onDeleted: (name: string) => void;
}

export function PlatformManagerDialog(props: PlatformManagerDialogProps) {
  const mutations = usePlatformMutations();
  const [name, setName] = useState("");
  const [nameError, setNameError] = useState("");
  const [editing, setEditing] = useState<Platform | null>(null);
  const [editName, setEditName] = useState("");
  const [editError, setEditError] = useState("");
  const [deleting, setDeleting] = useState<Platform | null>(null);
  const busy = mutations.create.isPending || mutations.update.isPending || mutations.remove.isPending;

  useEffect(() => {
    if (props.open) return;
    setName("");
    setNameError("");
    setEditing(null);
    setDeleting(null);
  }, [props.open]);

  async function addPlatform(event: FormEvent) {
    event.preventDefault();
    const value = validateName(name, setNameError);
    if (!value) return;
    try {
      await mutations.create.mutateAsync({ name: value });
      setName("");
      toast.success("平台已添加");
    } catch (error) {
      setNameError(platformErrorMessage(error));
    }
  }

  function beginEdit(platform: Platform) {
    setEditing(platform);
    setEditName(platform.name);
    setEditError("");
  }

  async function saveEdit(event: FormEvent) {
    event.preventDefault();
    if (!editing) return;
    const value = validateName(editName, setEditError);
    if (!value) return;
    try {
      const updated = await mutations.update.mutateAsync({ id: editing.id, input: { name: value } });
      props.onRenamed(editing.name, updated.name);
      setEditing(null);
      toast.success("平台名称已更新");
    } catch (error) {
      setEditError(platformErrorMessage(error));
    }
  }

  async function removePlatform() {
    if (!deleting) return;
    const target = deleting;
    try {
      await mutations.remove.mutateAsync(target.id);
      props.onDeleted(target.name);
      setDeleting(null);
      toast.success("平台已删除");
    } catch (error) {
      toast.error(platformErrorMessage(error));
    }
  }

  return (
    <>
      <Dialog open={props.open} onOpenChange={(open) => !busy && props.onOpenChange(open)}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle>管理账号平台</DialogTitle>
            <DialogDescription>添加平台供账号使用；重命名会同步更新该平台下的全部账号。</DialogDescription>
          </DialogHeader>

          <form className="rounded-xl border border-border bg-background/55 p-4" onSubmit={addPlatform} noValidate>
            <Label htmlFor="new-platform">新增平台</Label>
            <div className="mt-2 flex gap-2">
              <Input
                id="new-platform"
                value={name}
                onChange={(event) => { setName(event.target.value); setNameError(""); }}
                placeholder="例如 notion"
                autoComplete="off"
                aria-invalid={Boolean(nameError)}
                aria-describedby={nameError ? "new-platform-error" : undefined}
              />
              <Button type="submit" className="shrink-0" disabled={busy}>
                {mutations.create.isPending ? <LoaderCircle className="size-4 animate-spin" aria-hidden="true" /> : <Plus className="size-4" aria-hidden="true" />}
                添加
              </Button>
            </div>
            {nameError && <p id="new-platform-error" className="mt-2 text-xs font-medium text-destructive" role="alert">{nameError}</p>}
          </form>

          <section aria-label="已有平台" className="space-y-2">
            <div className="flex items-center justify-between gap-3">
              <h3 className="text-sm font-bold">已有平台</h3>
              <Badge variant="outline">{props.platforms.length} 个</Badge>
            </div>
            <div className="desktop-scrollbar max-h-72 space-y-2 overflow-y-auto pr-1">
              {props.loading && <p className="rounded-xl border border-border p-4 text-sm text-muted-foreground">正在读取平台…</p>}
              {props.error && (
                <div className="flex items-center justify-between gap-3 rounded-xl border border-destructive/20 bg-destructive/5 p-3" role="alert">
                  <p className="text-sm text-destructive">{props.error.message}</p>
                  <Button type="button" size="sm" variant="outline" onClick={props.onRetry}><RefreshCw className="size-3.5" aria-hidden="true" />重试</Button>
                </div>
              )}
              {!props.loading && !props.error && props.platforms.length === 0 && (
                <p className="rounded-xl border border-dashed border-border p-5 text-center text-sm text-muted-foreground">还没有账号平台。</p>
              )}
              {props.platforms.map((platform) => (
                editing?.id === platform.id
                  ? <EditRow key={platform.id} platform={platform} value={editName} error={editError} busy={busy} onChange={setEditName} onCancel={() => setEditing(null)} onSubmit={saveEdit} />
                  : (
                    <div key={platform.id} className="flex items-center gap-3 rounded-xl border border-border bg-card p-3">
                      <ProviderMark provider={platform.name} className="size-9 rounded-lg" />
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-bold">{providerLabel(platform.name)}</p>
                        <p className="mt-0.5 text-xs text-muted-foreground">
                          {platform.credential_count > 0 ? `${platform.credential_count} 个账号，需先移除账号才能删除` : "暂无关联账号"}
                        </p>
                      </div>
                      <Button type="button" variant="ghost" size="icon-sm" onClick={() => beginEdit(platform)} disabled={busy} aria-label={`修改 ${providerLabel(platform.name)}`}>
                        <Pencil className="size-4" aria-hidden="true" />
                      </Button>
                      <Button type="button" variant="ghost" size="icon-sm" onClick={() => setDeleting(platform)} disabled={busy || platform.credential_count > 0} aria-label={`删除 ${providerLabel(platform.name)}`} className="hover:bg-destructive/10 hover:text-destructive disabled:cursor-not-allowed">
                        <Trash2 className="size-4" aria-hidden="true" />
                      </Button>
                    </div>
                  )
              ))}
            </div>
          </section>
        </DialogContent>
      </Dialog>

      <AlertDialog open={deleting !== null} onOpenChange={(open) => !open && !mutations.remove.isPending && setDeleting(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除“{deleting ? providerLabel(deleting.name) : ""}”平台？</AlertDialogTitle>
            <AlertDialogDescription>平台会从筛选列表中移除。此操作无法撤销。</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={mutations.remove.isPending}>取消</AlertDialogCancel>
            <AlertDialogAction onClick={removePlatform} disabled={mutations.remove.isPending}>
              {mutations.remove.isPending && <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />}
              确认删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

function EditRow(props: { platform: Platform; value: string; error: string; busy: boolean; onChange: (value: string) => void; onCancel: () => void; onSubmit: (event: FormEvent) => void }) {
  const label = providerLabel(props.platform.name);
  return (
    <form className="rounded-xl border border-primary/25 bg-primary/5 p-3" onSubmit={props.onSubmit} noValidate>
      <Label htmlFor={`platform-${props.platform.id}`}>修改 {label}</Label>
      <div className="mt-2 flex gap-2">
        <Input id={`platform-${props.platform.id}`} value={props.value} onChange={(event) => props.onChange(event.target.value)} autoFocus aria-invalid={Boolean(props.error)} className="h-9" />
        <Button type="submit" size="icon-sm" disabled={props.busy} aria-label={`保存 ${label}`}><Check className="size-4" aria-hidden="true" /></Button>
        <Button type="button" size="icon-sm" variant="ghost" disabled={props.busy} onClick={props.onCancel} aria-label="取消修改"><X className="size-4" aria-hidden="true" /></Button>
      </div>
      {props.error && <p className="mt-2 text-xs font-medium text-destructive" role="alert">{props.error}</p>}
    </form>
  );
}

function validateName(value: string, setError: (message: string) => void) {
  const normalized = value.trim();
  if (!normalized) {
    setError("请输入平台名称");
    return "";
  }
  if ([...normalized].length > 100) {
    setError("平台名称不能超过 100 个字符");
    return "";
  }
  setError("");
  return normalized;
}

function platformErrorMessage(error: unknown) {
  if (error instanceof APIError) {
    if (error.code === "platform_exists") return "该平台名称已存在";
    if (error.code === "platform_in_use") return "该平台仍有关联账号，不能删除";
  }
  return error instanceof Error ? error.message : "操作失败，请重试";
}
