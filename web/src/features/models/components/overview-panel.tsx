import { useTranslation } from 'react-i18next'

import { formatDate } from '../format'
import { storageSummaryText, triStateLabel } from '../labels'
import type { ModelDetail } from '../types'
import { UsedByList } from './used-by-list'

interface OverviewPanelProps {
  detail: ModelDetail
}

export function OverviewPanel({ detail }: OverviewPanelProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'
  const latest = detail.versions.find((version) => version.is_latest)

  const specs: [string, string][] = latest
    ? [
        [t('models.fields.architecture'), latest.architecture || '—'],
        [t('models.fields.quantization'), latest.quantization || '—'],
        [t('models.fields.storage'), storageSummaryText(latest.storage, t)],
        [
          t('models.fields.storageResolved'),
          triStateLabel(latest.storage_resolved, t),
        ],
        [t('models.fields.inUse'), triStateLabel(latest.in_use, t)],
        [
          t('models.fields.compatibleProfiles'),
          String(latest.compatible_profiles.length),
        ],
        [t('models.fields.created'), formatDate(latest.created_at, locale)],
      ]
    : []

  return (
    <section
      className='ml-panel'
      role='tabpanel'
      id='ml-panel-overview'
      aria-labelledby='ml-tab-overview'
    >
      <div className='ml-panel-head'>
        <h2>{t('models.detail.overviewTitle')}</h2>
        <span className='ml-meta'>
          {t('models.detail.providedBy', {
            publisher: detail.publisher || t('models.fields.notAnnotated'),
          })}
        </span>
      </div>

      <h3 className='ml-sub-h' style={{ marginTop: 0 }}>
        {t('models.detail.boundaryTitle')}
      </h3>
      <div className='ml-boundary-grid'>
        <div>
          <span className='ml-b-label'>{t('models.fields.task')}</span>
          <p>{detail.task || t('models.fields.notAnnotated')}</p>
        </div>
        <div>
          <span className='ml-b-label'>{t('models.fields.license')}</span>
          <p>{detail.license || t('models.fields.notAnnotated')}</p>
        </div>
        <div>
          <span className='ml-b-label'>{t('models.fields.usage')}</span>
          <p>{detail.usage_boundary || t('models.fields.notAnnotated')}</p>
        </div>
      </div>

      {latest ? (
        <>
          <h3 className='ml-sub-h'>
            {t('models.detail.specTitle', {
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

      <h3 className='ml-sub-h'>{t('models.detail.usedByTitle')}</h3>
      <UsedByList refs={detail.used_by} />
    </section>
  )
}
