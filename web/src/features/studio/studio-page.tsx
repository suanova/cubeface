import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { ArrowRight } from 'lucide-react'
import { useMemo } from 'react'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { SiteHeader } from '@/components/site-header'
import { datasetListQueryOptions } from '@/features/datasets/queries'
import { PageHead } from '@/features/models/components/page-head'
import { formatRelativeTime } from '@/features/models/format'
import { modelListQueryOptions } from '@/features/models/queries'

import { pickRecent } from './recent'

type CatalogState = 'pending' | 'connected' | 'unavailable'

interface CatalogLike {
  isPending: boolean
  isError: boolean
  data?: { available: boolean; reason?: string; total?: number }
}

function catalogState(query: CatalogLike): CatalogState {
  if (query.isPending) return 'pending'
  if (query.isError || !query.data?.available) return 'unavailable'
  return 'connected'
}

export function StudioPage() {
  const { t, i18n } = useTranslation()
  const locale = i18n.language.startsWith('zh') ? 'zh' : 'en'

  const models = useQuery(modelListQueryOptions)
  const datasets = useQuery(datasetListQueryOptions)

  const modelState = catalogState(models)
  const datasetState = catalogState(datasets)

  const recentModels = useMemo(
    () =>
      models.data?.available
        ? pickRecent(models.data.models, (model) => model.updated_at)
        : [],
    [models.data]
  )
  const recentDatasets = useMemo(
    () =>
      datasets.data?.available
        ? pickRecent(datasets.data.datasets, (dataset) => dataset.updated_at)
        : [],
    [datasets.data]
  )

  const notes = [
    ...(models.data?.available
      ? [t('studio.headNotes.models', { count: models.data.total })]
      : []),
    ...(datasets.data?.available
      ? [t('studio.headNotes.datasets', { count: datasets.data.total })]
      : []),
    t('studio.headNotes.runtime'),
  ]

  const startMeta = (state: CatalogState, total: number | undefined, key: string) =>
    state === 'connected'
      ? t(key, { count: total ?? 0 })
      : t('studio.start.unavailable')

  return (
    <div className='ml-scope'>
      <SiteHeader />
      <main className='ml-main'>
        <div className='ml-container'>
          <PageHead
            kicker={t('studio.kicker')}
            title={t('studio.title')}
            lead={t('studio.lead')}
            notes={notes}
            actions={
              <>
                <Link to='/models' className='ml-btn ml-btn-primary'>
                  {t('studio.actions.models')}
                </Link>
                <Link to='/datasets' className='ml-btn ml-btn-secondary'>
                  {t('studio.actions.datasets')}
                </Link>
              </>
            }
          />

          <section className='st-console' aria-labelledby='st-console-title'>
            <div className='st-console-head'>
              <div>
                <p className='st-console-kicker'>{t('studio.console.kicker')}</p>
                <h2 className='st-console-title' id='st-console-title'>
                  {t('studio.console.title')}
                </h2>
              </div>
              <span className='st-pill'>
                <i className='st-dot' aria-hidden='true' />
                {t('studio.console.status')}
              </span>
            </div>
            <p className='st-console-text'>{t('studio.console.text')}</p>
            <dl className='st-console-grid'>
              <div className='st-console-stat'>
                <dt>{t('studio.console.fields.inference')}</dt>
                <dd>InferenceService</dd>
              </div>
              <div className='st-console-stat'>
                <dt>{t('studio.console.fields.workspace')}</dt>
                <dd>DevEnvironment</dd>
              </div>
              <div className='st-console-stat'>
                <dt>{t('studio.console.fields.quota')}</dt>
                <dd>{t('studio.console.values.planned')}</dd>
              </div>
              <div className='st-console-stat'>
                <dt>{t('studio.console.fields.logs')}</dt>
                <dd>{t('studio.console.values.planned')}</dd>
              </div>
            </dl>
          </section>

          <div className='st-grid'>
            <div className='st-col'>
              <RecentPanel
                title={t('studio.recent.models')}
                state={modelState}
                count={
                  models.data?.available
                    ? t('studio.recent.count', { count: models.data.total })
                    : undefined
                }
                reason={models.data?.reason}
              >
                {recentModels.map((model) => (
                  <Link
                    key={model.model}
                    to='/models/$name'
                    params={{ name: model.model }}
                    className='st-recent-row'
                  >
                    <span className='st-recent-main'>
                      <span className='st-recent-name'>
                        {model.display_name || model.model}
                      </span>
                      <span className='st-recent-meta'>
                        {[model.latest_version, model.publisher]
                          .filter(Boolean)
                          .join(' · ')}
                      </span>
                    </span>
                    <span className='st-recent-time'>
                      {formatRelativeTime(model.updated_at, locale)}
                    </span>
                    <ArrowRight className='st-recent-arrow' aria-hidden='true' />
                  </Link>
                ))}
              </RecentPanel>

              <RecentPanel
                title={t('studio.recent.datasets')}
                state={datasetState}
                count={
                  datasets.data?.available
                    ? t('studio.recent.count', { count: datasets.data.total })
                    : undefined
                }
                reason={datasets.data?.reason}
              >
                {recentDatasets.map((dataset) => (
                  <Link
                    key={dataset.dataset}
                    to='/datasets/$name'
                    params={{ name: dataset.dataset }}
                    className='st-recent-row'
                  >
                    <span className='st-recent-main'>
                      <span className='st-recent-name'>
                        {dataset.display_name || dataset.dataset}
                      </span>
                      <span className='st-recent-meta'>
                        {[dataset.format, dataset.publisher]
                          .filter(Boolean)
                          .join(' · ')}
                      </span>
                    </span>
                    <span className='st-recent-time'>
                      {formatRelativeTime(dataset.updated_at, locale)}
                    </span>
                    <ArrowRight className='st-recent-arrow' aria-hidden='true' />
                  </Link>
                ))}
              </RecentPanel>
            </div>

            <aside className='st-status-card' aria-labelledby='st-status-title'>
              <p className='st-side-kicker'>{t('studio.status.kicker')}</p>
              <h2 id='st-status-title'>{t('studio.status.title')}</h2>
              <ul className='st-status-list'>
                <StatusRow
                  label={t('studio.status.models')}
                  state={modelState}
                />
                <StatusRow
                  label={t('studio.status.datasets')}
                  state={datasetState}
                />
                <StatusRow
                  label={t('studio.status.runtime')}
                  state='notConnected'
                />
              </ul>
              <p className='st-side-note'>{t('studio.status.note')}</p>
            </aside>
          </div>

          <section className='st-start' aria-labelledby='st-start-title'>
            <div className='st-section-head is-bordered'>
              <h2 id='st-start-title'>{t('studio.start.title')}</h2>
              <span className='st-count'>{t('studio.start.note')}</span>
            </div>

            <Link to='/models' className='st-start-row'>
              <span className='st-start-idx'>01</span>
              <span className='st-start-main'>
                <span className='st-start-title'>
                  {t('studio.start.models.title')}
                </span>
                <span className='st-start-text'>
                  {t('studio.start.models.text')}
                </span>
              </span>
              <span className='st-start-meta'>
                {startMeta(modelState, models.data?.total, 'studio.start.models.meta')}
              </span>
              <span className='st-start-cta'>
                {t('studio.start.models.cta')}
                <ArrowRight aria-hidden='true' />
              </span>
            </Link>

            <Link to='/datasets' className='st-start-row'>
              <span className='st-start-idx'>02</span>
              <span className='st-start-main'>
                <span className='st-start-title'>
                  {t('studio.start.datasets.title')}
                </span>
                <span className='st-start-text'>
                  {t('studio.start.datasets.text')}
                </span>
              </span>
              <span className='st-start-meta'>
                {startMeta(
                  datasetState,
                  datasets.data?.total,
                  'studio.start.datasets.meta'
                )}
              </span>
              <span className='st-start-cta'>
                {t('studio.start.datasets.cta')}
                <ArrowRight aria-hidden='true' />
              </span>
            </Link>

            <Link to='/dashboard' className='st-start-row'>
              <span className='st-start-idx'>03</span>
              <span className='st-start-main'>
                <span className='st-start-title'>
                  {t('studio.start.dashboard.title')}
                </span>
                <span className='st-start-text'>
                  {t('studio.start.dashboard.text')}
                </span>
              </span>
              <span className='st-start-meta'>
                {t('studio.start.dashboard.meta')}
              </span>
              <span className='st-start-cta'>
                {t('studio.start.dashboard.cta')}
                <ArrowRight aria-hidden='true' />
              </span>
            </Link>
          </section>
        </div>
      </main>

      <footer className='ml-foot'>
        <div className='ml-container ml-foot-inner'>
          <span>© 2026 CubeFace</span>
          <span className='ml-mono'>{t('studio.footer.source')}</span>
        </div>
      </footer>
    </div>
  )
}

