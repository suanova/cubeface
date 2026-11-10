import { useTranslation } from 'react-i18next'

import { formatDate } from '../format'
import { storageSummaryText, triStateLabel } from '../labels'
import type { ModelVersionDetail } from '../types'

interface VersionListProps {
  versions: ModelVersionDetail[]
  profilesAvailable: boolean
}

export function VersionList({ versions, profilesAvailable }: VersionListProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'

  if (versions.length === 0) {
    return (
      <div className='ml-empty'>
        <strong>{t('models.states.emptyVersionsTitle')}</strong>
        <span>{t('models.states.emptyVersionsText')}</span>
      </div>
    )
  }

  return (
    <div className='ml-version-list'>
      {versions.map((version) => (
        <article className='ml-version-card' key={version.ref}>
          <div className='ml-version-head'>
            <span className='ml-ver'>{version.version || version.ref}</span>
            {version.is_latest ? (
              <span className='ml-latest-badge'>{t('models.fields.latestBadge')}</span>
            ) : null}
            <span className='ml-date'>{formatDate(version.created_at, locale)}</span>
          </div>

          {version.description ? (
            <p className='ml-version-note'>{version.description}</p>
          ) : null}

          <div className='ml-version-body'>
            <div className='ml-version-specs'>
              <span>
                {t('models.fields.architecture')}{' '}
                <b>{version.architecture || '—'}</b>
              </span>
              <span>
                {t('models.fields.quantization')}{' '}
                <b>{version.quantization || '—'}</b>
              </span>
              <span>
                {t('models.fields.storage')}{' '}
                <b>{storageSummaryText(version.storage, t)}</b>
              </span>
              <span>
                {t('models.fields.compatibleProfiles')}{' '}
                <b>
                  {profilesAvailable ? version.compatible_profiles.length : '—'}
                </b>
              </span>
            </div>
            <div className='ml-version-badges'>
              {version.storage_resolved !== null ? (
                <span className='ml-status-pill'>
                  <span className='ml-dot' aria-hidden='true' />
                  {t('models.fields.storageResolved')} ·{' '}
                  {triStateLabel(version.storage_resolved, t)}
                </span>
              ) : null}
              {version.in_use !== null ? (
                <span className='ml-status-pill'>
                  <span className='ml-dot' aria-hidden='true' />
                  {t('models.fields.inUse')} · {triStateLabel(version.in_use, t)}
                </span>
              ) : null}
            </div>
          </div>

          {version.used_by.length > 0 ? (
            <p className='ml-version-usedby'>
              {t('models.fields.usedBy')}:{' '}
              {version.used_by
                .map((ref) => `${ref.namespace}/${ref.name}`)
                .join(' · ')}
            </p>
          ) : null}
        </article>
      ))}
    </div>
  )
}
