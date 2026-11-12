import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { BrandLogo } from './brand-logo'

describe('BrandLogo', () => {
  it('renders the mark and wordmark', () => {
    render(<BrandLogo />)
    expect(screen.getByRole('img', { name: 'MedRouter' })).toBeTruthy()
    expect(screen.getByText('MedRouter')).toBeTruthy()
  })

  it('can render the mark alone', () => {
    render(<BrandLogo showWordmark={false} className='text-xl' />)
    expect(screen.getByRole('img', { name: 'MedRouter' })).toBeTruthy()
    expect(screen.queryByText('MedRouter')).toBeNull()
  })
})
