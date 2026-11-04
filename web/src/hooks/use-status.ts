import { useQuery } from '@tanstack/react-query'

import { getStatus } from '@/features/auth/api'

export function useStatus() {
  return useQuery({
    queryKey: ['status'],
    queryFn: getStatus,
    staleTime: 5 * 60_000,
    retry: 1,
  })
}
