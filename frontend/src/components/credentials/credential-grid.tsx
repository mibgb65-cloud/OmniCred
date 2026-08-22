import { useGSAP } from "@gsap/react";
import { gsap } from "gsap";
import { KeyRound, RefreshCw, SearchX } from "lucide-react";
import { useRef } from "react";

import type { Credential } from "@/api/types";
import { CredentialCard } from "@/components/credentials/credential-card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

gsap.registerPlugin(useGSAP);

const gridClassName = "grid gap-3.5 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 min-[2300px]:grid-cols-5";

interface CredentialGridProps {
  items: Credential[];
  isLoading: boolean;
  error: Error | null;
  isFiltered: boolean;
  onRetry: () => void;
  onCreate: () => void;
  onEdit: (item: Credential) => void;
  onDelete: (item: Credential) => void;
}

export function CredentialGrid(props: CredentialGridProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const animationKey = props.isLoading
    ? "loading"
    : props.error
      ? `error:${props.error.message}`
      : props.items.map((item) => `${item.id}:${item.updated_at}`).join("|") || "empty";

  useGSAP(() => {
    const root = containerRef.current;
    if (!root) return;
    const cards = gsap.utils.toArray<HTMLElement>("[data-credential-card]", root).slice(0, 18);
    const targets: gsap.TweenTarget = cards.length > 0 ? cards : root;

    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      gsap.set(targets, { clearProps: "opacity,transform,willChange" });
      return;
    }

    gsap.fromTo(
      targets,
      { opacity: 0, y: cards.length > 0 ? 8 : 4 },
      {
        opacity: 1,
        y: 0,
        duration: 0.28,
        ease: "power1.out",
        stagger: cards.length > 0 ? 0.03 : 0,
        overwrite: "auto",
        willChange: "transform,opacity",
        clearProps: "opacity,transform,willChange",
      },
    );
  }, { scope: containerRef, dependencies: [animationKey], revertOnUpdate: true });

  if (props.isLoading) {
    return (
      <div ref={containerRef} className={gridClassName} aria-label="正在加载账号">
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
      <div ref={containerRef} className="grid min-h-64 place-items-center rounded-xl border border-destructive/20 bg-destructive/5 p-8 text-center" role="alert">
        <div><RefreshCw className="mx-auto size-7 text-destructive" aria-hidden="true" /><h2 className="mt-4 text-lg font-bold">无法读取本地账号</h2><p className="mt-2 max-w-md text-sm text-muted-foreground">{props.error.message}</p><Button variant="outline" className="mt-5" onClick={props.onRetry}>重新加载</Button></div>
      </div>
    );
  }

  if (props.items.length === 0) {
    return (
      <div ref={containerRef} className="grid min-h-64 place-items-center rounded-xl border border-dashed border-border bg-card/50 p-8 text-center">
        <div>
          <div className="mx-auto grid size-14 place-items-center rounded-2xl bg-primary/10 text-primary">
            {props.isFiltered ? <SearchX className="size-6" aria-hidden="true" /> : <KeyRound className="size-6" aria-hidden="true" />}
          </div>
          <h2 className="mt-5 text-lg font-bold">{props.isFiltered ? "没有匹配的账号" : "还没有保存账号"}</h2>
          <p className="mt-2 text-sm text-muted-foreground">{props.isFiltered ? "换一个关键词或平台试试。" : "创建第一条本地凭据，开始使用 OmniCred。"}</p>
          {!props.isFiltered && <Button className="mt-5" onClick={props.onCreate}>新增第一个账号</Button>}
        </div>
      </div>
    );
  }

  return (
    <div ref={containerRef} className={gridClassName}>
      {props.items.map((item) => <CredentialCard key={item.id} item={item} onEdit={props.onEdit} onDelete={props.onDelete} />)}
    </div>
  );
}
