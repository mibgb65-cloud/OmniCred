import * as React from "react";

import { cn } from "@/lib/utils";

function Textarea({ className, ...props }: React.ComponentProps<"textarea">) {
  return (
    <textarea
      className={cn(
        "scrollbar-hidden min-h-24 w-full resize-y rounded-xl border border-input bg-background/70 px-3.5 py-3 text-sm leading-6 text-foreground shadow-sm outline-none transition-[border-color,box-shadow,background-color] placeholder:text-muted-foreground/70 focus-visible:border-primary/60 focus-visible:ring-2 focus-visible:ring-ring/25 disabled:cursor-not-allowed disabled:opacity-50",
        className,
      )}
      {...props}
    />
  );
}

export { Textarea };
