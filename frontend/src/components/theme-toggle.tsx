import { Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { setWindowTheme } from "@/lib/desktop-runtime";

type Theme = "light" | "dark";

function preferredTheme(): Theme {
  const saved = localStorage.getItem("omnicred-theme");
  if (saved === "light" || saved === "dark") return saved;
  return window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
}

export function ThemeToggle() {
  const [theme, setTheme] = useState<Theme>(preferredTheme);

  useEffect(() => {
    document.documentElement.classList.toggle("dark", theme === "dark");
    localStorage.setItem("omnicred-theme", theme);
    setWindowTheme(theme);
  }, [theme]);

  const next = theme === "dark" ? "light" : "dark";
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          type="button"
          variant="outline"
          size="icon"
          onClick={() => setTheme(next)}
          aria-label={`切换到${next === "dark" ? "暗色" : "亮色"}主题`}
        >
          {theme === "dark" ? <Sun className="size-4.5" aria-hidden="true" /> : <Moon className="size-4.5" aria-hidden="true" />}
        </Button>
      </TooltipTrigger>
      <TooltipContent>切换主题</TooltipContent>
    </Tooltip>
  );
}
