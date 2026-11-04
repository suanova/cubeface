import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'

import { selfQueryOptions } from '@/lib/auth-query'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async ({ context, location }) => {
    const user = await context.queryClient
      .ensureQueryData(selfQueryOptions)
      .catch(() => null)
    if (!user) {
      throw redirect({
        to: '/sign-in',
        search: { redirect: location.href },
        replace: true,
      })
    }
  },
  component: Outlet,
})
