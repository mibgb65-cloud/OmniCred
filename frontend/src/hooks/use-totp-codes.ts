import { useQuery } from "@tanstack/react-query";

import { listTOTPCodes } from "@/api/client";

export function useTOTPCodes(enabled: boolean) {
  return useQuery({
    queryKey: ["totp-codes"],
    queryFn: listTOTPCodes,
    enabled,
    refetchInterval: 1000,
    refetchIntervalInBackground: false,
  });
}
