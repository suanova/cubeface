import { queryOptions } from '@tanstack/react-query'

import { getSelf } from '@/features/auth/api'

// Shared auth/self query. Routes can ensure this before rendering protected
// or guest-only pages, and the root provider keeps the store in sync.
export const selfQueryOptions = queryOptions({
  queryKey: ['auth', 'self'] as const,
  queryFn: getSelf,
  staleTime: 30_000,
  retry: false,
})
