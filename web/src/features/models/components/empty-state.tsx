import { useTranslation } from 'react-i18next'

import type { CatalogNamespace } from './loading-states'

interface EmptyStateProps {
  filtered: boolean
  onReset?: () => void
  namespace?: CatalogNamespace
}

export function EmptyState({ filtered, onReset, namespace = 'models' }: EmptyStateProps) {
  const { t } = useTranslation()
  return (
    <div className='ml-empty'>
      <strong>
        {filtered
          ? t(`${namespace}.empty.title`)
          : t(`${namespace}.empty.catalogTitle`)}
      </strong>
      <span>
        {filtered ? t(`${namespace}.empty.text`) : t(`${namespace}.empty.catalogText`)}
      </span>
      {filtered && onReset ? (
        <div className='ml-state-actions' style={{ justifyContent: 'center' }}>
          <button type='button' className='ml-btn ml-btn-secondary ml-btn-sm' onClick={onReset}>
            {t(`${namespace}.filters.reset`)}
          </button>
        </div>
      ) : null}
    </div>
  )
}
