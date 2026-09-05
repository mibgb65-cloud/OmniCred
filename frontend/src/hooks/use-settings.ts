import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { checkUpdates, getSettingsStatus, migrateStorage } from "@/api/client";
import type { UpdateInfo } from "@/api/types";

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
  const client = useQueryClient();
  const key = ["update-info"];
  const cached = useQuery<UpdateInfo>({ queryKey: key, queryFn: checkUpdates, enabled: false });
  const mutation = useMutation({ mutationFn: checkUpdates, onSuccess: (info) => client.setQueryData(key, info) });
  return { ...mutation, data: cached.data };
}
