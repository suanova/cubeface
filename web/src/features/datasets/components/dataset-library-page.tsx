import { getRouteApi } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { ChevronDown, RotateCw, Search } from 'lucide-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { SiteHeader } from '@/components/site-header'
import { EmptyState } from '@/features/models/components/empty-state'
import { NoticeBar } from '@/features/models/components/notice-bar'
import { PageHead } from '@/features/models/components/page-head'
import { UnavailableState } from '@/features/models/components/unavailable-state'
import { formatRelativeTime } from '@/features/models/format'

import {
  activeDatasetFilterCount,
  buildTaskFacets,
  filterDatasets,
  sortDatasets,
} from '../filters'
import { datasetListQueryOptions } from '../queries'
import { DATASET_SORTS, resolveDatasetSort, type DatasetSearch, type DatasetSort } from '../search'
import { DatasetListLoading } from './dataset-list-loading'
import { DatasetListRow } from './dataset-list-row'

const routeApi = getRouteApi('/datasets/')

export function DatasetLibraryPage() {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'
  const search = routeApi.useSearch()
  const navigate = routeApi.useNavigate()
  const { data, isPending, isError, isFetching, refetch } = useQuery(
    datasetListQueryOptions
  )

  const filters = { q: search.q, task: search.task }
  const sort = resolveDatasetSort(search.sort)
  const activeCount = activeDatasetFilterCount(filters)

  const taskFacets = useMemo(() => buildTaskFacets(data?.datasets ?? []), [data])
  const visible = useMemo(
    () => sortDatasets(filterDatasets(data?.datasets ?? [], filters), sort),
    [data, search.q, search.task, sort]
  )
  const formatCount = useMemo(
    () =>
      new Set(
        (data?.datasets ?? [])
          .map((dataset) => dataset.format.trim())
          .filter(Boolean)
      ).size,
    [data]
  )

  const updateSearch = (patch: Partial<DatasetSearch>) => {
    void navigate({
      search: (previous) => ({ ...previous, ...patch }),
      replace: true,
    })
  }

  const resetFilters = () => {
    updateSearch({ q: undefined, task: undefined })
  }

  return (
    <div className='ml-scope'>
      <SiteHeader />
      <main className='ml-main'>
        <div className='ml-container'>
          <PageHead
            kicker={t('datasets.kicker')}
            title={t('datasets.title')}
            lead={t('datasets.lead')}
            notes={
              data?.available
                ? [
                    t('datasets.headNotes.datasets', { count: data.total }),
                    t('datasets.headNotes.versions', { count: data.version_total }),
                    t('datasets.headNotes.formats', { count: formatCount }),
                  ]
                : []
            }
            actions={
              <button
                type='button'
                className='ml-btn ml-btn-secondary'
                onClick={() => void refetch()}
                disabled={isFetching}
              >
                <RotateCw aria-hidden='true' />
                {isFetching ? t('datasets.refreshing') : t('datasets.refresh')}
              </button>
            }
          />

          {isPending ? <DatasetListLoading /> : null}

          {isError ? (
            <UnavailableState
              reason='API_ERROR'
              retryable
              namespace='datasets'
              onRetry={() => void refetch()}
            />
          ) : null}

          {data && !data.available ? (
            <UnavailableState
              reason={data.reason}
              retryable={data.retryable}
              namespace='datasets'
              onRetry={() => void refetch()}
            />
          ) : null}

          {data?.available ? (
            <>
              {data.stale ? (
                <NoticeBar label={t('datasets.states.warning')} warning>
                  {t('datasets.states.stale', {
                    time: formatRelativeTime(data.generated_at, locale),
                  })}
                </NoticeBar>
              ) : null}

              <NoticeBar label={t('datasets.notice.label')}>
                {t('datasets.notice.text')}
              </NoticeBar>

              {taskFacets.length > 0 || search.task ? (
                <div
                  className='ds-chiprow'
                  role='group'
                  aria-label={t('datasets.filters.tasks')}
                >
                  <button
                    type='button'
                    className='ds-chip'
                    aria-pressed={!search.task}
                    onClick={() => updateSearch({ task: undefined })}
                  >
                    {t('datasets.filters.all')}
                  </button>
                  {taskFacets.map((facet) => (
                    <button
                      key={facet.value}
                      type='button'
                      className='ds-chip'
                      aria-pressed={search.task === facet.value}
                      onClick={() =>
                        updateSearch({
                          task: search.task === facet.value ? undefined : facet.value,
                        })
                      }
                    >
                      {facet.value}
                    </button>
                  ))}
                </div>
              ) : null}

              <div className='ds-toolbar'>
                <div className='ml-toolbar-left'>
                  <label className='ml-search'>
                    <Search aria-hidden='true' />
                    <input
                      type='search'
                      value={search.q ?? ''}
                      onChange={(event) =>
                        updateSearch({ q: event.target.value || undefined })
                      }
                      placeholder={t('datasets.search.placeholder')}
                      aria-label={t('datasets.search.placeholder')}
                      autoComplete='off'
                    />
                    {search.q ? (
                      <button
                        type='button'
                        className='ml-search-clear'
                        onClick={() => updateSearch({ q: undefined })}
                        aria-label={t('datasets.search.clear')}
                      >
                        ×
                      </button>
                    ) : null}
                  </label>
                  <span className='ml-result-line'>
                    {activeCount > 0
                      ? t('datasets.results', {
                          count: visible.length,
                          total: data.total,
                        })
                      : t('datasets.resultsAll', { count: data.total })}
                  </span>
                  {activeCount > 0 ? (
                    <button
                      type='button'
                      className='ml-filter-reset'
                      onClick={resetFilters}
                    >
                      {t('datasets.filters.reset')}
                    </button>
                  ) : null}
                </div>

                <div className='ml-toolbar-actions'>
                  <div className='ml-select-wrap'>
                    <select
                      aria-label={t('datasets.sort.label')}
                      value={sort}
                      onChange={(event) => {
                        const value = event.target.value as DatasetSort
                        updateSearch({
                          sort: value === 'updated' ? undefined : value,
                        })
                      }}
                    >
                      {DATASET_SORTS.map((option) => (
                        <option key={option} value={option}>
                          {t(`datasets.sort.${option}`)}
                        </option>
                      ))}
                    </select>
                    <ChevronDown aria-hidden='true' />
                  </div>
                </div>
              </div>

              {visible.length > 0 ? (
                <div className='ds-list'>
                  {visible.map((dataset, index) => (
                    <DatasetListRow
                      key={dataset.dataset}
                      dataset={dataset}
                      index={index}
                    />
                  ))}
                </div>
              ) : (
                <EmptyState
                  filtered={activeCount > 0}
                  namespace='datasets'
                  onReset={resetFilters}
                />
              )}

              <p className='ml-grid-note'>{t('datasets.gridNote')}</p>
            </>
          ) : null}
        </div>
      </main>

      <footer className='ml-foot'>
        <div className='ml-container ml-foot-inner'>
          <span>© 2026 CubeFace</span>
          <span className='ml-mono'>{t('datasets.footer.source')}</span>
        </div>
      </footer>
    </div>
  )
}