function RecentPanel({
  title,
  count,
  state,
  reason,
  children,
}: {
  title: string
  count?: string
  state: CatalogState
  reason?: string
  children: ReactNode
}) {
  const { t } = useTranslation()
  const hasRows = Array.isArray(children) ? children.length > 0 : Boolean(children)

  let body: ReactNode
  if (state === 'pending') {
    body = <RecentSkeleton />
  } else if (state === 'unavailable') {
    body = (
      <div className='st-recent-list'>
        <p className='st-inline-note'>
          {t('studio.recent.unavailable', { reason: reason || 'API_ERROR' })}
        </p>
      </div>
    )
  } else if (hasRows) {
    body = <div className='st-recent-list'>{children}</div>
  } else {
    body = (
      <div className='st-recent-list'>
        <p className='st-inline-note'>{t('studio.recent.empty')}</p>
      </div>
    )
  }

  return (
    <section className='st-panel' aria-busy={state === 'pending'}>
      <div className='st-section-head'>
        <h3>{title}</h3>
        {count ? <span className='st-count'>{count}</span> : null}
      </div>
      {body}
    </section>
  )
}

function RecentSkeleton() {
  return (
    <div className='st-recent-list' aria-hidden='true'>
      {[0, 1, 2].map((row) => (
        <div key={row} className='st-recent-row is-skeleton'>
          <span className='st-recent-main'>
            <span className='st-skel-line' />
            <span className='st-skel-line is-short' />
          </span>
        </div>
      ))}
    </div>
  )
}

const STATUS_TONES: Record<CatalogState | 'notConnected', string> = {
  connected: 'is-on',
  pending: 'is-idle',
  unavailable: 'is-warn',
  notConnected: 'is-idle',
}

function StatusRow({
  label,
  state,
}: {
  label: string
  state: CatalogState | 'notConnected'
}) {
  const { t } = useTranslation()
  const tone = STATUS_TONES[state]

  return (
    <li className='st-status-row'>
      <span className={`st-dot ${tone}`} aria-hidden='true' />
      <span className='st-status-label'>{label}</span>
      <span className='st-status-value'>
        {t(`studio.status.states.${state}`)}
      </span>
    </li>
  )
}
