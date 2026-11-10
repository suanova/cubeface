import { RotateCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'

interface UnavailableStateProps {
  reason?: string
  retryable?: boolean
  onRetry?: () => void
  wide?: boolean
}

const REASON_KEYS: Record<string, string> = {
  DISABLED: 'disabled',
  NO_CLUSTER: 'noCluster',
  CRD_MISSING: 'crdMissing',
  FORBIDDEN: 'forbidden',
  UNAUTHORIZED: 'unauthorized',
  TIMEOUT: 'timeout',
  API_ERROR: 'apiError',
}

export function UnavailableState({
  reason,
  retryable = false,
  onRetry,
  wide = true,
}: UnavailableStateProps) {
  const { t } = useTranslation()
  const key = REASON_KEYS[reason ?? ''] ?? 'apiError'

  return (
    <div className={wide ? 'ml-state-card is-wide' : 'ml-state-card'} role='status'>
      <p className='ml-state-kicker'>{t('models.states.unavailable.kicker')}</p>
      <h2 className='ml-state-title'>{t('models.states.unavailable.title')}</h2>
      <p className='ml-state-text'>{t(`models.states.unavailable.${key}`)}</p>
      {reason ? <span className='ml-state-code'>{reason}</span> : null}
      {retryable && onRetry ? (
        <div className='ml-state-actions'>
          <button type='button' className='ml-btn ml-btn-secondary' onClick={onRetry}>
            <RotateCw aria-hidden='true' />
            {t('models.states.unavailable.retry')}
          </button>
        </div>
      ) : null}
    </div>
  )
}
