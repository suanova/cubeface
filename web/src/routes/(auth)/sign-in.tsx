import { createFileRoute, redirect } from '@tanstack/react-router'
import { z } from 'zod'

import { SignInForm } from '@/features/auth/sign-in/sign-in-form'
import { selfQueryOptions } from '@/lib/auth-query'

const searchSchema = z.object({
  redirect: z.string().optional(),
  oauth_error: z.string().optional(),
})

export const Route = createFileRoute('/(auth)/sign-in')({
  validateSearch: searchSchema,
  beforeLoad: async ({ context }) => {
    const user = await context.queryClient
      .ensureQueryData(selfQueryOptions)
      .catch(() => null)
    if (user) throw redirect({ to: '/dashboard', replace: true })
  },
  component: SignInRoute,
})

function SignInRoute() {
  const { redirect: redirectTo, oauth_error: oauthError } = Route.useSearch()
  return <SignInForm redirectTo={redirectTo} oauthError={oauthError} />
}
