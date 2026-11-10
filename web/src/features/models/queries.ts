import { queryOptions } from '@tanstack/react-query'

import { fetchModelDetail, fetchModelList, ModelNotFoundError } from './api'

export const modelListQueryOptions = queryOptions({
  queryKey: ['models', 'list'],
  queryFn: fetchModelList,
  staleTime: 30_000,
  retry: 1,
})

export function modelDetailQueryOptions(name: string) {
  return queryOptions({
    queryKey: ['models', 'detail', name],
    queryFn: () => fetchModelDetail(name),
    staleTime: 30_000,
    retry: (failureCount, error) =>
      !(error instanceof ModelNotFoundError) && failureCount < 1,
  })
}
