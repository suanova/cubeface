import { useTranslation } from 'react-i18next'

import { formatNumber } from '@/features/models/format'

import { storageSummaryText } from '../labels'
import type { DatasetDetail } from '../types'

interface DatasetOverviewPanelProps {
  detail: DatasetDetail
}

export function DatasetOverviewPanel({ detail }: DatasetOverviewPanelProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'
  const latest = detail.versions.find((version) => version.is_latest)

  const specs: [string, string][] = latest
    ? [
        [t('datasets.fields.format'), latest.format || '—'],
        [t('datasets.fields.size'), latest.size || '—'],
        [
          t('datasets.fields.records'),
          latest.record_count != null
            ? formatNumber(latest.record_count, locale)
            : '—',
        ],
        [t('datasets.fields.storage'), storageSummaryText(latest.storage, t)],
        [t('datasets.fields.latest'), detail.latest_version || '—'],
      ]
    : []

  return (
    <section
      className='ml-panel'
      role='tabpanel'
      id='ds-panel-overview'
      aria-labelledby='ds-tab-overview'
    >
      <div className='ml-panel-head'>
        <h2>{t('datasets.detail.overviewTitle')}</h2>
        <span className='ml-meta'>
          {t('datasets.detail.providedBy', {
            publisher: detail.publisher || t('datasets.fields.notAnnotated'),
          })}
        </span>
      </div>

      <h3 className='ml-sub-h' style={{ marginTop: 0 }}>
        {t('datasets.detail.boundaryTitle')}
      </h3>
      <div className='ml-boundary-grid ds-boundary'>
        <div>
          <span className='ml-b-label'>{t('datasets.fields.access')}</span>
          <p>{detail.access || t('datasets.fields.notAnnotated')}</p>
        </div>
        <div>
          <span className='ml-b-label'>{t('datasets.fields.deidentification')}</span>
          <p>{detail.deidentification || t('datasets.fields.notAnnotated')}</p>
        </div>
        <div>
          <span className='ml-b-label'>{t('datasets.fields.ethics')}</span>
          <p>{detail.ethics || t('datasets.fields.notAnnotated')}</p>
        </div>
        <div>
          <span className='ml-b-label'>{t('datasets.fields.sourceRepo')}</span>
          <p>{detail.source_repo || t('datasets.fields.notAnnotated')}</p>
        </div>
      </div>

      {latest ? (
        <>
          <h3 className='ml-sub-h'>
            {t('datasets.detail.specTitle', {
              version: detail.latest_version || latest.ref,
            })}
          </h3>
          <dl className='ml-spec-list'>
            {specs.map(([label, value]) => (
              <div className='ml-spec-row' key={label}>
                <span>{label}</span>
                <b>{value}</b>
              </div>
            ))}
          </dl>
        </>
      ) : null}

      <h3 className='ml-sub-h'>{t('datasets.detail.descriptionTitle')}</h3>
      <p className='ml-readme'>{detail.description || t('datasets.fields.notAnnotated')}</p>
    </section>
  )
}
