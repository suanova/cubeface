import type { ModelSummary } from './types'
import { DEFAULT_MODEL_SORT, type ModelSort } from './search'

export interface ModelFilterState {
  q?: string
  task?: string
  arch?: string
  quant?: string
  license?: string
  usage?: string
}

export const FACET_KEYS = ['task', 'arch', 'quant', 'license', 'usage'] as const

export type FacetKey = (typeof FACET_KEYS)[number]

export interface FacetValue {
  value: string
  count: number
}

export type Facets = Record<FacetKey, FacetValue[]>

const FACET_FIELDS: Record<FacetKey, keyof ModelSummary> = {
  task: 'task',
  arch: 'architecture',
  quant: 'quantization',
  license: 'license',
  usage: 'usage_boundary',
}

function facetField(model: ModelSummary, key: FacetKey): string {
  const raw = model[FACET_FIELDS[key]]
  return typeof raw === 'string' ? raw.trim() : ''
}

export function matchesKeywords(model: ModelSummary, query: string | undefined): boolean {
  const q = query?.trim().toLowerCase()
  if (!q) return true
  const haystack = [
    model.display_name,
    model.model,
    model.publisher,
    model.description,
    model.task,
    model.architecture,
    model.quantization,
    model.license,
    model.usage_boundary,
    ...model.tags,
  ]
    .join(' ')
    .toLowerCase()
  return haystack.includes(q)
}

export function filterModels(
  models: ModelSummary[],
  filters: ModelFilterState
): ModelSummary[] {
  return models.filter((model) => {
    if (!matchesKeywords(model, filters.q)) return false
    for (const key of FACET_KEYS) {
      const wanted = filters[key]?.trim()
      if (wanted && facetField(model, key) !== wanted) return false
    }
    return true
  })
}

function updatedTime(model: ModelSummary): number {
  const value = Date.parse(model.updated_at)
  return Number.isNaN(value) ? 0 : value
}

export function sortModels(
  models: ModelSummary[],
  sort: ModelSort = DEFAULT_MODEL_SORT
): ModelSummary[] {
  const copy = [...models]
  switch (sort) {
    case 'name':
      return copy.sort((a, b) =>
        a.display_name.localeCompare(b.display_name, undefined, {
          sensitivity: 'base',
          numeric: true,
        })
      )
    case 'updated':
      return copy.sort((a, b) => updatedTime(b) - updatedTime(a))
    case 'compatibility':
      return copy.sort(
        (a, b) =>
          b.compatible_profile_count - a.compatible_profile_count ||
          updatedTime(b) - updatedTime(a)
      )
    case 'featured':
    default:
      return copy.sort(
        (a, b) =>
          Number(b.featured) - Number(a.featured) || updatedTime(b) - updatedTime(a)
      )
  }
}

export function buildFacets(models: ModelSummary[]): Facets {
  const facets: Facets = { task: [], arch: [], quant: [], license: [], usage: [] }
  for (const key of FACET_KEYS) {
    const counts = new Map<string, number>()
    for (const model of models) {
      const value = facetField(model, key)
      if (!value) continue
      counts.set(value, (counts.get(value) ?? 0) + 1)
    }
    facets[key] = [...counts.entries()]
      .map(([value, count]) => ({ value, count }))
      .sort(
        (a, b) => b.count - a.count || a.value.localeCompare(b.value)
      )
  }
  return facets
}

export function activeFilterCount(filters: ModelFilterState): number {
  let count = filters.q?.trim() ? 1 : 0
  for (const key of FACET_KEYS) {
    if (filters[key]?.trim()) count += 1
  }
  return count
}

export function hasFacet(facets: Facets): boolean {
  return FACET_KEYS.some((key) => facets[key].length > 0)
}
