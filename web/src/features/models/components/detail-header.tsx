import { Link } from '@tanstack/react-router'
import { ArrowLeft, Check, Copy } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { formatRelativeTime, initials } from '../format'
import { modelDescription } from '../labels'
import type { ModelDetail } from '../types'

interface DetailHeaderProps {
  detail: ModelDetail
  copied: boolean
  onCopy: () => void
}

export function DetailHeader({ detail, copied, onCopy }: DetailHeaderProps) {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'
  const mark =
    detail.publisher_initials || initials(detail.publisher, detail.display_name)
  const meta = [
    t('models.detail.meta.updated', {
      time: formatRelativeTime(detail.updated_at, locale),
    }),
    detail.license
      ? t('models.detail.meta.license', { license: detail.license })
      : null,
    detail.usage_boundary
      ? `${t('models.fields.usage')} · ${detail.usage_boundary}`
      : null,
    t('models.detail.meta.versions', { count: detail.version_count }),
  ].filter((value): value is string => Boolean(value))

  return (
    <>
      <nav className='ml-crumbs' aria-label={t('models.detail.crumbs')}>
        <Link to='/models'>{t('models.detail.allModels')}</Link>
        <span className='ml-sep'>/</span>
        <span>{detail.publisher || t('models.cards.unknownPublisher')}</span>
        <span className='ml-sep'>/</span>
        <span className='ml-current'>{detail.display_name}</span>
      </nav>

      <div className='ml-detail-head'>
        <div className='ml-detail-head-main'>
          <div className='ml-detail-org'>
            <span className='ml-org-mark' aria-hidden='true'>
              {mark}
            </span>
            <span>{detail.publisher || t('models.cards.unknownPublisher')}</span>
          </div>
          <h1 className='ml-detail-title'>{detail.display_name}</h1>
          <p className='ml-detail-desc'>{modelDescription(detail, t)}</p>
          <div className='ml-detail-meta'>
            {meta.map((item) => (
              <span key={item}>{item}</span>
            ))}
          </div>
        </div>

        <div className='ml-detail-actions'>
          <Link to='/models' className='ml-btn ml-btn-secondary'>
            <ArrowLeft aria-hidden='true' />
            {t('models.detail.back')}
          </Link>
          <button type='button' className='ml-btn ml-btn-primary' onClick={onCopy}>
            {copied ? <Check aria-hidden='true' /> : <Copy aria-hidden='true' />}
            {copied ? t('models.detail.copied') : t('models.detail.copy')}
          </button>
        </div>
      </div>
    </>
  )
}
