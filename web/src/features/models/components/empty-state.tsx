import { useTranslation } from 'react-i18next'

interface EmptyStateProps {
  filtered: boolean
  onReset?: () => void
}

export function EmptyState({ filtered, onReset }: EmptyStateProps) {
  const { t } = useTranslation()
  return (
    <div className='ml-empty'>
      <strong>
        {filtered ? t('models.empty.title') : t('models.empty.catalogTitle')}
      </strong>
      <span>{filtered ? t('models.empty.text') : t('models.empty.catalogText')}</span>
      {filtered && onReset ? (
        <div className='ml-state-actions' style={{ justifyContent: 'center' }}>
          <button type='button' className='ml-btn ml-btn-secondary ml-btn-sm' onClick={onReset}>
            {t('models.filters.reset')}
          </button>
        </div>
      ) : null}
    </div>
  )
}
