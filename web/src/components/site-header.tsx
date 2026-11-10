import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { BrandLogo } from '@/components/brand-logo'
import { LanguageToggle } from '@/components/language-toggle'
import { ThemeToggle } from '@/components/theme-toggle'
import { useAuthStore } from '@/stores/auth-store'

const NAV = [
  { to: '/models' as const, label: 'Models' },
  { to: '/datasets' as const, label: 'Datasets' },
  { to: '/studio' as const, label: 'Studio' },
]

function userInitials(name: string): string {
  const parts = name.trim().split(/[\s._-]+/).filter(Boolean)
  const letters = parts.length > 1 ? [parts[0][0], parts[1][0]] : [name.trim()[0]]
  return letters.join('').toUpperCase() || 'U'
}

export function SiteHeader() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const initialized = useAuthStore((state) => state.initialized)

  return (
    <header className='site-topbar'>
      <div className='site-container site-topbar-inner'>
        <Link to='/' className='site-brand'>
          <span className='site-brand-mark' aria-hidden='true'>
            <BrandLogo showWordmark={false} />
          </span>
          <span className='site-brand-name'>CubeFace</span>
          <span className='site-brand-sub'>{t('brandSub')}</span>
        </Link>

        <nav className='site-mainnav' aria-label={t('Main navigation')}>
          {NAV.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              activeProps={{ className: 'is-active' }}
            >
              {t(item.label)}
            </Link>
          ))}
        </nav>

        <div className='site-topbar-actions'>
          <ThemeToggle />
          <LanguageToggle />
          {!initialized ? null : user ? (
            <Link
              to='/dashboard'
              className='site-avatar'
              aria-label={t('Dashboard')}
              title={t('Dashboard')}
            >
              {userInitials(user.display_name || user.username)}
            </Link>
          ) : (
            <>
              <Link to='/sign-in' className='site-topbar-signin'>
                {t('Sign in')}
              </Link>
              <Link to='/sign-up' className='site-btn site-btn-primary'>
                {t('Sign up')}
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  )
}
