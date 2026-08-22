import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { checkUpdates, getSettingsStatus, migrateStorage } from "@/api/client";

const settingsKey = ["settings-status"] as const;

export function useSettingsStatus() {
  return useQuery({ queryKey: settingsKey, queryFn: getSettingsStatus, refetchInterval: 30_000 });
}

export function useStorageMigration() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: migrateStorage,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: settingsKey }),
  });
}

export function useUpdateCheck() {
  return useMutation({ mutationFn: checkUpdates });
}
