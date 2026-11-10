import { describe, expect, it } from 'vitest'

import { formatDate, formatNumber, formatRelativeTime, initials } from './format'

describe('formatRelativeTime', () => {
  const now = Date.parse('2026-11-10T12:00:00Z')

  it('formats recent times in both locales', () => {
    expect(formatRelativeTime('2026-11-07T12:00:00Z', 'en', now)).toBe('3 days ago')
    expect(formatRelativeTime('2026-11-07T12:00:00Z', 'zh', now)).toBe('3天前')
  })

  it('handles invalid and empty values', () => {
    expect(formatRelativeTime('', 'en', now)).toBe('—')
    expect(formatRelativeTime('not-a-date', 'en', now)).toBe('—')
  })

  it('falls back to an absolute date beyond a year', () => {
    expect(formatRelativeTime('2024-01-01T00:00:00Z', 'en', now)).toMatch(/2024/)
  })
})

describe('formatDate', () => {
  it('formats dates by locale', () => {
    expect(formatDate('2026-11-07T12:00:00Z', 'en')).toMatch(/Nov/)
    expect(formatDate('2026-11-07T12:00:00Z', 'zh')).toMatch(/2026/)
    expect(formatDate(undefined, 'en')).toBe('—')
  })
})

describe('formatNumber', () => {
  it('uses locale grouping', () => {
    expect(formatNumber(12345, 'en')).toBe('12,345')
  })
})

describe('initials', () => {
  it('prefers the publisher and falls back to the model', () => {
    expect(initials('Alibaba Cloud', 'qwen')).toBe('AC')
    expect(initials('', 'qwen2.5-7b')).toBe('QW')
    expect(initials('', 'Q')).toBe('Q')
    expect(initials('', '')).toBe('?')
  })
})
