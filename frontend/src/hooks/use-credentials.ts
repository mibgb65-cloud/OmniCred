import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { createCredential, deleteCredential, listCredentials, updateCredential } from "@/api/client";
import type { CredentialFilter, CredentialInput } from "@/api/types";

const credentialKey = ["credentials"] as const;

export function useCredentials(filter: CredentialFilter) {
  return useQuery({
    queryKey: [...credentialKey, filter.provider ?? "", filter.query ?? ""],
    queryFn: () => listCredentials(filter),
    placeholderData: (previousData) => previousData,
  });
}

export function useCredentialMutations() {
  const queryClient = useQueryClient();
  const refresh = () => Promise.all([
    queryClient.invalidateQueries({ queryKey: credentialKey }),
    queryClient.invalidateQueries({ queryKey: ["platforms"] }),
  ]);

  const create = useMutation({
    mutationFn: createCredential,
    onSuccess: refresh,
  });
  const createMany = useMutation({
    mutationFn: async (inputs: CredentialInput[]) => {
      const items = [];
      for (const input of inputs) items.push(await createCredential(input));
      return items;
    },
    onSuccess: refresh,
  });
  const update = useMutation({
    mutationFn: ({ id, input }: { id: number; input: CredentialInput }) => updateCredential(id, input),
    onSuccess: refresh,
  });
  const remove = useMutation({
    mutationFn: deleteCredential,
    onSuccess: refresh,
  });

  return { create, createMany, update, remove };
}
