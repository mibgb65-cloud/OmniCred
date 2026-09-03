import { RefreshCw, SearchX, UserRound } from "lucide-react";

import type { IdentityProfile } from "@/api/types";
import { IdentityCard } from "@/components/identities/identity-card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

interface IdentityGridProps {
  items: IdentityProfile[];
  isLoading: boolean;
  error: Error | null;
  isFiltered: boolean;
  onRetry: () => void;
  onCreate: () => void;
  onEdit: (profile: IdentityProfile) => void;
  onDelete: (profile: IdentityProfile) => void;
}

const gridClassName = "grid gap-3.5 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 min-[2300px]:grid-cols-5";

export function IdentityGrid(props: IdentityGridProps) {
  if (props.isLoading) {
    return (
      <div className={gridClassName} aria-label="正在加载身份资料">
        {[0, 1, 2, 3].map((item) => (
          <div key={item} className="rounded-xl border border-border bg-card p-4">
            <div className="flex gap-3"><Skeleton className="size-10" /><div className="flex-1 space-y-2"><Skeleton className="h-4 w-28" /><Skeleton className="h-3 w-36" /></div></div>
            <Skeleton className="mt-4 h-16 w-full" /><Skeleton className="mt-2.5 h-16 w-full" />
          </div>
        ))}
      </div>
    );
  }

  if (props.error) {
    return (
      <div className="grid min-h-64 place-items-center rounded-xl border border-destructive/20 bg-destructive/5 p-8 text-center" role="alert">
        <div><RefreshCw className="mx-auto size-7 text-destructive" aria-hidden="true" /><h2 className="mt-4 text-lg font-bold">无法读取身份资料</h2><p className="mt-2 max-w-md text-sm text-muted-foreground">{props.error.message}</p><Button variant="outline" className="mt-5" onClick={props.onRetry}>重新加载</Button></div>
      </div>
    );
  }

  if (props.items.length === 0) {
    return (
      <div className="grid min-h-64 place-items-center rounded-xl border border-dashed border-border bg-card/50 p-8 text-center">
        <div>
          <div className="mx-auto grid size-14 place-items-center rounded-2xl bg-primary/10 text-primary">
            {props.isFiltered ? <SearchX className="size-6" aria-hidden="true" /> : <UserRound className="size-6" aria-hidden="true" />}
          </div>
          <h2 className="mt-5 text-lg font-bold">{props.isFiltered ? "没有匹配的身份资料" : "还没有保存身份资料"}</h2>
          <p className="mt-2 text-sm text-muted-foreground">{props.isFiltered ? "换一个姓名、国家或联系方式试试。" : "创建第一条身份资料，统一保存姓名、地址和联系方式。"}</p>
          {!props.isFiltered && <Button className="mt-5" onClick={props.onCreate}>新增第一份资料</Button>}
        </div>
      </div>
    );
  }

  return (
    <div className={gridClassName}>
      {props.items.map((profile) => <IdentityCard key={profile.id} profile={profile} onEdit={props.onEdit} onDelete={props.onDelete} />)}
    </div>
  );
}
