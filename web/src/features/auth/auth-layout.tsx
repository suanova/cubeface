import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { BrandLogo } from '@/components/brand-logo'
import { LanguageToggle } from '@/components/language-toggle'
import { ThemeToggle } from '@/components/theme-toggle'

export function AuthLayout({ children }: { children: React.ReactNode }) {
  const { t } = useTranslation()

  return (
    <div className='relative grid min-h-svh place-items-center overflow-hidden px-4 py-14 sm:py-20'>
      <div className='absolute right-4 top-4 z-10 flex items-center gap-1 sm:right-6 sm:top-6'>
        <ThemeToggle />
        <LanguageToggle />
      </div>

      <div className='flex w-full max-w-md flex-col items-center gap-6'>
        <Link to='/' className='site-brand transition-opacity hover:opacity-80'>
          <span className='site-brand-mark' aria-hidden='true'>
            <BrandLogo showWordmark={false} />
          </span>
          <span className='site-brand-name'>CubeFace</span>
          <span className='site-brand-sub'>{t('brandSub')}</span>
        </Link>

        <div className='w-full rounded-2xl border bg-card p-6 text-card-foreground shadow-sm sm:p-8'>
          {children}
        </div>
      </div>
    </div>
  )
}
