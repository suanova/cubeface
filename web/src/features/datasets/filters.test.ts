import { describe, expect, it } from 'vitest'

import type { DatasetSummary } from './types'
import {
  activeDatasetFilterCount,
  buildTaskFacets,
  filterDatasets,
  matchesDatasetKeywords,
  sortDatasets,
} from './filters'

export function dataset(overrides: Partial<DatasetSummary> = {}): DatasetSummary {
  return {
    dataset: 'demo-dataset',
    display_name: 'Demo Dataset',
    publisher: 'Cube Labs',
    publisher_initials: 'CL',
    description: 'Instruction tuning corpus',
    format: 'Parquet',
    task: 'instruction-tuning',
    license: 'CC-BY-4.0',
    tags: ['sft', 'parquet'],
    access: 'public',
    deidentification: 'PII removed',
    ethics: '',
    source_repo: 'https://example.com/demo',
    usage_boundary: 'Training only',
    size: '1.2Gi',
    record_count: 120000,
    storage_strategy: 'S3',
    latest_version: 'v1.1.0',
    version_count: 2,
    featured: false,
    created_at: '2026-09-01T08:00:00Z',
    updated_at: '2026-10-01T08:00:00Z',
    ...overrides,
  }
}

const datasets = [
  dataset(),
  dataset({
    dataset: 'clinical-notes',
    display_name: 'Clinical Notes',
    publisher: 'Med Lab',
    publisher_initials: 'ML',
    description: 'De-identified clinical notes',
    format: 'JSONL',
    task: 'clinical',
    license: 'Restricted',
    tags: ['clinical'],
    access: 'gated',
    size: '480Mi',
    record_count: 12000,
    latest_version: 'v2.0.0',
    version_count: 1,
    updated_at: '2026-08-01T08:00:00Z',
  }),
  dataset({
    dataset: 'retrieval-pairs',
    display_name: 'Retrieval Pairs',
    publisher: 'Cube Labs',
    publisher_initials: 'CL',
    description: 'Question/answer pairs',
    format: 'Parquet',
    task: 'retrieval',
    tags: [],
    size: '2.4Gi',
    record_count: 900000,
    updated_at: '2026-11-01T08:00:00Z',
  }),
]

describe('matchesDatasetKeywords', () => {
  it('matches name, identity, publisher and tags case-insensitively', () => {
    expect(matchesDatasetKeywords(datasets[0], 'CLINICAL')).toBe(false)
    expect(matchesDatasetKeywords(datasets[0], 'cube labs')).toBe(true)
    expect(matchesDatasetKeywords(datasets[0], 'sft')).toBe(true)
    expect(matchesDatasetKeywords(datasets[0], undefined)).toBe(true)
  })
})

describe('filterDatasets', () => {
  it('combines free-text search with the task filter', () => {
    expect(filterDatasets(datasets, { q: 'parquet' })).toHaveLength(2)
    expect(filterDatasets(datasets, { q: 'parquet', task: 'retrieval' })).toHaveLength(1)
    expect(filterDatasets(datasets, { task: 'clinical' })[0].dataset).toBe(
      'clinical-notes'
    )
    expect(filterDatasets(datasets, { task: 'missing' })).toHaveLength(0)
  })
})

describe('sortDatasets', () => {
  it('sorts by updated time by default and by name on request', () => {
    expect(sortDatasets(datasets).map((item) => item.dataset)).toEqual([
      'retrieval-pairs',
      'demo-dataset',
      'clinical-notes',
    ])
    expect(sortDatasets(datasets, 'name').map((item) => item.dataset)).toEqual([
      'clinical-notes',
      'demo-dataset',
      'retrieval-pairs',
    ])
  })
})

describe('buildTaskFacets', () => {
  it('counts tasks and sorts by count then name', () => {
    expect(buildTaskFacets(datasets)).toEqual([
      { value: 'clinical', count: 1 },
      { value: 'instruction-tuning', count: 1 },
      { value: 'retrieval', count: 1 },
    ])
  })
})

describe('activeDatasetFilterCount', () => {
  it('counts search and task filters', () => {
    expect(activeDatasetFilterCount({})).toBe(0)
    expect(activeDatasetFilterCount({ q: '  ' })).toBe(0)
    expect(activeDatasetFilterCount({ q: 'x', task: 'clinical' })).toBe(2)
  })
})
