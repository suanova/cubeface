import { useTranslation } from 'react-i18next'

export type DetailTab = 'overview' | 'runtimes' | 'versions'

interface DetailTabsProps {
  active: DetailTab
  onChange: (tab: DetailTab) => void
  versionCount: number
  runtimeCount: number
}

export function DetailTabs({
  active,
  onChange,
  versionCount,
  runtimeCount,
}: DetailTabsProps) {
  const { t } = useTranslation()
  const tabs: { id: DetailTab; label: string; count?: number }[] = [
    { id: 'overview', label: t('models.detail.tabs.overview') },
    { id: 'runtimes', label: t('models.detail.tabs.runtimes'), count: runtimeCount },
    { id: 'versions', label: t('models.detail.tabs.versions'), count: versionCount },
  ]

  return (
    <div className='ml-tabs' role='tablist' aria-label={t('models.detail.tabs.aria')}>
      {tabs.map((tab, index) => (
        <button
          key={tab.id}
          type='button'
          role='tab'
          id={`ml-tab-${tab.id}`}
          aria-selected={active === tab.id}
          aria-controls={`ml-panel-${tab.id}`}
          tabIndex={active === tab.id ? 0 : -1}
          className='ml-tab'
          onClick={() => onChange(tab.id)}
          onKeyDown={(event) => {
            if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return
            event.preventDefault()
            const offset = event.key === 'ArrowRight' ? 1 : -1
            const next = tabs[(index + offset + tabs.length) % tabs.length]
            onChange(next.id)
            document.getElementById(`ml-tab-${next.id}`)?.focus()
          }}
        >
          {tab.label}
          {tab.count !== undefined ? <span className='ml-count'>{tab.count}</span> : null}
        </button>
      ))}
    </div>
  )
}
