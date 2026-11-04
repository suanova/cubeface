import GoogleColor from '@lobehub/icons/es/Google/components/Color'
import GithubIcon from '@lobehub/icons/es/Github'
import { Globe } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { buttonVariants } from '@/components/ui/button'
import type { OAuthProviderInfo } from '@/features/auth/types'
import { withBasePath } from '@/lib/base-path'
import { cn } from '@/lib/utils'

interface OAuthButtonsProps {
  providers: OAuthProviderInfo[]
  redirectTo?: string
}

export function OAuthButtons({ providers, redirectTo }: OAuthButtonsProps) {
  const { t } = useTranslation()
  if (providers.length === 0) return null

  const redirect = encodeURIComponent(redirectTo || '/dashboard')

  return (
    <div className='space-y-3'>
      <div className='relative'>
        <div className='absolute inset-0 flex items-center'>
          <span className='w-full border-t' />
        </div>
        <div className='relative flex justify-center text-xs uppercase'>
          <span className='bg-card px-2 text-muted-foreground'>{t('Or continue with')}</span>
        </div>
      </div>

      <div className='flex flex-col gap-2'>
        {providers.map((provider) => {
          const label =
            provider.id === 'github'
              ? t('Continue with GitHub')
              : provider.id === 'google'
                ? t('Continue with Google')
                : t('Continue with {{name}}', { name: provider.name })
          return (
            <a
              key={provider.id}
              href={withBasePath(`/api/oauth/${provider.id}/start?redirect=${redirect}`)}
              className={cn(
                buttonVariants({ variant: 'outline' }),
                'h-11 w-full justify-center gap-2 rounded-lg'
              )}
            >
              {provider.id === 'github' ? (
                <GithubIcon className='h-4 w-4 dark:invert' />
              ) : provider.id === 'google' ? (
                <GoogleColor className='h-4 w-4' />
              ) : (
                <Globe className='h-4 w-4' />
              )}
              {label}
            </a>
          )
        })}
      </div>
    </div>
  )
}
