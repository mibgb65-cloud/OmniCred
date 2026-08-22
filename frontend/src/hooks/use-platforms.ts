import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { createPlatform, deletePlatform, listPlatforms, updatePlatform } from "@/api/client";
import type { PlatformInput } from "@/api/types";

const platformKey = ["platforms"] as const;

export function usePlatforms() {
  return useQuery({ queryKey: platformKey, queryFn: listPlatforms });
}

export function usePlatformMutations() {
  const queryClient = useQueryClient();
  const refresh = () => Promise.all([
    queryClient.invalidateQueries({ queryKey: platformKey }),
    queryClient.invalidateQueries({ queryKey: ["credentials"] }),
  ]);

  const create = useMutation({ mutationFn: createPlatform, onSuccess: refresh });
  const update = useMutation({
    mutationFn: ({ id, input }: { id: number; input: PlatformInput }) => updatePlatform(id, input),
    onSuccess: refresh,
  });
  const remove = useMutation({ mutationFn: deletePlatform, onSuccess: refresh });

  return { create, update, remove };
}
