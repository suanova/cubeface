import { useTranslation } from 'react-i18next'

import { formatNumber } from '@/features/models/format'

import type { DatasetDetail } from '../types'

interface DatasetInfoCardProps {
  detail: DatasetDetail
}

export function DatasetInfoCard({ detail }: DatasetInfoCardProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'
  const rows = [
    [t('datasets.fields.datasetName'), detail.dataset],
    [t('datasets.fields.format'), detail.format || '—'],
    [t('datasets.fields.size'), detail.size || '—'],
    [
      t('datasets.fields.records'),
      detail.record_count != null ? formatNumber(detail.record_count, locale) : '—',
    ],
    [t('datasets.fields.latest'), detail.latest_version || '—'],
    [t('datasets.fields.versions'), String(detail.version_count)],
    [t('datasets.fields.license'), detail.license || '—'],
    [t('datasets.fields.storage'), detail.storage_strategy || '—'],
  ]

  return (
    <div className='ml-side-card'>
      <h3>{t('datasets.detail.sideInfo')}</h3>
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
