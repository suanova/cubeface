import { useTranslation } from 'react-i18next'

function Skeleton({ width, height = 11 }: { width: string; height?: number }) {
  return <span className='ml-skeleton' style={{ width, height }} />
}

export function DatasetListLoading({ count = 6 }: { count?: number }) {
  const { t } = useTranslation()
  return (
    <div className='ds-list' aria-busy='true' aria-label={t('datasets.states.loading')}>
      {Array.from({ length: count }, (_, index) => (
        <div className='ds-row ds-row-skeleton' key={index}>
          <span className='ds-index'>{String(index + 1).padStart(2, '0')}</span>
          <span className='ds-main'>
            <Skeleton width='34%' height={14} />
            <div style={{ marginTop: 8 }}>
              <Skeleton width='88%' />
            </div>
          </span>
          <span className='ds-meta'>
            <Skeleton width='56px' height={11} />
            <Skeleton width='72px' height={10} />
          </span>
          <span className='ds-meta'>
            <Skeleton width='64px' height={11} />
            <Skeleton width='48px' height={10} />
          </span>
          <span />
        </div>
      ))}
    </div>
  )
}
