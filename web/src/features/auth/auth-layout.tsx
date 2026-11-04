import { Link } from '@tanstack/react-router'

import { BrandLogo } from '@/components/brand-logo'
import { LanguageToggle } from '@/components/language-toggle'
import { ThemeToggle } from '@/components/theme-toggle'

export function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className='relative grid min-h-svh place-items-center overflow-hidden px-4 py-14 sm:py-20'>
      <div className='absolute right-4 top-4 z-10 flex items-center gap-1 sm:right-6 sm:top-6'>
        <ThemeToggle />
        <LanguageToggle />
      </div>

      <div className='flex w-full max-w-md flex-col items-center gap-6'>
        <Link to='/' className='text-xl transition-opacity hover:opacity-80'>
          <BrandLogo />
        </Link>

        <div className='w-full rounded-2xl border bg-card p-6 text-card-foreground shadow-sm sm:p-8'>
          {children}
        </div>
      </div>
    </div>
  )
}
