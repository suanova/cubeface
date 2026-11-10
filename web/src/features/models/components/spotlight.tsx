import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { formatRelativeTime } from '../format'
import { modelDescription } from '../labels'
import type { ModelSummary } from '../types'

interface SpotlightProps {
  model: ModelSummary
}

export function Spotlight({ model }: SpotlightProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'
  const facts = [
    {
      label: t('models.spotlight.publisher'),
      value: model.publisher || t('models.fields.notAnnotated'),
    },
    { label: t('models.spotlight.latest'), value: model.latest_version || '—' },
    { label: t('models.spotlight.runtimes'), value: String(model.compatible_profile_count) },
    { label: t('models.spotlight.usedBy'), value: String(model.used_by_count) },
  ]

  return (
    <section className='ml-spotlight' aria-label={t('models.spotlight.aria')}>
      <div>
        <p className='ml-sp-kicker'>
          <b>
            {model.featured
              ? t('models.spotlight.featured')
              : t('models.spotlight.recent')}
          </b>
          <span>·</span>
          <span>{model.latest_version || '—'}</span>
        </p>
        <h2 className='ml-sp-title'>{model.display_name}</h2>
        <p className='ml-sp-desc'>{modelDescription(model, t)}</p>
        <div className='ml-sp-foot'>
          <span className='ml-sp-meta'>
            {[
              t('models.cards.versions', { count: model.version_count }),
              t('models.cards.runtimes', { count: model.compatible_profile_count }),
              t('models.cards.updated', {
                time: formatRelativeTime(model.updated_at, locale),
              }),
            ].join(' · ')}
          </span>
          <Link
            to='/models/$name'
            params={{ name: model.model }}
            className='ml-sp-link'
          >
            {t('models.spotlight.open')} <span aria-hidden='true'>→</span>
          </Link>
        </div>
      </div>
      <div className='ml-sp-facts' aria-label={t('models.spotlight.facts')}>
        {facts.map((fact) => (
          <div className='ml-sp-fact' key={fact.label}>
            <span>{fact.label}</span>
            <b>{fact.value}</b>
          </div>
        ))}
      </div>
    </section>
  )
}
