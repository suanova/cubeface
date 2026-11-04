import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SiteHeader } from '@/components/site-header'
import { Button } from '@/components/ui/button'
import { logoutRequest } from '@/features/auth/api'
import { selfQueryOptions } from '@/lib/auth-query'
import { useAuthStore } from '@/stores/auth-store'

function formatDate(value: string | null | undefined): string {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return date.toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function DashboardPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const storedUser = useAuthStore((state) => state.user)
  const clear = useAuthStore((state) => state.clear)
  const { data: fetchedUser } = useQuery(selfQueryOptions)
  const user = storedUser ?? fetchedUser ?? null

  if (!user) return null

  const role =
    user.role >= 100 ? t('Administrator') : user.role >= 1 ? t('Member') : t('Guest')

  const onSignOut = async () => {
    try {
      await logoutRequest()
    } catch {
      // Logout is idempotent server-side; clear the client state regardless.
    }
    queryClient.setQueryData(['auth', 'self'], null)
    clear()
    toast.success(t('Signed out'))
    navigate({ to: '/', replace: true })
  }

  return (
    <div className='min-h-svh bg-background'>
      <SiteHeader />
      <main className='mx-auto w-full max-w-5xl px-6 pb-24 pt-28 sm:pt-32'>
        <div className='flex flex-wrap items-start justify-between gap-4'>
          <div>
            <p className='font-mono text-xs uppercase tracking-widest text-muted-foreground'>
              {t('Dashboard')}
            </p>
            <h1 className='mt-2 text-3xl font-semibold tracking-tight sm:text-4xl'>
              {t('Welcome, {{name}}', { name: user.display_name || user.username })}
            </h1>
          </div>
          <Button variant='outline' className='rounded-full' onClick={() => void onSignOut()}>
            {t('Sign out')}
          </Button>
        </div>

        <div className='mt-10 grid gap-6 md:grid-cols-[1.15fr_0.85fr]'>
          <section className='rounded-2xl border bg-card p-6 text-card-foreground shadow-sm'>
            <h2 className='text-lg font-semibold'>{t('Your account')}</h2>
            <dl className='mt-5 grid gap-4 sm:grid-cols-2'>
              <div>
                <dt className='text-xs uppercase tracking-wide text-muted-foreground'>
                  {t('Username')}
                </dt>
                <dd className='mt-1 truncate text-sm font-medium'>{user.username}</dd>
              </div>
              <div>
                <dt className='text-xs uppercase tracking-wide text-muted-foreground'>
                  {t('Email')}
                </dt>
                <dd className='mt-1 flex min-w-0 flex-wrap items-center gap-2 text-sm font-medium'>
                  <span className='break-all'>{user.email || '—'}</span>
                  {user.email ? (
                    <span className='shrink-0 rounded-full border border-border px-2 py-0.5 text-[11px] text-muted-foreground'>
                      {user.email_verified ? t('Email verified') : t('Not verified')}
                    </span>
                  ) : null}
                </dd>
              </div>
              <div>
                <dt className='text-xs uppercase tracking-wide text-muted-foreground'>
                  {t('Role')}
                </dt>
                <dd className='mt-1 text-sm font-medium'>{role}</dd>
              </div>
              <div>
                <dt className='text-xs uppercase tracking-wide text-muted-foreground'>
                  {t('Joined')}
                </dt>
                <dd className='mt-1 text-sm font-medium'>{formatDate(user.created_at)}</dd>
              </div>
            </dl>
          </section>

          <section className='flex flex-col justify-between rounded-2xl border bg-card p-6 text-card-foreground shadow-sm'>
            <div>
              <span className='inline-flex items-center rounded-full border border-border bg-muted px-3 py-1 font-mono text-[11px] uppercase tracking-widest text-muted-foreground'>
                {t('Coming soon')}
              </span>
              <h2 className='mt-4 text-lg font-semibold'>{t('Model hub')}</h2>
              <p className='mt-2 text-sm text-muted-foreground'>
                {t(
                  'Model hosting is under construction. Soon you will be able to publish, discover and run models here.'
                )}
              </p>
            </div>
          </section>
        </div>
      </main>
    </div>
  )
}
