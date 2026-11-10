import { ChevronDown, Search, SlidersHorizontal } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import type { FacetKey, ModelFilterState } from '../filters'
import { FACET_KEYS } from '../filters'
import { MODEL_SORTS, type ModelSort } from '../search'

interface CatalogToolbarProps {
  query: string
  onQueryChange: (value: string) => void
  resultCount: number
  total: number
  filtered: boolean
  activeCount: number
  filters: ModelFilterState
  onClear: (key: 'q' | FacetKey) => void
  onReset: () => void
  sort: ModelSort
  onSortChange: (sort: ModelSort) => void
  filtersOpen: boolean
  onToggleFilters: () => void
}

export function CatalogToolbar({
  query,
  onQueryChange,
  resultCount,
  total,
  filtered,
  activeCount,
  filters,
  onClear,
  onReset,
  sort,
  onSortChange,
  filtersOpen,
  onToggleFilters,
}: CatalogToolbarProps) {
  const { t } = useTranslation()

  return (
    <div className='ml-toolbar'>
      <div className='ml-toolbar-left'>
        <label className='ml-search'>
          <Search aria-hidden='true' />
          <input
            type='search'
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder={t('models.search.placeholder')}
            aria-label={t('models.search.placeholder')}
            autoComplete='off'
          />
          {query ? (
            <button
              type='button'
              className='ml-search-clear'
              onClick={() => onQueryChange('')}
              aria-label={t('models.search.clear')}
            >
              ×
            </button>
          ) : null}
        </label>
        <span className='ml-result-line'>
          {filtered
            ? t('models.results', { count: resultCount, total })
            : t('models.resultsAll', { count: total })}
        </span>
        <div className='ml-active-chips'>
          {query ? (
            <button
              type='button'
              className='ml-chip-clear'
              onClick={() => onClear('q')}
              title={t('models.search.clear')}
            >
              {query} <i aria-hidden='true'>×</i>
            </button>
          ) : null}
          {FACET_KEYS.map((key) => {
            const value = filters[key]
            if (!value) return null
            return (
              <button
                key={key}
                type='button'
                className='ml-chip-clear'
                onClick={() => onClear(key)}
                title={`${t(`models.filters.groups.${key}`)}: ${value}`}
              >
                {value} <i aria-hidden='true'>×</i>
              </button>
            )
          })}
          {activeCount > 0 ? (
            <button type='button' className='ml-filter-reset' onClick={onReset}>
              {t('models.filters.reset')}
            </button>
          ) : null}
        </div>
      </div>

      <div className='ml-toolbar-actions'>
        <button
          type='button'
          className='ml-btn ml-btn-secondary ml-btn-sm ml-filter-toggle'
          onClick={onToggleFilters}
          aria-expanded={filtersOpen}
        >
          <SlidersHorizontal aria-hidden='true' />
          {t('models.filters.toggle')}
          {activeCount > 0 ? ` · ${activeCount}` : ''}
        </button>
        <div className='ml-select-wrap'>
          <select
            aria-label={t('models.sort.label')}
            value={sort}
            onChange={(event) => onSortChange(event.target.value as ModelSort)}
          >
            {MODEL_SORTS.map((option) => (
              <option key={option} value={option}>
                {t(`models.sort.${option}`)}
              </option>
            ))}
          </select>
          <ChevronDown aria-hidden='true' />
        </div>
      </div>
    </div>
  )
}
