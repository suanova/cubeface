import { useTranslation } from 'react-i18next'

import {
  formatList,
  servingModeLabel,
  triStateLabel,
} from '../labels'
import type { RuntimeProfileSummary } from '../types'
import { NoticeBar } from './notice-bar'

interface RuntimeProfileListProps {
  profiles: RuntimeProfileSummary[]
  profilesAvailable: boolean
}

export function RuntimeProfileList({
  profiles,
  profilesAvailable,
}: RuntimeProfileListProps) {
  const { t } = useTranslation()

  if (!profilesAvailable) {
    return (
      <>
        <NoticeBar label={t('models.states.warning')} warning>
          {t('models.states.profilesUnavailableDetail')}
        </NoticeBar>
        <div className='ml-empty'>
          <strong>{t('models.states.emptyRuntimeTitle')}</strong>
          <span>{t('models.states.emptyRuntimeText')}</span>
        </div>
      </>
    )
  }

  if (profiles.length === 0) {
    return (
      <div className='ml-empty'>
        <strong>{t('models.states.emptyRuntimeTitle')}</strong>
        <span>{t('models.states.emptyRuntimeText')}</span>
      </div>
    )
  }

  return (
    <div className='ml-profile-list'>
      {profiles.map((profile) => {
        const engine = [profile.engine.name, profile.engine.version]
          .filter(Boolean)
          .join(' ')
        return (
          <article className='ml-profile-card' key={profile.name}>
            <div className='ml-profile-head'>
              <div>
                <h3 className='ml-profile-name'>{profile.name}</h3>
                <p className='ml-profile-engine'>
                  {t('models.fields.engine')} · {engine || '—'}
                </p>
              </div>
              <div className='ml-profile-badges'>
                {profile.assets_resolved !== null ? (
                  <span className='ml-status-pill'>
                    <span className='ml-dot' aria-hidden='true' />
                    {t('models.fields.assetsResolved')} ·{' '}
                    {triStateLabel(profile.assets_resolved, t)}
                  </span>
                ) : null}
                {profile.deprecated === true ? (
                  <span className='ml-status-pill is-warn'>
                    {t('models.fields.deprecated')}
                  </span>
                ) : null}
                {profile.drifted === true ? (
                  <span className='ml-status-pill is-warn'>
                    {t('models.fields.drifted')}
                  </span>
                ) : null}
              </div>
            </div>

            {profile.description ? (
              <p className='ml-profile-desc'>{profile.description}</p>
            ) : null}

            <div className='ml-profile-grid'>
              <div className='ml-profile-field'>
                <span>{t('models.fields.vendor')}</span>
                <b>{profile.vendor || '—'}</b>
              </div>
              <div className='ml-profile-field'>
                <span>{t('models.fields.accelerators')}</span>
                <b>{formatList(profile.accelerator_models)}</b>
              </div>
              <div className='ml-profile-field'>
                <span>{t('models.fields.servingMode')}</span>
                <b>{servingModeLabel(profile.serving_mode, t)}</b>
              </div>
              <div className='ml-profile-field'>
                <span>{t('models.fields.gpuPerPod')}</span>
                <b>{profile.gpu_per_pod === null ? '—' : profile.gpu_per_pod}</b>
              </div>
              <div className='ml-profile-field'>
                <span>{t('models.fields.architecture')}</span>
                <b>{formatList(profile.architectures)}</b>
              </div>
              <div className='ml-profile-field'>
                <span>{t('models.fields.quantization')}</span>
                <b>{formatList(profile.quantizations)}</b>
              </div>
              <div className='ml-profile-field'>
                <span>{t('models.fields.compatibleVersions')}</span>
                <b>{formatList(profile.compatible_versions)}</b>
              </div>
              <div className='ml-profile-field'>
                <span>{t('models.fields.usedByCount')}</span>
                <b>{profile.used_by_count}</b>
              </div>
            </div>

            {profile.roles.length > 0 ? (
              <div className='ml-profile-roles'>
                {profile.roles.map((role) => (
                  <div className='ml-role-row' key={role.name}>
                    <span>{role.name}</span>
                    <span>
                      {formatList(
                        [role.vendor, ...role.accelerator_models].filter(Boolean)
                      )}
                    </span>
                    <b>
                      {role.gpu_per_pod === null
                        ? '—'
                        : `${role.gpu_per_pod} GPU`}
                    </b>
                  </div>
                ))}
              </div>
            ) : null}
          </article>
        )
      })}
    </div>
  )
}
