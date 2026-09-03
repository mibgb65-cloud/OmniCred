import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  createIdentityProfile,
  deleteIdentityProfile,
  listIdentityProfiles,
  updateIdentityProfile,
} from "@/api/client";
import type { IdentityProfileFilter, IdentityProfileInput } from "@/api/types";

const identityKey = ["identities"] as const;

export function useIdentities(filter: IdentityProfileFilter) {
  return useQuery({
    queryKey: [...identityKey, filter.country ?? "", filter.query ?? ""],
    queryFn: () => listIdentityProfiles(filter),
    placeholderData: (previousData) => previousData,
  });
}

export function useIdentityMutations() {
  const queryClient = useQueryClient();
  const refresh = () => queryClient.invalidateQueries({ queryKey: identityKey });

  const create = useMutation({ mutationFn: createIdentityProfile, onSuccess: refresh });
  const update = useMutation({
    mutationFn: ({ id, input }: { id: number; input: IdentityProfileInput }) => updateIdentityProfile(id, input),
    onSuccess: refresh,
  });
  const remove = useMutation({ mutationFn: deleteIdentityProfile, onSuccess: refresh });

  return { create, update, remove };
}
