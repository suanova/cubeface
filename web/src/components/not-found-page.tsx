import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { SiteHeader } from '@/components/site-header'
import { buttonVariants } from '@/components/ui/button'

export function NotFoundPage() {
  const { t } = useTranslation()
  return (
    <div className='min-h-svh bg-background'>
      <SiteHeader />
      <main className='mx-auto flex min-h-svh max-w-2xl flex-col items-start justify-center px-6 py-24'>
        <span className='font-mono text-sm text-muted-foreground'>404</span>
        <h1 className='mt-3 text-4xl font-semibold tracking-tight sm:text-5xl'>
          {t('Page not found')}
        </h1>
        <p className='mt-4 text-lg text-muted-foreground'>
          {t("The page you are looking for doesn't exist.")}
        </p>
        <Link to='/' className={buttonVariants({ variant: 'outline' }) + ' mt-8 rounded-full'}>
          {t('Back to home')}
        </Link>
      </main>
    </div>
  )
}
