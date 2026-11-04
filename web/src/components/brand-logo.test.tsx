import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { BrandLogo } from './brand-logo'

describe('BrandLogo', () => {
  it('renders the mark and wordmark', () => {
    render(<BrandLogo />)
    expect(screen.getByRole('img', { name: 'CubeFace' })).toBeTruthy()
    expect(screen.getByText('CubeFace')).toBeTruthy()
  })

  it('can render the mark alone', () => {
    render(<BrandLogo showWordmark={false} className='text-xl' />)
    expect(screen.getByRole('img', { name: 'CubeFace' })).toBeTruthy()
    expect(screen.queryByText('CubeFace')).toBeNull()
  })
})
