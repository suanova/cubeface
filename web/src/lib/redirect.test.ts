import { describe, expect, it } from 'vitest'

import { safeRedirect } from './redirect'

describe('safeRedirect', () => {
  it.each([null, undefined, '', '   '])('falls back for empty input %j', (value) => {
    expect(safeRedirect(value)).toBe('/dashboard')
  })

  it('keeps same-origin paths', () => {
    expect(safeRedirect('/dashboard')).toBe('/dashboard')
    expect(safeRedirect('/sign-in?redirect=%2Fdashboard#top')).toBe(
      '/sign-in?redirect=%2Fdashboard#top'
    )
    expect(safeRedirect('  /models/author/model  ')).toBe('/models/author/model')
  })

  it.each([
    'https://evil.example',
    'http://evil.example/dashboard',
    '//evil.example',
    '///evil.example',
    'javascript:alert(1)',
    'dashboard',
    '/\\evil.example',
    '/\tevil',
    '/\nevil',
    '/\r\nevil',
  ])('rejects %j', (value) => {
    expect(safeRedirect(value)).toBe('/dashboard')
  })
})
