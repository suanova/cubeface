import { Link } from '@tanstack/react-router'
import { ArrowLeft, Check, Copy } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { formatRelativeTime, initials } from '@/features/models/format'

import { datasetDescription } from '../labels'
import type { DatasetDetail } from '../types'

interface DatasetDetailHeaderProps {
  detail: DatasetDetail
  copied: boolean
  onCopy: () => void
}

export function DatasetDetailHeader({
  detail,
  copied,
  onCopy,
}: DatasetDetailHeaderProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'
  const mark =
    detail.publisher_initials || initials(detail.publisher, detail.display_name)
  const meta = [
    t('datasets.detail.meta.updated', {
      time: formatRelativeTime(detail.updated_at, locale),
    }),
    detail.license
      ? t('datasets.detail.meta.license', { license: detail.license })
      : null,
    detail.format
      ? t('datasets.detail.meta.format', { format: detail.format })
      : null,
    t('datasets.detail.meta.versions', { count: detail.version_count }),
  ].filter((value): value is string => Boolean(value))

  return (
    <>
      <nav className='ml-crumbs' aria-label={t('datasets.detail.crumbs')}>
        <Link to='/datasets'>{t('datasets.detail.allDatasets')}</Link>
        <span className='ml-sep'>/</span>
        <span>{detail.publisher || t('datasets.cards.unknownPublisher')}</span>
        <span className='ml-sep'>/</span>
        <span className='ml-current'>{detail.display_name}</span>
      </nav>

      <div className='ml-detail-head'>
        <div className='ml-detail-head-main'>
          <div className='ml-detail-org'>
            <span className='ml-org-mark' aria-hidden='true'>
              {mark}
            </span>
            <span>{detail.publisher || t('datasets.cards.unknownPublisher')}</span>
          </div>
          <h1 className='ml-detail-title'>{detail.display_name}</h1>
          <p className='ml-detail-desc'>{datasetDescription(detail, locale, t)}</p>
          <div className='ml-detail-meta'>
            {meta.map((item) => (
              <span key={item}>{item}</span>
            ))}
          </div>
        </div>

        <div className='ml-detail-actions'>
          <Link to='/datasets' className='ml-btn ml-btn-secondary'>
            <ArrowLeft aria-hidden='true' />
            {t('datasets.detail.back')}
          </Link>
          <button type='button' className='ml-btn ml-btn-primary' onClick={onCopy}>
            {copied ? <Check aria-hidden='true' /> : <Copy aria-hidden='true' />}
            {copied ? t('datasets.detail.copied') : t('datasets.detail.copy')}
          </button>
        </div>
      </div>
    </>
  )
}
