import { useTranslation } from 'react-i18next'

import type { FacetKey, FacetValue, Facets } from '../filters'
import type { ModelFilterState } from '../filters'

interface FilterSidebarProps {
  facets: Facets
  filters: ModelFilterState
  open: boolean
  onChange: (key: FacetKey, value: string | undefined) => void
  onReset: () => void
  onClose: () => void
}

const GROUPS: FacetKey[] = ['task', 'arch', 'quant', 'license', 'usage']

export function FilterSidebar({
  facets,
  filters,
  open,
  onChange,
  onReset,
  onClose,
}: FilterSidebarProps) {
  const { t } = useTranslation()

  return (
    <aside
      className={open ? 'ml-filters is-open' : 'ml-filters'}
      aria-label={t('models.filters.title')}
    >
      <div className='ml-filter-drawer-head'>
        <span>{t('models.filters.title')}</span>
        <button type='button' onClick={onClose}>
          {t('models.filters.close')}
        </button>
      </div>

      {GROUPS.map((key) => {
        const options: FacetValue[] = [...facets[key]]
        const selected = filters[key]
        if (selected && !options.some((option) => option.value === selected)) {
          options.push({ value: selected, count: 0 })
        }
        if (options.length === 0) return null
        return (
          <div className='ml-filter-group' key={key}>
            <h2 className='ml-filter-title'>{t(`models.filters.groups.${key}`)}</h2>
            <div role='group' aria-label={t(`models.filters.groups.${key}`)}>
              <button
                type='button'
                className={selected ? 'ml-filter-opt' : 'ml-filter-opt is-active'}
                aria-pressed={!selected}
                onClick={() => onChange(key, undefined)}
              >
                {t('models.filters.all')}
              </button>
              {options.map((option) => (
                <button
                  key={option.value}
                  type='button'
                  className={
                    selected === option.value
                      ? 'ml-filter-opt is-active'
                      : 'ml-filter-opt'
                  }
                  aria-pressed={selected === option.value}
                  onClick={() => onChange(key, option.value)}
                >
                  {option.value}
                </button>
              ))}
            </div>
          </div>
        )
      })}

      <button type='button' className='ml-filter-reset' onClick={onReset}>
        {t('models.filters.reset')}
      </button>
    </aside>
  )
}
