import { describe, expect, it } from 'vitest'

import { DEFAULT_DATASET_SORT, datasetSearchSchema, resolveDatasetSort } from './search'

describe('datasetSearchSchema', () => {
  it('accepts an empty search', () => {
    expect(datasetSearchSchema.parse({})).toEqual({})
  })

  it('accepts known values and trims strings', () => {
    expect(
      datasetSearchSchema.parse({ q: ' parquet ', task: 'sft', sort: 'name' })
    ).toEqual({ q: 'parquet', task: 'sft', sort: 'name' })
  })

  it('drops invalid and unknown values instead of failing', () => {
    expect(datasetSearchSchema.parse({ q: 12, sort: 'downloads' })).toEqual({
      q: undefined,
      sort: undefined,
    })
    expect(datasetSearchSchema.parse({ q: 'x'.repeat(121) })).toEqual({ q: undefined })
  })
})

describe('resolveDatasetSort', () => {
  it('falls back to updated for missing or unknown sorts', () => {
    expect(resolveDatasetSort(undefined)).toBe(DEFAULT_DATASET_SORT)
    expect(resolveDatasetSort('downloads')).toBe(DEFAULT_DATASET_SORT)
    expect(resolveDatasetSort('name')).toBe('name')
  })
})
