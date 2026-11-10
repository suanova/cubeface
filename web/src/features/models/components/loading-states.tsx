import { useTranslation } from 'react-i18next'

function Skeleton({ width, height = 12 }: { width: string; height?: number }) {
  return <span className='ml-skeleton' style={{ width, height }} />
}

export type CatalogNamespace = 'models' | 'datasets'

export function CatalogLoading({
  count = 6,
  namespace = 'models',
}: {
  count?: number
  namespace?: CatalogNamespace
}) {
  const { t } = useTranslation()
  return (
    <div
      className='ml-model-grid'
      aria-busy='true'
      aria-label={t(`${namespace}.states.loading`)}
    >
      {Array.from({ length: count }, (_, index) => (
        <div className='ml-skeleton-card' key={index}>
          <div style={{ display: 'flex', gap: 9, alignItems: 'flex-start' }}>
            <span className='ml-skeleton' style={{ width: 28, height: 28, flex: 'none' }} />
            <div style={{ flex: 1 }}>
              <Skeleton width='55%' height={10} />
              <Skeleton width='80%' height={14} />
            </div>
          </div>
          <div style={{ marginTop: 16 }}>
            <Skeleton width='100%' height={11} />
            <Skeleton width='88%' height={11} />
          </div>
          <div style={{ marginTop: 16, display: 'flex', gap: 6 }}>
            <Skeleton width='64px' height={20} />
            <Skeleton width='48px' height={20} />
          </div>
          <div style={{ marginTop: 18, paddingTop: 12, borderTop: '1px solid var(--ml-border)' }}>
            <Skeleton width='45%' height={10} />
          </div>
        </div>
      ))}
    </div>
  )
}

export function DetailLoading({ namespace = 'models' }: { namespace?: CatalogNamespace }) {
  const { t } = useTranslation()
  return (
    <div
      className='ml-detail-grid'
      aria-busy='true'
      aria-label={t(`${namespace}.states.loading`)}
    >
      <div className='ml-panel'>
        <Skeleton width='30%' height={18} />
        <div style={{ marginTop: 18 }}>
          <Skeleton width='100%' height={12} />
          <Skeleton width='94%' height={12} />
          <Skeleton width='72%' height={12} />
        </div>
        <div style={{ marginTop: 28, display: 'grid', gap: 10 }}>
          <Skeleton width='100%' height={54} />
          <Skeleton width='100%' height={54} />
        </div>
      </div>
      <div className='ml-detail-side'>
        <div className='ml-side-card'>
          <Skeleton width='55%' height={12} />
          <div style={{ marginTop: 14 }}>
            <Skeleton width='100%' height={10} />
            <Skeleton width='100%' height={10} />
            <Skeleton width='80%' height={10} />
          </div>
        </div>
        <div className='ml-side-card'>
          <Skeleton width='65%' height={12} />
          <div style={{ marginTop: 14 }}>
            <Skeleton width='100%' height={10} />
            <Skeleton width='70%' height={10} />
          </div>
        </div>
      </div>
    </div>
  )
}
