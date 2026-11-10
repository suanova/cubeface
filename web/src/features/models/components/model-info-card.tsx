import { useTranslation } from 'react-i18next'

import type { ModelDetail } from '../types'

interface ModelInfoCardProps {
  detail: ModelDetail
}

export function ModelInfoCard({ detail }: ModelInfoCardProps) {
  const { t } = useTranslation()
  const latest = detail.versions.find((version) => version.is_latest)
  const rows = [
    [t('models.fields.architecture'), detail.architecture || '—'],
    [t('models.fields.quantization'), detail.quantization || '—'],
    [t('models.fields.latest'), detail.latest_version || '—'],
    [t('models.fields.versions'), String(detail.version_count)],
    [t('models.fields.license'), detail.license || '—'],
    [
      t('models.fields.storage'),
      latest?.storage.strategy || detail.storage_strategy || '—',
    ],
    [t('models.fields.runtimes'), String(detail.compatible_profiles.length)],
    [t('models.fields.usedBy'), String(detail.used_by.length)],
  ]

  return (
    <div className='ml-side-card'>
      <h3>{t('models.detail.sideInfo')}</h3>
      <dl className='ml-kv'>
        {rows.map(([label, value]) => (
          <div key={label}>
            <dt>{label}</dt>
            <dd>{value}</dd>
          </div>
        ))}
      </dl>
    </div>
  )
}
