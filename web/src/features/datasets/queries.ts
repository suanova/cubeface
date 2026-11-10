import { queryOptions } from '@tanstack/react-query'

import { DatasetNotFoundError, fetchDatasetDetail, fetchDatasetList } from './api'

export const datasetListQueryOptions = queryOptions({
  queryKey: ['datasets', 'list'],
  queryFn: fetchDatasetList,
  staleTime: 30_000,
  retry: 1,
})

export function datasetDetailQueryOptions(name: string) {
  return queryOptions({
    queryKey: ['datasets', 'detail', name],
    queryFn: () => fetchDatasetDetail(name),
    staleTime: 30_000,
    retry: (failureCount, error) =>
      !(error instanceof DatasetNotFoundError) && failureCount < 1,
  })
}
