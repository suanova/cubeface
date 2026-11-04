import { QueryClient } from '@tanstack/react-query'
import { createRouter, RouterProvider } from '@tanstack/react-router'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import '@/i18n/config'
import '@/styles/index.css'

import { basePath } from '@/lib/base-path'

import { routeTree } from './routeTree.gen'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

const router = createRouter({
  routeTree,
  context: { queryClient },
  defaultPreload: 'intent',
  scrollRestoration: true,
  ...(basePath ? { basepath: basePath } : {}),
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

const container = document.getElementById('root')
if (!container) throw new Error('Root element not found')

createRoot(container).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>
)
