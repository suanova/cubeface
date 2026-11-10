import { z } from 'zod'

export const MODEL_SORTS = ['featured', 'updated', 'name', 'compatibility'] as const

export type ModelSort = (typeof MODEL_SORTS)[number]

export const DEFAULT_MODEL_SORT: ModelSort = 'featured'

const filterValue = z.string().trim().max(120).optional().catch(undefined)

// Search parameters are shareable. Unknown or malformed values fall back to
// "not set" instead of making the page fail to render.
export const modelSearchSchema = z.object({
  q: filterValue,
  task: filterValue,
  arch: filterValue,
  quant: filterValue,
  license: filterValue,
  usage: filterValue,
  sort: z.enum(MODEL_SORTS).optional().catch(undefined),
})

export type ModelSearch = z.infer<typeof modelSearchSchema>

export function resolveSort(value: string | undefined): ModelSort {
  return MODEL_SORTS.includes(value as ModelSort)
    ? (value as ModelSort)
    : DEFAULT_MODEL_SORT
}
