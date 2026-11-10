import { getRouteApi } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { RotateCw } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { SiteHeader } from '@/components/site-header'

import {
  activeFilterCount,
  buildFacets,
  filterModels,
  hasFacet,
  sortModels,
  type FacetKey,
  type ModelFilterState,
} from '../filters'
import { formatRelativeTime } from '../format'
import { modelListQueryOptions } from '../queries'
import { resolveSort, type ModelSearch } from '../search'
import { CatalogLoading } from './loading-states'
import { CatalogToolbar } from './catalog-toolbar'
import { EmptyState } from './empty-state'
import { FilterSidebar } from './filter-sidebar'
import { ModelCard } from './model-card'
import { NoticeBar } from './notice-bar'
import { PageHead } from './page-head'
import { Spotlight } from './spotlight'
import { UnavailableState } from './unavailable-state'

const routeApi = getRouteApi('/models/')

function newestModel<T extends { updated_at: string }>(models: T[]): T {
  return [...models].sort(
    (a, b) => Date.parse(b.updated_at) - Date.parse(a.updated_at)
  )[0]
}

export function ModelLibraryPage() {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'
  const search = routeApi.useSearch()
  const navigate = routeApi.useNavigate()
  const { data, isPending, isError, isFetching, refetch } = useQuery(
    modelListQueryOptions
  )
  const [filtersOpen, setFiltersOpen] = useState(false)

  const filters: ModelFilterState = {
    q: search.q,
    task: search.task,
    arch: search.arch,
    quant: search.quant,
    license: search.license,
    usage: search.usage,
  }
  const sort = resolveSort(search.sort)
  const activeCount = activeFilterCount(filters)

  const facets = useMemo(() => buildFacets(data?.models ?? []), [data])
  const visible = useMemo(
    () => sortModels(filterModels(data?.models ?? [], filters), sort),
    [data, search.q, search.task, search.arch, search.quant, search.license, search.usage, sort]
  )

  const updateSearch = (patch: Partial<ModelSearch>) => {
    void navigate({
      search: (previous) => ({ ...previous, ...patch }),
      replace: true,
    })
  }

  const updateFacet = (key: FacetKey, value: string | undefined) => {
    updateSearch({ [key]: value })
  }

  const resetFilters = () => {
    updateSearch({
      q: undefined,
      task: undefined,
      arch: undefined,
      quant: undefined,
      license: undefined,
      usage: undefined,
    })
  }

  const spotlight = useMemo(() => {
    if (!data || data.models.length === 0) return null
    return data.models.find((model) => model.featured) ?? newestModel(data.models)
  }, [data])

  return (
    <div className='ml-scope'>
      <SiteHeader />
      <main className='ml-main'>
        <div className='ml-container'>
          <PageHead
            kicker={t('models.kicker')}
            title={t('models.title')}
            lead={t('models.lead')}
            notes={
              data?.available
                ? [
                    t('models.headNotes.models', { count: data.total }),
                    t('models.headNotes.versions', { count: data.version_total }),
                    ...(data.profiles_available
                      ? [
                          t('models.headNotes.profiles', {
                            count: data.profile_total,
                          }),
                        ]
                      : []),
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
                {isFetching ? t('models.refreshing') : t('models.refresh')}
              </button>
            }
          />

          {isPending ? <CatalogLoading /> : null}

          {isError ? (
            <UnavailableState
              reason='API_ERROR'
              retryable
              onRetry={() => void refetch()}
            />
          ) : null}

          {data && !data.available ? (
            <UnavailableState
              reason={data.reason}
              retryable={data.retryable}
              onRetry={() => void refetch()}
            />
          ) : null}

          {data?.available ? (
            <>
              {spotlight ? <Spotlight model={spotlight} /> : null}

              {data.stale ? (
                <NoticeBar label={t('models.states.warning')} warning>
                  {t('models.states.stale', {
                    time: formatRelativeTime(data.generated_at, locale),
                  })}
                </NoticeBar>
              ) : null}

              {!data.profiles_available ? (
                <NoticeBar label={t('models.states.warning')} warning>
                  {t('models.states.profileUnavailable')}
                </NoticeBar>
              ) : null}

              <NoticeBar label={t('models.notice.label')}>
                {t('models.notice.text')}
              </NoticeBar>

              <div className='ml-catalog-shell'>
                {hasFacet(facets) || activeCount > 0 ? (
                  <FilterSidebar
                    facets={facets}
                    filters={filters}
                    open={filtersOpen}
                    onChange={updateFacet}
                    onReset={resetFilters}
                    onClose={() => setFiltersOpen(false)}
                  />
                ) : null}

                <div className='ml-catalog-main'>
                  <CatalogToolbar
                    query={search.q ?? ''}
                    onQueryChange={(value) => updateSearch({ q: value || undefined })}
                    resultCount={visible.length}
                    total={data.total}
                    filtered={activeCount > 0}
                    activeCount={activeCount}
                    filters={filters}
                    onClear={(key) =>
                      key === 'q'
                        ? updateSearch({ q: undefined })
                        : updateFacet(key, undefined)
                    }
                    onReset={resetFilters}
                    sort={sort}
                    onSortChange={(value) =>
                      updateSearch({ sort: value === 'featured' ? undefined : value })
                    }
                    filtersOpen={filtersOpen}
                    onToggleFilters={() => setFiltersOpen((open) => !open)}
                  />

                  {visible.length > 0 ? (
                    <div className='ml-model-grid'>
                      {visible.map((model) => (
                        <ModelCard key={model.model} model={model} />
                      ))}
                    </div>
                  ) : (
                    <EmptyState filtered={activeCount > 0} onReset={resetFilters} />
                  )}

                  <p className='ml-grid-note'>{t('models.gridNote')}</p>
                </div>
              </div>
            </>
          ) : null}
        </div>
      </main>

      <footer className='ml-foot'>
        <div className='ml-container ml-foot-inner'>
          <span>© 2026 CubeFace</span>
          <span className='ml-mono'>{t('models.footer.source')}</span>
        </div>
      </footer>
    </div>
  )
}
