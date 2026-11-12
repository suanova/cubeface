import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { useEffect, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { PasswordInput } from '@/components/password-input'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { loginRequest } from '@/features/auth/api'
import { OAuthButtons } from '@/features/auth/oauth-buttons'
import { loginSchema, type LoginValues } from '@/features/auth/schemas'
import type { AuthUser } from '@/features/auth/types'
import { useStatus } from '@/hooks/use-status'
import { apiErrorMessage } from '@/lib/api'
import { withBasePath } from '@/lib/base-path'
import { safeRedirect } from '@/lib/redirect'
import { useAuthStore } from '@/stores/auth-store'

interface SignInFormProps {
  redirectTo?: string
  oauthError?: string
}

export function SignInForm({ redirectTo, oauthError }: SignInFormProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const setUser = useAuthStore((state) => state.setUser)
  const setInitialized = useAuthStore((state) => state.setInitialized)
  const { data: status } = useStatus()
  const [showResendHint, setShowResendHint] = useState(false)
  const oauthErrorShown = useRef(false)

  const form = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { username: '', password: '' },
  })

  useEffect(() => {
    if (!oauthError || oauthErrorShown.current) return
    oauthErrorShown.current = true
    toast.error(t('Sign-in could not be completed. Please try again.'))
  }, [oauthError, t])

  const onSubmit = form.handleSubmit(async (values) => {
    try {
      const response = await loginRequest({
        username: values.username.trim(),
        password: values.password,
      })
      if (!response.success || !response.data) {
        toast.error(t(response.message || 'Sign-in failed'))
        return
      }
      const user = response.data.user
      queryClient.setQueryData<AuthUser>(['auth', 'self'], user)
      setUser(user)
      setInitialized(true)
      toast.success(t('Welcome back!'))
      window.location.assign(withBasePath(safeRedirect(redirectTo)))
    } catch (error) {
      const message = apiErrorMessage(error)
      setShowResendHint(message === 'Please verify your email address before signing in.')
      toast.error(message ? t(message) : t('Sign-in failed'))
    }
  })

  return (
    <div className='space-y-6'>
      <div className='space-y-1.5 text-center'>
        <h1 className='text-2xl font-semibold tracking-tight'>{t('Welcome back')}</h1>
        <p className='text-sm text-muted-foreground'>
          {t('Sign in to continue to MedRouter')}
        </p>
      </div>

      <form className='space-y-4' onSubmit={onSubmit} noValidate>
        <div className='space-y-2'>
          <Label htmlFor='username'>{t('Username or email')}</Label>
          <Input
            id='username'
            autoComplete='username'
            autoFocus
            placeholder={t('Enter your username or email')}
            aria-invalid={Boolean(form.formState.errors.username)}
            {...form.register('username')}
          />
          {form.formState.errors.username ? (
            <p className='text-sm text-destructive'>
              {t(form.formState.errors.username.message ?? '')}
            </p>
          ) : null}
        </div>

        <div className='space-y-2'>
          <div className='flex items-center justify-between'>
            <Label htmlFor='password'>{t('Password')}</Label>
          </div>
          <PasswordInput
            id='password'
            autoComplete='current-password'
            placeholder={t('Enter your password')}
            aria-invalid={Boolean(form.formState.errors.password)}
            {...form.register('password')}
          />
          {form.formState.errors.password ? (
            <p className='text-sm text-destructive'>
              {t(form.formState.errors.password.message ?? '')}
            </p>
          ) : null}
        </div>

        {showResendHint ? (
          <p className='rounded-md border border-border bg-muted/50 px-3 py-2 text-sm text-muted-foreground'>
            {t('Need a new verification code? Create your account again with the same email.')}
          </p>
        ) : null}

        <Button className='h-11 w-full rounded-lg' type='submit' disabled={form.formState.isSubmitting}>
          {form.formState.isSubmitting ? t('Signing in...') : t('Sign in')}
        </Button>
      </form>

      {status?.oauth_providers?.length ? (
        <OAuthButtons providers={status.oauth_providers} redirectTo={redirectTo} />
      ) : null}

      <p className='text-center text-sm text-muted-foreground'>
        {t("Don't have an account?")}{' '}
        <Link
          to='/sign-up'
          search={{ redirect: redirectTo }}
          className='font-medium text-foreground underline-offset-4 hover:underline'
        >
          {t('Sign up')}
        </Link>
      </p>
    </div>
  )
}
