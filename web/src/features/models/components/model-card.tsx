import { Link } from '@tanstack/react-router'
import { Cpu, Layers } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { formatRelativeTime, initials } from '../format'
import { modelDescription } from '../labels'
import type { ModelSummary } from '../types'

interface ModelCardProps {
  model: ModelSummary
}

function isGated(boundary: string): boolean {
  return /申请|gated|restricted|research only|仅研究|需申请/i.test(boundary)
}

export function ModelCard({ model }: ModelCardProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'
  const mark =
    model.publisher_initials || initials(model.publisher, model.display_name)

  return (
    <Link
      to='/models/$name'
      params={{ name: model.model }}
      className='ml-model-card'
    >
      <div className='ml-model-top'>
        <span className='ml-org-mark' aria-hidden='true'>
          {mark}
        </span>
        <div className='ml-model-id'>
          <span className='ml-model-org'>
            {model.publisher || t('models.cards.unknownPublisher')}
          </span>
          <span className='ml-model-name'>{model.display_name}</span>
        </div>
        {model.task ? <span className='ml-task-chip'>{model.task}</span> : null}
      </div>

      <p className='ml-model-desc'>{modelDescription(model, t)}</p>

      <div className='ml-model-tags'>
        {model.usage_boundary ? (
          <span
            className={
              isGated(model.usage_boundary)
                ? 'ml-usage-badge is-gated'
                : 'ml-usage-badge'
            }
          >
            {model.usage_boundary}
          </span>
        ) : null}
        {model.tags.slice(0, 2).map((tag) => (
          <span className='ml-tag' key={tag}>
            {tag}
          </span>
        ))}
      </div>

      <div className='ml-model-foot'>
        <span
          className='ml-metric'
          title={t('models.cards.versions', { count: model.version_count })}
        >
          <Layers aria-hidden='true' />
          {model.version_count}
        </span>
        <span
          className='ml-metric'
          title={t('models.cards.runtimes', {
            count: model.compatible_profile_count,
          })}
        >
          <Cpu aria-hidden='true' />
          {model.compatible_profile_count}
        </span>
        <span className='ml-updated'>
          {t('models.cards.updated', {
            time: formatRelativeTime(model.updated_at, locale),
          })}
        </span>
      </div>
    </Link>
  )
}
