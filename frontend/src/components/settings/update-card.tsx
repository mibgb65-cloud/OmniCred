import { CircleCheck, Download, ExternalLink, LoaderCircle, RefreshCw, RotateCw, X } from "lucide-react";
import { toast } from "sonner";

import type { RuntimeStatus } from "@/api/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useAppUpdate } from "@/hooks/use-app-update";
import { useUpdateCheck } from "@/hooks/use-settings";
import { hasUpdateBridge, openExternal } from "@/lib/desktop-runtime";

export function UpdateCard({ info }: { info: RuntimeStatus }) {
  const updates = useUpdateCheck();
  const { state, download, cancel, install } = useAppUpdate();
  const progress = state.data;
  const transferring = progress?.phase === "downloading" || progress?.phase === "verifying";
  const ready = progress?.phase === "ready";
  const installing = install.isPending || install.isSuccess || progress?.phase === "installing";
  const busy = transferring || download.isPending || installing;
  const percentage = progress?.total ? Math.min(100, Math.floor(progress.downloaded / progress.total * 100)) : 0;
  const error = progress?.error || install.error?.message || download.error?.message || cancel.error?.message || updates.error?.message || state.error?.message;

  async function perform(action: () => Promise<unknown>) {
    try { await action(); }
    catch (error) { toast.error(error instanceof Error ? error.message : "更新操作失败，请重试"); }
  }

  return (
    <section className="space-y-4 rounded-2xl border border-border bg-card p-5 shadow-sm" aria-labelledby="update-title">
      <div className="flex items-start gap-3">
        <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary"><RefreshCw className="size-5" aria-hidden="true" /></span>
        <div><h2 id="update-title" className="font-bold">版本与更新</h2><p className="mt-1 text-xs leading-5 text-muted-foreground">在应用内下载并校验，准备完成后由你决定何时重启。</p></div>
      </div>
      <div className="flex flex-wrap items-center justify-between gap-2 rounded-xl border border-border bg-background/55 p-3.5">
        <div><p className="text-xs text-muted-foreground">当前版本</p><p className="mt-1 font-mono text-lg font-bold">v{info.version}</p></div>
        {updates.data?.update_available && <Badge variant="success">发现 {updates.data.latest_version}</Badge>}
      </div>
      {updates.data?.status === "no_releases" && <p className="text-sm text-muted-foreground">仓库尚未发布正式 Release。</p>}
      {updates.data?.status === "ok" && !updates.data.update_available && <p className="flex items-center gap-2 text-sm text-muted-foreground"><CircleCheck className="size-4 text-primary" aria-hidden="true" />当前已是最新版本。</p>}
      {updates.data?.unavailable_reason && <p className="text-sm leading-6 text-muted-foreground">{updates.data.unavailable_reason}</p>}
      {transferring && (
        <div className="space-y-2 rounded-xl border border-border bg-background/55 p-3.5">
          <p role="status" className="flex items-center gap-2 text-sm"><LoaderCircle className="size-4 animate-spin motion-reduce:animate-none" aria-hidden="true" />{progress.phase === "verifying" ? "下载完成，正在校验安装包…" : "正在下载安装包，可继续使用应用"}</p>
          <div role="progressbar" aria-label="安装包下载进度" aria-valuemin={0} aria-valuemax={100} aria-valuenow={percentage} className="h-1.5 overflow-hidden rounded-full bg-muted">
            <div className="h-full origin-left rounded-full bg-primary" style={{ transform: `scaleX(${percentage / 100})` }} />
          </div>
          <p className="text-xs tabular-nums text-muted-foreground">{formatBytes(progress.downloaded)} / {formatBytes(progress.total)} · {percentage}%</p>
        </div>
      )}
      {ready && !installing && <p role="status" className="flex items-start gap-2 text-sm leading-6"><CircleCheck className="mt-1 size-4 shrink-0 text-primary" aria-hidden="true" />{progress.version} 已下载并通过 SHA-256 校验，可重启更新。</p>}
      {installing && <p role="status" className="text-sm leading-6 text-muted-foreground">正在准备更新，请完成可能出现的 Windows 管理员授权。随后应用将退出，安装完成后自动打开。</p>}
      {error && !installing && <p role="alert" className="rounded-xl border border-destructive/25 bg-destructive/5 p-3 text-sm text-destructive">{error}</p>}
      <div className="flex flex-wrap gap-2">
        {ready && <Button type="button" onClick={() => perform(() => install.mutateAsync())} disabled={installing}>
          {installing ? <LoaderCircle className="size-4 animate-spin motion-reduce:animate-none" aria-hidden="true" /> : <RotateCw className="size-4" aria-hidden="true" />}重启并更新
        </Button>}
        {updates.data?.download_available && hasUpdateBridge() && !ready && !transferring && !installing && <Button type="button" onClick={() => perform(() => download.mutateAsync())} disabled={download.isPending || updates.isPending}>
          <Download className="size-4" aria-hidden="true" />{progress?.phase === "error" ? "重新下载" : "下载更新"}
        </Button>}
        {transferring && <Button type="button" variant="outline" onClick={() => perform(() => cancel.mutateAsync())} disabled={cancel.isPending}><X className="size-4" aria-hidden="true" />取消下载</Button>}
        <Button type="button" variant="outline" onClick={() => perform(() => updates.mutateAsync())} disabled={busy || ready || updates.isPending}>
          {updates.isPending ? <LoaderCircle className="size-4 animate-spin motion-reduce:animate-none" aria-hidden="true" /> : <RefreshCw className="size-4" aria-hidden="true" />}检查更新
        </Button>
        <Button type="button" variant="ghost" onClick={() => openExternal(updates.data?.release_url || `${info.repository_url}/releases`)}>查看发布页<ExternalLink className="size-4" aria-hidden="true" /></Button>
      </div>
      {ready && <p className="text-xs leading-5 text-muted-foreground">重启前请保存正在编辑的内容。更新会保留账号数据库和设置；Windows 可能请求一次管理员授权。</p>}
    </section>
  );
}

function formatBytes(value: number) {
  return `${(value / 1024 / 1024).toFixed(1)} MB`;
}
