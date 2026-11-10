import { useTranslation } from 'react-i18next'

import type { UsedByRef } from '../types'

interface UsedByListProps {
  refs: UsedByRef[]
  emptyTitle?: string
  emptyText?: string
}

export function UsedByList({ refs, emptyTitle, emptyText }: UsedByListProps) {
  const { t } = useTranslation()
  if (refs.length === 0) {
    return (
      <div className='ml-empty ml-empty-compact'>
        <strong>{emptyTitle ?? t('models.states.emptyUsedByTitle')}</strong>
        <span>{emptyText ?? t('models.states.emptyUsedByText')}</span>
      </div>
    )
  }
  return (
    <div className='ml-usedby-list'>
      {refs.map((ref) => (
        <div className='ml-usedby-row' key={`${ref.namespace}/${ref.name}`}>
          <span className='ml-usedby-name'>
            <span>{ref.namespace} / </span>
            {ref.name}
          </span>
          {ref.version ? <span className='ml-usedby-version'>{ref.version}</span> : null}
        </div>
      ))}
    </div>
  )
}
