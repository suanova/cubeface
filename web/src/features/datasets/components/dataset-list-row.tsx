import { Link } from '@tanstack/react-router'
import { ChevronRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { formatRelativeTime } from '@/features/models/format'

import { datasetDescription, recordCountLabel } from '../labels'
import type { DatasetSummary } from '../types'

interface DatasetListRowProps {
  dataset: DatasetSummary
  index: number
}

export function DatasetListRow({ dataset, index }: DatasetListRowProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'

  return (
    <Link
      to='/datasets/$name'
      params={{ name: dataset.dataset }}
      className='ds-row'
    >
      <span className='ds-index'>{String(index + 1).padStart(2, '0')}</span>

      <span className='ds-main'>
        <span className='ds-name'>
          {dataset.display_name}
          {dataset.task ? <span className='ml-task-chip'>{dataset.task}</span> : null}
        </span>
        <span className='ds-desc'>{datasetDescription(dataset, locale, t)}</span>
      </span>

      <span className='ds-meta'>
        <b>{dataset.size || t('datasets.fields.unknown')}</b>
        <span>{recordCountLabel(dataset.record_count, locale, t)}</span>
        <span className='ds-org'>
          {dataset.publisher || t('datasets.cards.unknownPublisher')}
        </span>
      </span>

      <span className='ds-meta'>
        <b>{t('datasets.rows.versions', { count: dataset.version_count })}</b>
        <span>{formatRelativeTime(dataset.updated_at, locale)}</span>
      </span>

      <span className='ds-open'>
        <span className='ds-open-label'>{t('datasets.rows.open')}</span>
        <ChevronRight aria-hidden='true' />
      </span>
    </Link>
  )
}
