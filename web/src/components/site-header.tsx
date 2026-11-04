import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { BrandLogo } from '@/components/brand-logo'
import { LanguageToggle } from '@/components/language-toggle'
import { ThemeToggle } from '@/components/theme-toggle'
import { buttonVariants } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth-store'

const NAV = [
  { to: '/models' as const, label: 'Models' },
  { to: '/pricing' as const, label: 'Pricing' },
  { to: '/docs' as const, label: 'Docs' },
]

export function SiteHeader() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const initialized = useAuthStore((state) => state.initialized)

  return (
    <header className='fixed inset-x-0 top-0 z-50'>
      <div className='mx-auto mt-3 flex h-14 max-w-[64rem] items-center justify-between gap-2 rounded-full border border-border/70 bg-background/75 px-3 shadow-sm backdrop-blur-md sm:px-4'>
        <div className='flex min-w-0 items-center gap-6'>
          <Link to='/' className='flex shrink-0 items-center text-[15px]'>
            <BrandLogo />
          </Link>
          <nav className='hidden items-center gap-5 sm:flex'>
            {NAV.map((item) => (
              <Link
                key={item.to}
                to={item.to}
                className='text-sm text-muted-foreground transition-colors hover:text-foreground'
                activeProps={{ className: 'text-foreground' }}
              >
                {t(item.label)}
              </Link>
            ))}
          </nav>
        </div>

        <div className='flex shrink-0 items-center gap-1'>
          <ThemeToggle />
          <LanguageToggle />
          {!initialized ? null : user ? (
            <Link
              to='/dashboard'
              className={buttonVariants({ variant: 'default', size: 'sm' }) + ' ml-1 rounded-full'}
            >
              {t('Dashboard')}
            </Link>
          ) : (
            <>
              <Link
                to='/sign-in'
                className={buttonVariants({ variant: 'ghost', size: 'sm' }) + ' ml-1 hidden rounded-full sm:inline-flex'}
              >
                {t('Sign in')}
              </Link>
              <Link
                to='/sign-up'
                className={buttonVariants({ variant: 'default', size: 'sm' }) + ' rounded-full'}
              >
                {t('Sign up')}
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  )
}
