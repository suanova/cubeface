import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query'
import {
  createRootRouteWithContext,
  Outlet,
  type ErrorComponentProps,
} from '@tanstack/react-router'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import { NotFoundPage } from '@/components/not-found-page'
import { Toaster } from '@/components/ui/sonner'
import { ThemeProvider } from '@/context/theme-provider'
import { selfQueryOptions } from '@/lib/auth-query'
import { useAuthStore } from '@/stores/auth-store'

export interface RouterContext {
  queryClient: QueryClient
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootComponent,
  notFoundComponent: NotFoundPage,
  errorComponent: RootError,
})

function RootComponent() {
  const { queryClient } = Route.useRouteContext()
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <AuthBootstrap />
        <Outlet />
        <Toaster />
      </ThemeProvider>
    </QueryClientProvider>
  )
}

// AuthBootstrap hydrates the Zustand auth store from the shared self query.
// It never blocks rendering, so public pages appear immediately.
function AuthBootstrap() {
  const { data, isFetched } = useQuery(selfQueryOptions)
  const setUser = useAuthStore((state) => state.setUser)
  const setInitialized = useAuthStore((state) => state.setInitialized)

  useEffect(() => {
    if (!isFetched) return
    setUser(data ?? null)
    setInitialized(true)
  }, [data, isFetched, setInitialized, setUser])

  return null
}

function RootError({ error }: ErrorComponentProps) {
  const { t } = useTranslation()
  const detail = error instanceof Error ? error.message : ''
  return (
    <div className='grid min-h-svh place-items-center px-6'>
      <div className='max-w-md text-center'>
        <h1 className='text-2xl font-semibold'>{t('Something went wrong.')}</h1>
        <p className='mt-2 text-sm text-muted-foreground'>
          {detail || t('Something went wrong. Please try again.')}
        </p>
        <button
          type='button'
          onClick={() => window.location.reload()}
          className='mt-6 rounded-full border border-border px-4 py-2 text-sm hover:bg-accent'
        >
          {t('Reload')}
        </button>
      </div>
    </div>
  )
}
