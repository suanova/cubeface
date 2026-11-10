import { describe, expect, it } from 'vitest'

import { DEFAULT_MODEL_SORT, modelSearchSchema, resolveSort } from './search'

describe('modelSearchSchema', () => {
  it('accepts an empty search', () => {
    expect(modelSearchSchema.parse({})).toEqual({})
  })

  it('accepts known values and trims strings', () => {
    expect(
      modelSearchSchema.parse({ q: ' qwen ', task: 'chat', sort: 'updated' })
    ).toEqual({ q: 'qwen', task: 'chat', sort: 'updated' })
  })

  it('drops invalid and unknown values instead of failing', () => {
    expect(modelSearchSchema.parse({ q: 12, sort: 'downloads' })).toEqual({
      q: undefined,
      sort: undefined,
    })
    expect(modelSearchSchema.parse({ q: 'x'.repeat(121) })).toEqual({ q: undefined })
  })
})

describe('resolveSort', () => {
  it('falls back to featured for missing or unknown sorts', () => {
    expect(resolveSort(undefined)).toBe(DEFAULT_MODEL_SORT)
    expect(resolveSort('downloads')).toBe(DEFAULT_MODEL_SORT)
    expect(resolveSort('compatibility')).toBe('compatibility')
  })
})
