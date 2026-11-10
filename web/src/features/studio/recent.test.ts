import { describe, expect, it } from 'vitest'

import { pickRecent } from './recent'

const items = [
  { name: 'a', updated_at: '2026-10-01T08:00:00Z' },
  { name: 'b', updated_at: '2026-11-01T08:00:00Z' },
  { name: 'c', updated_at: 'not-a-date' },
  { name: 'd', updated_at: '2026-09-01T08:00:00Z' },
]

const byUpdated = (item: (typeof items)[number]) => item.updated_at

describe('pickRecent', () => {
  it('sorts newest first and limits the result', () => {
    expect(pickRecent(items, byUpdated).map((item) => item.name)).toEqual([
      'b',
      'a',
      'd',
      'c',
    ])
    expect(pickRecent(items, byUpdated, 2).map((item) => item.name)).toEqual([
      'b',
      'a',
    ])
  })

  it('treats unparsable timestamps as oldest', () => {
    const picked = pickRecent(items, byUpdated)

    expect(picked[picked.length - 1].name).toBe('c')
  })

  it('does not mutate the source list', () => {
    const source = [...items]

    pickRecent(source, byUpdated)

    expect(source).toEqual(items)
  })
})
