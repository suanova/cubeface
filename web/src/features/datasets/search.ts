import { z } from 'zod'

export const DATASET_SORTS = ['updated', 'name'] as const

export type DatasetSort = (typeof DATASET_SORTS)[number]

export const DEFAULT_DATASET_SORT: DatasetSort = 'updated'

const filterValue = z.string().trim().max(120).optional().catch(undefined)

// Search parameters are shareable. Unknown or malformed values fall back to
// "not set" instead of making the page fail to render.
export const datasetSearchSchema = z.object({
  q: filterValue,
  task: filterValue,
  sort: z.enum(DATASET_SORTS).optional().catch(undefined),
})

export type DatasetSearch = z.infer<typeof datasetSearchSchema>

export function resolveDatasetSort(value: string | undefined): DatasetSort {
  return DATASET_SORTS.includes(value as DatasetSort)
    ? (value as DatasetSort)
    : DEFAULT_DATASET_SORT
}
