import { useTranslation } from 'react-i18next'

import { formatDate, formatNumber } from '@/features/models/format'

import { storageSummaryText } from '../labels'
import type { DatasetVersionDetail } from '../types'

interface DatasetVersionListProps {
  versions: DatasetVersionDetail[]
}

export function DatasetVersionList({ versions }: DatasetVersionListProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'

  if (versions.length === 0) {
    return (
      <div className='ml-empty ml-empty-compact'>
        <strong>{t('datasets.states.emptyVersionsTitle')}</strong>
        <span>{t('datasets.states.emptyVersionsText')}</span>
      </div>
    )
  }

  return (
    <div className='ds-version-list'>
      {versions.map((version) => (
        <article className='ds-version' key={version.ref}>
          <div className='ds-version-head'>
            <span className='ds-version-name'>{version.version}</span>
            {version.is_latest ? (
              <span className='ml-tag'>{t('datasets.fields.latestBadge')}</span>
            ) : null}
            <span className='ds-version-ref ml-mono'>{version.ref}</span>
          </div>

          <dl className='ml-spec-list'>
            <div className='ml-spec-row'>
              <span>{t('datasets.fields.format')}</span>
              <b>{version.format || '—'}</b>
            </div>
            <div className='ml-spec-row'>
              <span>{t('datasets.fields.size')}</span>
              <b>{version.size || '—'}</b>
            </div>
            <div className='ml-spec-row'>
              <span>{t('datasets.fields.records')}</span>
              <b>
                {version.record_count != null
                  ? formatNumber(version.record_count, locale)
                  : '—'}
              </b>
            </div>
            <div className='ml-spec-row'>
              <span>{t('datasets.fields.storage')}</span>
              <b>{storageSummaryText(version.storage, t)}</b>
            </div>
            <div className='ml-spec-row'>
              <span>{t('datasets.fields.created')}</span>
              <b>{formatDate(version.created_at, locale)}</b>
            </div>
          </dl>

          {version.description ? (
            <p className='ds-version-desc'>{version.description}</p>
          ) : null}
        </article>
      ))}
    </div>
  )
}
