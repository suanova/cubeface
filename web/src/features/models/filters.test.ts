import { describe, expect, it } from 'vitest'

import type { ModelSummary } from './types'
import {
  activeFilterCount,
  buildFacets,
  filterModels,
  matchesKeywords,
  sortModels,
} from './filters'

export function model(overrides: Partial<ModelSummary> = {}): ModelSummary {
  return {
    model: 'qwen2.5-7b-instruct',
    display_name: 'Qwen2.5-7B-Instruct',
    publisher: 'Alibaba Cloud',
    publisher_initials: 'AC',
    description: 'Chat model',
    architecture: 'qwen2',
    quantization: 'bf16',
    task: 'chat',
    license: 'Apache-2.0',
    tags: ['chat', 'qwen'],
    usage_boundary: '',
    storage_strategy: 'S3',
    latest_version: 'v1.2',
    version_count: 2,
    compatible_profile_count: 1,
    used_by_count: 0,
    featured: false,
    created_at: '2026-09-01T08:00:00Z',
    updated_at: '2026-10-01T08:00:00Z',
    ...overrides,
  }
}

const models = [
  model(),
  model({
    model: 'llama-3-8b',
    display_name: 'Llama 3 8B',
    publisher: 'Meta',
    publisher_initials: 'ME',
    description: 'General assistant',
    architecture: 'llama',
    quantization: 'w8a8',
    task: 'text',
    license: 'Llama-3',
    tags: ['llama'],
    compatible_profile_count: 0,
    featured: true,
    updated_at: '2026-11-01T08:00:00Z',
  }),
  model({
    model: 'medgemma-4b',
    display_name: 'MedGemma 4B',
    publisher: 'Google Health',
    publisher_initials: 'GH',
    architecture: 'gemma',
    quantization: 'bf16',
    task: 'chat',
    license: 'Apache-2.0',
    tags: ['medical'],
    compatible_profile_count: 2,
    updated_at: '2026-10-15T08:00:00Z',
  }),
]

describe('filterModels', () => {
  it('matches keywords across name, publisher and tags', () => {
    expect(matchesKeywords(models[0], 'alibaba')).toBe(true)
    expect(matchesKeywords(models[2], 'medical')).toBe(true)
    expect(matchesKeywords(models[1], 'qwen')).toBe(false)
    expect(filterModels(models, { q: 'google' })).toEqual([models[2]])
  })

  it('combines facets with AND semantics', () => {
    expect(filterModels(models, { task: 'chat' })).toHaveLength(2)
    expect(filterModels(models, { task: 'chat', quant: 'bf16' })).toHaveLength(2)
    expect(
      filterModels(models, { task: 'chat', license: 'Apache-2.0', arch: 'gemma' })
    ).toEqual([models[2]])
    expect(filterModels(models, { task: 'chat', arch: 'llama' })).toHaveLength(0)
  })
})

describe('sortModels', () => {
  it('sorts featured first by default', () => {
    expect(sortModels(models)[0].model).toBe('llama-3-8b')
  })

  it('sorts by updated time', () => {
    expect(sortModels(models, 'updated').map((m) => m.model)).toEqual([
      'llama-3-8b',
      'medgemma-4b',
      'qwen2.5-7b-instruct',
    ])
  })

  it('sorts by name and compatibility', () => {
    expect(sortModels(models, 'name').map((m) => m.model)).toEqual([
      'llama-3-8b',
      'medgemma-4b',
      'qwen2.5-7b-instruct',
    ])
    expect(sortModels(models, 'compatibility')[0].model).toBe('medgemma-4b')
  })

  it('does not mutate the input', () => {
    const original = models.map((m) => m.model)
    sortModels(models, 'name')
    expect(models.map((m) => m.model)).toEqual(original)
  })
})

describe('facets', () => {
  it('counts distinct values and drops empty ones', () => {
    const facets = buildFacets(models)
    expect(facets.task).toEqual([
      { value: 'chat', count: 2 },
      { value: 'text', count: 1 },
    ])
    expect(facets.usage).toEqual([])
    expect(facets.license[0]).toEqual({ value: 'Apache-2.0', count: 2 })
  })

  it('counts active filters including search', () => {
    expect(activeFilterCount({})).toBe(0)
    expect(activeFilterCount({ q: 'x', task: 'chat', arch: '' })).toBe(2)
  })
})
