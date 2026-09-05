import { useDeferredValue, useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import type { Credential, CredentialInput } from "@/api/types";
import { AppHeader } from "@/components/app-header";
import { CredentialFormDialog } from "@/components/credentials/credential-form-dialog";
import { CredentialGrid } from "@/components/credentials/credential-grid";
import { CredentialSidebar } from "@/components/credentials/credential-sidebar";
import { CredentialToolbar } from "@/components/credentials/credential-toolbar";
import { DeleteCredentialDialog } from "@/components/credentials/delete-credential-dialog";
import { GithubAccountGeneratorDialog } from "@/components/credentials/github-account-generator-dialog";
import { IdentitiesPage } from "@/components/identities/identities-page";
import { PlatformManagerDialog } from "@/components/platforms/platform-manager-dialog";
import { SettingsPage } from "@/components/settings/settings-page";
import { UpdateNotification } from "@/components/settings/update-notification";
import { Badge } from "@/components/ui/badge";
import { WindowTitlebar } from "@/components/window-titlebar";
import { useCredentialMutations, useCredentials } from "@/hooks/use-credentials";
import { usePlatforms } from "@/hooks/use-platforms";

export default function App() {
  const [page, setPage] = useState<"credentials" | "identities" | "settings">("credentials");
  const [search, setSearch] = useState("");
  const [provider, setProvider] = useState("");
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Credential | null>(null);
  const [deleting, setDeleting] = useState<Credential | null>(null);
  const [platformManagerOpen, setPlatformManagerOpen] = useState(false);
  const [githubGeneratorOpen, setGithubGeneratorOpen] = useState(false);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const deferredSearch = useDeferredValue(search.trim());
  const credentials = useCredentials({ provider, query: deferredSearch });
  const platforms = usePlatforms();
  const mutations = useCredentialMutations();
  const items = credentials.data?.items ?? [];
  const platformItems = platforms.data?.items ?? [];

  function createNew() {
    setPage("credentials");
    setEditing(null);
    setFormOpen(true);
  }

  function openIdentities() {
    setPage("identities");
  }

  function edit(item: Credential) {
    setPage("credentials");
    setEditing(item);
    setFormOpen(true);
  }

  useEffect(() => {
    function handleShortcut(event: KeyboardEvent) {
      if (!(event.ctrlKey || event.metaKey) || event.altKey || document.querySelector('[role="dialog"], [role="alertdialog"]')) return;
      const key = event.key.toLowerCase();
      if (key === "k" || key === "f") {
        event.preventDefault();
        setPage("credentials");
        window.setTimeout(() => {
          searchInputRef.current?.focus();
          searchInputRef.current?.select();
        });
      } else if (key === "n") {
        event.preventDefault();
        createNew();
      }
    }

    window.addEventListener("keydown", handleShortcut);
    return () => window.removeEventListener("keydown", handleShortcut);
  }, []);

  async function save(input: CredentialInput) {
    try {
      if (editing) {
        await mutations.update.mutateAsync({ id: editing.id, input });
        toast.success("账号已更新");
      } else {
        await mutations.create.mutateAsync(input);
        toast.success("账号已保存到本机");
      }
      setFormOpen(false);
      setEditing(null);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存失败，请重试");
    }
  }

  async function remove() {
    if (!deleting) return;
    try {
      await mutations.remove.mutateAsync(deleting.id);
      toast.success("账号记录已删除");
      setDeleting(null);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除失败，请重试");
    }
  }

  async function addGeneratedGithub(input: CredentialInput) {
    try {
      await mutations.create.mutateAsync(input);
      setGithubGeneratorOpen(false);
      toast.success("GitHub 账号已加入账号池");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存失败，请重试");
    }
  }

  async function saveBatch(inputs: CredentialInput[]) {
    try {
      await mutations.createMany.mutateAsync(inputs);
      setFormOpen(false);
      toast.success(`已导入 ${inputs.length} 个账号`);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "批量导入失败，请重试");
    }
  }

  const saving = mutations.create.isPending || mutations.createMany.isPending || mutations.update.isPending;

  return (
    <div className="flex h-dvh min-h-0 flex-col overflow-hidden bg-background">
      <WindowTitlebar />
      <UpdateNotification onOpenSettings={() => setPage("settings")} />
      <a href="#main-content" className="sr-only z-50 rounded-lg bg-card px-3 py-2 text-sm font-semibold focus:not-sr-only focus:absolute focus:left-3 focus:top-3 focus:ring-2 focus:ring-ring">
        跳到主要内容
      </a>
      <AppHeader />
      <div className="flex min-h-0 flex-1 flex-col min-[720px]:flex-row">
        <CredentialSidebar
          activePage={page}
          provider={provider}
          platforms={platformItems}
          resultCount={items.length}
          onProviderChange={(value) => { setProvider(value); setPage("credentials"); }}
          onManagePlatforms={() => setPlatformManagerOpen(true)}
          onOpenSettings={() => setPage("settings")}
          onOpenIdentities={openIdentities}
        />

        {page === "settings" ? <SettingsPage /> : page === "identities" ? <IdentitiesPage /> : <main id="main-content" className="flex min-h-0 min-w-0 flex-1 flex-col" tabIndex={-1}>
          <header className="shrink-0 border-b border-border bg-background/92 px-4 py-4 sm:px-5 lg:px-6">
            <div className="flex flex-col gap-4 min-[1100px]:flex-row min-[1100px]:items-center">
              <div className="min-w-52 shrink-0">
                <div className="flex items-center gap-2.5">
                  <h1 className="text-lg font-bold tracking-tight">账号凭据</h1>
                  {!credentials.isLoading && <Badge variant="outline">{items.length} 条</Badge>}
                </div>
                <p className="mt-1 text-xs text-muted-foreground">查找、复制并维护保存在本机的账号。</p>
              </div>
              <CredentialToolbar
                search={search}
                searchInputRef={searchInputRef}
                onSearchChange={setSearch}
                onGenerateGithub={() => setGithubGeneratorOpen(true)}
                onCreate={createNew}
              />
            </div>
          </header>

          <section
            className="desktop-scrollbar relative min-h-0 flex-1 overflow-y-auto px-4 py-4 sm:px-5 sm:py-5 lg:px-6"
            aria-label="账号列表"
            aria-busy={credentials.isFetching}
          >
            {credentials.isFetching && !credentials.isLoading && (
              <div className="pointer-events-none absolute inset-x-0 top-0 h-0.5 overflow-hidden" aria-hidden="true">
                <div className="query-progress h-full w-1/3 rounded-full bg-primary" />
              </div>
            )}
            <CredentialGrid
              items={items}
              isLoading={credentials.isLoading}
              error={credentials.error}
              isFiltered={Boolean(provider || deferredSearch)}
              onRetry={() => credentials.refetch()}
              onCreate={createNew}
              onEdit={edit}
              onDelete={setDeleting}
            />
          </section>
        </main>}
      </div>

      <CredentialFormDialog open={formOpen} credential={editing} initialProvider={provider} platforms={platformItems} pending={saving} onOpenChange={setFormOpen} onSubmit={save} onSubmitBatch={saveBatch} />
      <DeleteCredentialDialog credential={deleting} pending={mutations.remove.isPending} onOpenChange={(open) => !open && setDeleting(null)} onConfirm={remove} />
      <GithubAccountGeneratorDialog open={githubGeneratorOpen} pending={mutations.create.isPending} onOpenChange={setGithubGeneratorOpen} onConfirm={addGeneratedGithub} />
      <PlatformManagerDialog
        open={platformManagerOpen}
        platforms={platformItems}
        loading={platforms.isLoading}
        error={platforms.error}
        onOpenChange={setPlatformManagerOpen}
        onRetry={() => platforms.refetch()}
        onRenamed={(previousName, nextName) => {
          if (provider.toLowerCase() === previousName.toLowerCase()) setProvider(nextName);
        }}
        onDeleted={(name) => {
          if (provider.toLowerCase() === name.toLowerCase()) setProvider("");
        }}
      />
    </div>
  );
}
