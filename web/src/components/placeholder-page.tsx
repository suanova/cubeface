import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { SiteHeader } from '@/components/site-header'
import { buttonVariants } from '@/components/ui/button'

interface PlaceholderPageProps {
  title: string
  description?: string
}

export function PlaceholderPage({ title, description }: PlaceholderPageProps) {
  const { t } = useTranslation()
  return (
    <div className='min-h-svh bg-background'>
      <SiteHeader />
      <main className='mx-auto flex min-h-svh max-w-3xl flex-col items-start justify-center px-6 py-24'>
        <span className='mb-5 inline-flex items-center rounded-full border border-border bg-muted px-3 py-1 font-mono text-xs uppercase tracking-widest text-muted-foreground'>
          {t('Coming soon')}
        </span>
        <h1 className='text-4xl font-semibold tracking-tight sm:text-5xl'>{t(title)}</h1>
        <p className='mt-4 max-w-xl text-lg text-muted-foreground'>
          {description ? t(description) : t('This page is not built yet.')}
        </p>
        <Link
          to='/'
          className={buttonVariants({ variant: 'outline' }) + ' mt-8 rounded-full'}
        >
          {t('Back to home')}
        </Link>
      </main>
    </div>
  )
}
