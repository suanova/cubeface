import { cn } from '@/lib/utils'

interface BrandLogoProps {
  className?: string
  showWordmark?: boolean
}

// Original CubeFace mark: an isometric cube made of three faces.
export function BrandLogo({ className, showWordmark = true }: BrandLogoProps) {
  return (
    <span className={cn('inline-flex items-center gap-2 font-semibold tracking-tight', className)}>
      <svg
        viewBox='0 0 32 32'
        className='h-[1.5em] w-[1.5em] shrink-0'
        role='img'
        aria-label='CubeFace'
      >
        <polygon points='16,2 29,9.5 16,17 3,9.5' fill='#FB6415' />
        <polygon points='3,9.5 16,17 16,30 3,22.5' fill='#C2470A' />
        <polygon points='29,9.5 16,17 16,30 29,22.5' fill='#4C66FF' />
        <polygon points='16,2 29,9.5 16,17 3,9.5' fill='none' stroke='currentColor' strokeOpacity='0.08' />
      </svg>
      {showWordmark && <span className='leading-none'>CubeFace</span>}
    </span>
  )
}
