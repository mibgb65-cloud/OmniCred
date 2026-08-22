import "@fontsource-variable/plus-jakarta-sans";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { Toaster } from "sonner";

import App from "@/App";
import { TooltipProvider } from "@/components/ui/tooltip";
import "@/index.css";

const savedTheme = localStorage.getItem("omnicred-theme");
const useDark = savedTheme === "dark" || (savedTheme !== "light" && !window.matchMedia("(prefers-color-scheme: light)").matches);
document.documentElement.classList.toggle("dark", useDark);

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 10_000, refetchOnWindowFocus: false },
    mutations: { retry: false },
  },
});

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <TooltipProvider delayDuration={300}>
        <App />
        <Toaster richColors closeButton position="bottom-right" />
      </TooltipProvider>
    </QueryClientProvider>
  </StrictMode>,
);
