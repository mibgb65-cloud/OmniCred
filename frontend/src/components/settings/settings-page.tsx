import {
  Activity,
  Clock3,
  Database,
  ExternalLink,
  FolderInput,
  GitBranch,
  HardDrive,
  LoaderCircle,
  Server,
  ShieldAlert,
  Trash2,
} from "lucide-react";
import { useState, type ReactNode } from "react";
import { toast } from "sonner";

import { UpdateCard } from "@/components/settings/update-card";
import { Badge } from "@/components/ui/badge";
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
import { Button } from "@/components/ui/button";
import { useSettingsStatus, useStorageMigration } from "@/hooks/use-settings";
import { chooseDatabasePath, openExternal, uninstallApplication } from "@/lib/desktop-runtime";

export function SettingsPage() {
  const [uninstallOpen, setUninstallOpen] = useState(false);
  const [uninstalling, setUninstalling] = useState(false);
  const status = useSettingsStatus();
  const storage = useStorageMigration();
  const info = status.data;

  async function selectStorage() {
    try {
      const path = await chooseDatabasePath();
      if (!path) return;
      const result = await storage.mutateAsync(path);
      toast.success(result.restart_required ? "数据库已复制，重启后使用新位置" : "数据位置已更新");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "迁移数据库失败");
    }
  }

  async function uninstall() {
    setUninstalling(true);
    try {
      await uninstallApplication();
    } catch (error) {
      setUninstalling(false);
      toast.error(error instanceof Error ? error.message : "无法启动卸载程序");
    }
  }

  return (
    <>
      <main id="main-content" className="scrollbar-hidden min-h-0 flex-1 overflow-y-auto px-5 py-6 lg:px-8 lg:py-8" tabIndex={-1}>
      <div className="mx-auto max-w-5xl">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">设置</h1>
            <p className="mt-1.5 text-sm text-muted-foreground">管理本地数据、版本更新和桌面运行状态。</p>
          </div>
          {info && <Badge variant={info.database_ok ? "success" : "outline"}><Activity className="size-3.5" aria-hidden="true" />{info.database_ok ? "运行正常" : "数据库异常"}</Badge>}
        </div>

        {status.isLoading && <div className="mt-8 flex items-center gap-2 text-sm text-muted-foreground"><LoaderCircle className="size-4 animate-spin" aria-hidden="true" />正在读取设置…</div>}
        {status.error && <div className="mt-8 rounded-xl border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive" role="alert">{status.error.message}</div>}

        {info && (
          <div className="mt-7 grid gap-4 lg:grid-cols-2">
            <SettingsCard icon={Database} title="数据存储" description="迁移会复制完整数据库，旧文件保留为备份。">
              <PathBlock label="当前数据库" value={info.database_path} />
              {info.pending_database_path && (
                <div className="rounded-xl border border-amber-500/25 bg-amber-500/8 p-3 text-sm text-amber-800 dark:text-amber-200">
                  <p className="font-semibold">重启后生效</p>
                  <p className="mt-1 overflow-wrap-anywhere font-mono text-xs">{info.pending_database_path}</p>
                </div>
              )}
              <Button type="button" variant="outline" onClick={selectStorage} disabled={storage.isPending || Boolean(info.pending_database_path)}>
                {storage.isPending ? <LoaderCircle className="size-4 animate-spin" aria-hidden="true" /> : <FolderInput className="size-4" aria-hidden="true" />}
                {info.pending_database_path ? "等待重启" : "迁移到新位置"}
              </Button>
            </SettingsCard>

            <UpdateCard info={info} />

            <SettingsCard icon={Activity} title="运行状态" description="所有服务仅在当前电脑上运行。">
              <dl className="grid grid-cols-2 gap-3">
                <StatusItem icon={Clock3} label="运行时长" value={formatUptime(info.uptime_seconds)} />
                <StatusItem icon={HardDrive} label="账号数量" value={`${info.credential_count} 条`} />
                <StatusItem icon={Server} label="本地 API" value={info.api_address} />
                <StatusItem icon={Database} label="SQLite" value={info.database_ok ? "已连接" : "异常"} />
              </dl>
            </SettingsCard>

            <SettingsCard icon={GitBranch} title="项目与安全" description="发布、反馈和版本记录均通过 GitHub 管理。">
              <button type="button" onClick={() => openExternal(info.repository_url)} className="flex w-full cursor-pointer items-center justify-between gap-3 rounded-xl border border-border bg-background/55 p-3.5 text-left transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
                <span className="min-w-0"><span className="block text-sm font-semibold">mibgb65-cloud/OmniCred</span><span className="mt-1 block truncate text-xs text-muted-foreground">{info.repository_url}</span></span>
                <ExternalLink className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
              </button>
              <div className="flex gap-2.5 rounded-xl border border-amber-500/20 bg-amber-500/8 p-3 text-amber-800 dark:text-amber-200">
                <ShieldAlert className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
                <p className="text-xs leading-5">当前数据库中的密码仍为明文。迁移或发布前，请勿将真实数据库文件提交到 GitHub。</p>
              </div>
            </SettingsCard>

            <section className="flex flex-col gap-4 rounded-2xl border border-destructive/25 bg-destructive/5 p-5 lg:col-span-2 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex items-start gap-3">
                <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-destructive/10 text-destructive"><Trash2 className="size-5" aria-hidden="true" /></span>
                <div>
                  <h2 className="font-bold">卸载 OmniCred</h2>
                  <p className="mt-1 text-xs leading-5 text-muted-foreground">
                    {info.uninstall_available ? "启动 Windows 卸载程序；本地数据库和设置默认保留。" : "当前为便携版或开发版，没有可用的安装器卸载入口。"}
                  </p>
                </div>
              </div>
              <Button type="button" variant="destructive" onClick={() => setUninstallOpen(true)} disabled={!info.uninstall_available} className="shrink-0">
                <Trash2 className="size-4" aria-hidden="true" />卸载应用
              </Button>
            </section>
          </div>
        )}
      </div>
      </main>

      <AlertDialog open={uninstallOpen} onOpenChange={(open) => !uninstalling && setUninstallOpen(open)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认卸载 OmniCred？</AlertDialogTitle>
            <AlertDialogDescription>应用将退出并打开 Windows 卸载程序。账号数据库和设置文件默认保留，之后重新安装仍可继续使用。</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={uninstalling}>取消</AlertDialogCancel>
            <AlertDialogAction onClick={uninstall} disabled={uninstalling}>
              {uninstalling && <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />}
              启动卸载程序
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

function SettingsCard(props: { icon: typeof Database; title: string; description: string; children: ReactNode }) {
  const Icon = props.icon;
  return (
    <section className="space-y-4 rounded-2xl border border-border bg-card p-5 shadow-sm">
      <div className="flex items-start gap-3"><span className="grid size-10 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary"><Icon className="size-5" aria-hidden="true" /></span><div><h2 className="font-bold">{props.title}</h2><p className="mt-1 text-xs leading-5 text-muted-foreground">{props.description}</p></div></div>
      {props.children}
    </section>
  );
}

function PathBlock({ label, value }: { label: string; value: string }) {
  return <div className="rounded-xl border border-border bg-background/55 p-3.5"><p className="text-xs text-muted-foreground">{label}</p><p className="mt-1.5 overflow-wrap-anywhere font-mono text-xs leading-5">{value}</p></div>;
}

function StatusItem(props: { icon: typeof Database; label: string; value: string }) {
  const Icon = props.icon;
  return <div className="rounded-xl border border-border bg-background/55 p-3"><dt className="flex items-center gap-1.5 text-xs text-muted-foreground"><Icon className="size-3.5" aria-hidden="true" />{props.label}</dt><dd className="mt-1.5 overflow-wrap-anywhere text-sm font-bold">{props.value}</dd></div>;
}

function formatUptime(seconds: number) {
  if (seconds < 60) return `${seconds} 秒`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)} 分钟`;
  return `${Math.floor(seconds / 3600)} 小时 ${Math.floor(seconds % 3600 / 60)} 分`;
}
