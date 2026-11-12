import { zodResolver } from '@hookform/resolvers/zod'
import { Link, useNavigate } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { PasswordInput } from '@/components/password-input'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { registerRequest } from '@/features/auth/api'
import { OAuthButtons } from '@/features/auth/oauth-buttons'
import { registerSchema, type RegisterValues } from '@/features/auth/schemas'
import { useEmailVerification } from '@/features/auth/use-email-verification'
import { useStatus } from '@/hooks/use-status'
import { apiErrorMessage } from '@/lib/api'

interface SignUpFormProps {
  redirectTo?: string
}

export function SignUpForm({ redirectTo }: SignUpFormProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { data: status } = useStatus()
  const verificationEnabled = status?.email_verification_enabled ?? true
  const registerEnabled = status?.register_enabled ?? true
  const { isSending, isActive, secondsLeft, sendCode } = useEmailVerification()

  const form = useForm<RegisterValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      username: '',
      email: '',
      verificationCode: '',
      password: '',
      confirmPassword: '',
    },
  })

  const onSubmit = form.handleSubmit(async (values) => {
    const code = values.verificationCode?.trim() ?? ''
    if (verificationEnabled && !code) {
      form.setError('verificationCode', { message: 'Please enter the verification code' })
      return
    }

    try {
      const response = await registerRequest({
        username: values.username.trim(),
        email: values.email.trim(),
        password: values.password,
        verification_code: code,
      })
      if (!response.success) {
        toast.error(t(response.message || 'Registration failed'))
        return
      }
      toast.success(t('Account created! Please sign in'))
      navigate({
        to: '/sign-in',
        search: { redirect: redirectTo },
        replace: true,
      })
    } catch (error) {
      const message = apiErrorMessage(error)
      toast.error(message ? t(message) : t('Registration failed'))
    }
  })

  if (!registerEnabled) {
    return (
      <div className='space-y-6 text-center'>
        <div className='space-y-1.5'>
          <h1 className='text-2xl font-semibold tracking-tight'>{t('Registration closed')}</h1>
          <p className='text-sm text-muted-foreground'>
            {t('New account registration is currently disabled.')}
          </p>
        </div>
        <Link
          to='/sign-in'
          className='inline-flex text-sm font-medium underline-offset-4 hover:underline'
        >
          {t('Back to sign in')}
        </Link>
      </div>
    )
  }

  return (
    <div className='space-y-6'>
      <div className='space-y-1.5 text-center'>
        <h1 className='text-2xl font-semibold tracking-tight'>{t('Create an account')}</h1>
        <p className='text-sm text-muted-foreground'>
          {t('Join MedRouter to publish and discover models')}
        </p>
      </div>

      <form className='space-y-4' onSubmit={onSubmit} noValidate>
        <div className='space-y-2'>
          <Label htmlFor='username'>{t('Username')}</Label>
          <Input
            id='username'
            autoComplete='username'
            autoFocus
            placeholder={t('Choose a username')}
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
          <Label htmlFor='email'>{t('Email')}</Label>
          <Input
            id='email'
            type='email'
            autoComplete='email'
            placeholder={t('Enter your email')}
            aria-invalid={Boolean(form.formState.errors.email)}
            {...form.register('email')}
          />
          {form.formState.errors.email ? (
            <p className='text-sm text-destructive'>
              {t(form.formState.errors.email.message ?? '')}
            </p>
          ) : null}
        </div>

        {verificationEnabled ? (
          <div className='space-y-2'>
            <Label htmlFor='verificationCode'>{t('Verification code')}</Label>
            <div className='flex gap-2'>
              <Input
                id='verificationCode'
                inputMode='numeric'
                autoComplete='one-time-code'
                maxLength={6}
                placeholder={t('6-digit code')}
                aria-invalid={Boolean(form.formState.errors.verificationCode)}
                {...form.register('verificationCode')}
              />
              <Button
                type='button'
                variant='outline'
                className='h-10 shrink-0 px-3'
                disabled={isSending || isActive}
                onClick={() => {
                  void sendCode(form.getValues('email'))
                }}
              >
                {isActive
                  ? t('Resend ({{seconds}}s)', { seconds: secondsLeft })
                  : isSending
                    ? t('Sending...')
                    : t('Send code')}
              </Button>
            </div>
            {form.formState.errors.verificationCode ? (
              <p className='text-sm text-destructive'>
                {t(form.formState.errors.verificationCode.message ?? '')}
              </p>
            ) : null}
          </div>
        ) : null}

        <div className='space-y-2'>
          <Label htmlFor='password'>{t('Password')}</Label>
          <PasswordInput
            id='password'
            autoComplete='new-password'
            placeholder={t('Enter your password')}
            aria-invalid={Boolean(form.formState.errors.password)}
            {...form.register('password')}
          />
          {form.formState.errors.password ? (
            <p className='text-sm text-destructive'>
              {t(form.formState.errors.password.message ?? '')}
            </p>
          ) : (
            <p className='text-xs text-muted-foreground'>
              {t('8-20 characters with uppercase, lowercase and a digit.')}
            </p>
          )}
        </div>

        <div className='space-y-2'>
          <Label htmlFor='confirmPassword'>{t('Confirm password')}</Label>
          <PasswordInput
            id='confirmPassword'
            autoComplete='new-password'
            placeholder={t('Re-enter your password')}
            aria-invalid={Boolean(form.formState.errors.confirmPassword)}
            {...form.register('confirmPassword')}
          />
          {form.formState.errors.confirmPassword ? (
            <p className='text-sm text-destructive'>
              {t(form.formState.errors.confirmPassword.message ?? '')}
            </p>
          ) : null}
        </div>

        <Button className='h-11 w-full rounded-lg' type='submit' disabled={form.formState.isSubmitting}>
          {form.formState.isSubmitting ? t('Creating account...') : t('Create account')}
        </Button>
      </form>

      {status?.oauth_providers?.length ? (
        <OAuthButtons providers={status.oauth_providers} redirectTo={redirectTo} />
      ) : null}

      <p className='text-center text-sm text-muted-foreground'>
        {t('Already have an account?')}{' '}
        <Link
          to='/sign-in'
          search={{ redirect: redirectTo }}
          className='font-medium text-foreground underline-offset-4 hover:underline'
        >
          {t('Sign in')}
        </Link>
      </p>
    </div>
  )
}
