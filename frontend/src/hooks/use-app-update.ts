import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { cancelUpdate, downloadUpdate, getUpdateState, hasUpdateBridge, restartToUpdate } from "@/lib/desktop-runtime";

const updateKey = ["app-update-state"] as const;

export function useAppUpdate() {
  const client = useQueryClient();
  const state = useQuery({
    queryKey: updateKey,
    queryFn: getUpdateState,
    enabled: hasUpdateBridge(),
    refetchInterval: (query) => {
      const phase = query.state.data?.phase;
      if (phase === "installing") return false;
      return phase === "downloading" || phase === "verifying" ? 750 : 5_000;
    },
    refetchIntervalInBackground: true,
    retry: false,
  });
  const download = useMutation({ mutationFn: downloadUpdate, onSuccess: (data) => client.setQueryData(updateKey, data) });
  const cancel = useMutation({ mutationFn: cancelUpdate, onSuccess: () => client.invalidateQueries({ queryKey: updateKey }) });
  const install = useMutation({ mutationFn: restartToUpdate, onSettled: () => client.invalidateQueries({ queryKey: updateKey }) });
  return { state, download, cancel, install };
}
