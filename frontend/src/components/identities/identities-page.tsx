import { useDeferredValue, useRef, useState } from "react";
import { Plus, Search } from "lucide-react";
import { toast } from "sonner";

import type { IdentityProfile, IdentityProfileInput } from "@/api/types";
import { DeleteIdentityDialog } from "@/components/identities/delete-identity-dialog";
import { IdentityFormDialog } from "@/components/identities/identity-form-dialog";
import { IdentityGrid } from "@/components/identities/identity-grid";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useIdentities, useIdentityMutations } from "@/hooks/use-identities";

export function IdentitiesPage() {
  const [search, setSearch] = useState("");
  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<IdentityProfile | null>(null);
  const [deleting, setDeleting] = useState<IdentityProfile | null>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const deferredSearch = useDeferredValue(search.trim());
  const profiles = useIdentities({ query: deferredSearch });
  const mutations = useIdentityMutations();
  const items = profiles.data?.items ?? [];

  function createNew() {
    setEditing(null);
    setFormOpen(true);
  }

  function edit(profile: IdentityProfile) {
    setEditing(profile);
    setFormOpen(true);
  }

  async function save(input: IdentityProfileInput) {
    try {
      if (editing) {
        await mutations.update.mutateAsync({ id: editing.id, input });
        toast.success("身份资料已更新");
      } else {
        await mutations.create.mutateAsync(input);
        toast.success("身份资料已保存到本机");
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
      toast.success("身份资料已删除");
      setDeleting(null);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除失败，请重试");
    }
  }

  const saving = mutations.create.isPending || mutations.update.isPending;

  return (
    <main id="main-content" className="flex min-h-0 min-w-0 flex-1 flex-col" tabIndex={-1}>
      <header className="shrink-0 border-b border-border bg-background/92 px-4 py-4 sm:px-5 lg:px-6">
        <div className="flex flex-col gap-4 min-[1000px]:flex-row min-[1000px]:items-center">
          <div className="min-w-52 shrink-0">
            <div className="flex items-center gap-2.5">
              <h1 className="text-lg font-bold tracking-tight">身份资料</h1>
              {!profiles.isLoading && <Badge variant="outline">{items.length} 条</Badge>}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">按国家保存姓名、地址、联系方式和登录信息。</p>
          </div>
          <section aria-label="身份资料搜索和操作" className="flex min-w-0 flex-1 flex-col gap-2.5 sm:flex-row sm:justify-end">
            <div className="relative min-w-0 flex-1 min-[1000px]:max-w-xl">
              <Search className="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
              <Input
                ref={searchInputRef}
                type="search"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="搜索姓名、国家、城市、电话或邮箱"
                aria-label="搜索身份资料"
                className="h-10 pl-10"
              />
            </div>
            <Button type="button" onClick={createNew} aria-label="新增身份资料" className="h-10 shrink-0 sm:min-w-36">
              <Plus className="size-4" aria-hidden="true" />
              新增身份资料
            </Button>
          </section>
        </div>
      </header>

      <section className="desktop-scrollbar relative min-h-0 flex-1 overflow-y-auto px-4 py-4 sm:px-5 sm:py-5 lg:px-6" aria-label="身份资料列表" aria-busy={profiles.isFetching}>
        {profiles.isFetching && !profiles.isLoading && (
          <div className="pointer-events-none absolute inset-x-0 top-0 h-0.5 overflow-hidden" aria-hidden="true">
            <div className="query-progress h-full w-1/3 rounded-full bg-primary" />
          </div>
        )}
        <IdentityGrid
          items={items}
          isLoading={profiles.isLoading}
          error={profiles.error}
          isFiltered={Boolean(deferredSearch)}
          onRetry={() => profiles.refetch()}
          onCreate={createNew}
          onEdit={edit}
          onDelete={setDeleting}
        />
      </section>

      <IdentityFormDialog open={formOpen} profile={editing} pending={saving} onOpenChange={setFormOpen} onSubmit={save} />
      <DeleteIdentityDialog profile={deleting} pending={mutations.remove.isPending} onOpenChange={(open) => !open && setDeleting(null)} onConfirm={remove} />
    </main>
  );
}
