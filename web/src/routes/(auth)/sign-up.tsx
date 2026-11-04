import { createFileRoute, redirect } from '@tanstack/react-router'
import { z } from 'zod'

import { SignUpForm } from '@/features/auth/sign-up/sign-up-form'
import { selfQueryOptions } from '@/lib/auth-query'

const searchSchema = z.object({
  redirect: z.string().optional(),
})

export const Route = createFileRoute('/(auth)/sign-up')({
  validateSearch: searchSchema,
  beforeLoad: async ({ context }) => {
    const user = await context.queryClient
      .ensureQueryData(selfQueryOptions)
      .catch(() => null)
    if (user) throw redirect({ to: '/dashboard', replace: true })
  },
  component: SignUpRoute,
})

function SignUpRoute() {
  const { redirect: redirectTo } = Route.useSearch()
  return <SignUpForm redirectTo={redirectTo} />
}
