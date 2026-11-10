import { DEFAULT_DATASET_SORT, type DatasetSort } from './search'
import type { DatasetSummary } from './types'

export interface DatasetFilterState {
  q?: string
  task?: string
}

export interface DatasetFacetValue {
  value: string
  count: number
}

export function matchesDatasetKeywords(
  dataset: DatasetSummary,
  query: string | undefined
): boolean {
  const q = query?.trim().toLowerCase()
  if (!q) return true
  const haystack = [
    dataset.display_name,
    dataset.dataset,
    dataset.publisher,
    dataset.description,
    dataset.task,
    dataset.format,
    dataset.license,
    dataset.access,
    dataset.usage_boundary,
    ...dataset.tags,
  ]
    .join(' ')
    .toLowerCase()
  return haystack.includes(q)
}

export function filterDatasets(
  datasets: DatasetSummary[],
  filters: DatasetFilterState
): DatasetSummary[] {
  return datasets.filter((dataset) => {
    if (!matchesDatasetKeywords(dataset, filters.q)) return false
    const task = filters.task?.trim()
    if (task && dataset.task.trim() !== task) return false
    return true
  })
}

function updatedTime(dataset: DatasetSummary): number {
  const value = Date.parse(dataset.updated_at)
  return Number.isNaN(value) ? 0 : value
}

export function sortDatasets(
  datasets: DatasetSummary[],
  sort: DatasetSort = DEFAULT_DATASET_SORT
): DatasetSummary[] {
  const copy = [...datasets]
  switch (sort) {
    case 'name':
      return copy.sort((a, b) =>
        a.display_name.localeCompare(b.display_name, undefined, {
          sensitivity: 'base',
          numeric: true,
        })
      )
    case 'updated':
    default:
      return copy.sort((a, b) => updatedTime(b) - updatedTime(a))
  }
}

export function buildTaskFacets(datasets: DatasetSummary[]): DatasetFacetValue[] {
  const counts = new Map<string, number>()
  for (const dataset of datasets) {
    const task = dataset.task.trim()
    if (!task) continue
    counts.set(task, (counts.get(task) ?? 0) + 1)
  }
  return [...counts.entries()]
    .map(([value, count]) => ({ value, count }))
    .sort((a, b) => b.count - a.count || a.value.localeCompare(b.value))
}

export function activeDatasetFilterCount(filters: DatasetFilterState): number {
  let count = filters.q?.trim() ? 1 : 0
  if (filters.task?.trim()) count += 1
  return count
}
