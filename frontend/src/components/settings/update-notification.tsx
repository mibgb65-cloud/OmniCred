import { useEffect, useRef } from "react";
import { toast } from "sonner";

import { useAppUpdate } from "@/hooks/use-app-update";

export function UpdateNotification({ onOpenSettings }: { onOpenSettings: () => void }) {
  const { state } = useAppUpdate();
  const previous = useRef<string | undefined>(undefined);
  useEffect(() => {
    const current = state.data;
    if (current?.phase === "ready" && previous.current !== "ready") {
      toast.success(`${current.version} 已下载并校验完成`, {
        description: "可在设置中重启并更新。",
        duration: 10_000,
        action: { label: "查看更新", onClick: onOpenSettings },
      });
    }
    if (current?.phase === "error" && previous.current !== "error") {
      toast.error(current.error || "更新下载失败，请在设置中重试");
    }
    previous.current = current?.phase;
  }, [state.data, onOpenSettings]);
  return null;
}
