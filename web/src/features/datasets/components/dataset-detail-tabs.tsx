import { useTranslation } from 'react-i18next'

export type DatasetTab = 'overview' | 'versions'

interface DatasetDetailTabsProps {
  active: DatasetTab
  onChange: (tab: DatasetTab) => void
  versionCount: number
}

export function DatasetDetailTabs({
  active,
  onChange,
  versionCount,
}: DatasetDetailTabsProps) {
  const { t } = useTranslation()
  const tabs: { id: DatasetTab; label: string; count?: number }[] = [
    { id: 'overview', label: t('datasets.detail.tabs.overview') },
    { id: 'versions', label: t('datasets.detail.tabs.versions'), count: versionCount },
  ]

  return (
    <div className='ml-tabs' role='tablist' aria-label={t('datasets.detail.tabs.aria')}>
      {tabs.map((tab, index) => (
        <button
          key={tab.id}
          type='button'
          role='tab'
          id={`ds-tab-${tab.id}`}
          aria-selected={active === tab.id}
          aria-controls={`ds-panel-${tab.id}`}
          tabIndex={active === tab.id ? 0 : -1}
          className='ml-tab'
          onClick={() => onChange(tab.id)}
          onKeyDown={(event) => {
            if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return
            event.preventDefault()
            const offset = event.key === 'ArrowRight' ? 1 : -1
            const next = tabs[(index + offset + tabs.length) % tabs.length]
            onChange(next.id)
            document.getElementById(`ds-tab-${next.id}`)?.focus()
          }}
        >
          {tab.label}
          {tab.count !== undefined ? <span className='ml-count'>{tab.count}</span> : null}
        </button>
      ))}
    </div>
  )
}
