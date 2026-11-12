import { getRouteApi, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SiteHeader } from '@/components/site-header'
import { DetailLoading } from '@/features/models/components/loading-states'
import { UnavailableState } from '@/features/models/components/unavailable-state'

import { DatasetNotFoundError } from '../api'
import { datasetDetailQueryOptions } from '../queries'
import { DatasetDetailHeader } from './dataset-detail-header'
import { DatasetDetailTabs, type DatasetTab } from './dataset-detail-tabs'
import { DatasetInfoCard } from './dataset-info-card'
import { DatasetOverviewPanel } from './dataset-overview-panel'
import { DatasetVersionList } from './dataset-version-list'

const routeApi = getRouteApi('/datasets/$name')

export function DatasetDetailPage() {
  const { t } = useTranslation()
  const { name } = routeApi.useParams()
  const { data, isPending, isError, error, refetch } = useQuery(
    datasetDetailQueryOptions(name)
  )
  const [activeTab, setActiveTab] = useState<DatasetTab>('overview')
  const [copied, setCopied] = useState(false)
  const timer = useRef<number | undefined>(undefined)

  useEffect(() => () => window.clearTimeout(timer.current), [])

  const copyName = () => {
    if (!data) return
    const done = () => {
      setCopied(true)
      toast.success(t('datasets.detail.copied'))
      window.clearTimeout(timer.current)
      timer.current = window.setTimeout(() => setCopied(false), 2000)
    }
    if (!navigator.clipboard) {
      toast.error(t('datasets.detail.copyFailed'))
      return
    }
    void navigator.clipboard.writeText(data.dataset).then(done, () => {
      toast.error(t('datasets.detail.copyFailed'))
    })
  }

  const notFound = isError && error instanceof DatasetNotFoundError

  return (
    <div className='ml-scope'>
      <SiteHeader />
      <main className='ml-main'>
        <div className='ml-container'>
          {isPending ? (
            <>
              <div style={{ paddingTop: 26 }}>
                <span className='ml-skeleton' style={{ width: 220, height: 12 }} />
              </div>
              <div style={{ padding: '24px 0' }}>
                <span className='ml-skeleton' style={{ width: 300, height: 32 }} />
              </div>
              <DetailLoading namespace='datasets' />
            </>
          ) : null}

          {notFound ? (
            <div className='ml-state-card is-wide' role='status' style={{ marginTop: 40 }}>
              <p className='ml-state-kicker'>404</p>
              <h1 className='ml-state-title'>{t('datasets.states.notFound.title')}</h1>
              <p className='ml-state-text'>{t('datasets.states.notFound.text')}</p>
              <div className='ml-state-actions'>
                <Link to='/datasets' className='ml-btn ml-btn-secondary'>
                  <ArrowLeft aria-hidden='true' />
                  {t('datasets.states.notFound.back')}
                </Link>
              </div>
            </div>
          ) : null}

          {isError && !notFound ? (
            <div style={{ paddingTop: 40 }}>
              <UnavailableState
                reason='API_ERROR'
                retryable
                namespace='datasets'
                onRetry={() => void refetch()}
              />
            </div>
          ) : null}

          {data && !data.available ? (
            <div style={{ paddingTop: 40 }}>
              <UnavailableState
                reason={data.reason}
                retryable={data.retryable}
                namespace='datasets'
                onRetry={() => void refetch()}
              />
            </div>
          ) : null}

          {data?.available ? (
            <>
              <DatasetDetailHeader detail={data} copied={copied} onCopy={copyName} />

              <DatasetDetailTabs
                active={activeTab}
                onChange={setActiveTab}
                versionCount={data.versions.length}
              />

              <div className='ml-detail-grid'>
                {activeTab === 'overview' ? (
                  <DatasetOverviewPanel detail={data} />
                ) : null}

                {activeTab === 'versions' ? (
                  <section
                    className='ml-panel'
                    role='tabpanel'
                    id='ds-panel-versions'
                    aria-labelledby='ds-tab-versions'
                  >
                    <div className='ml-panel-head'>
                      <h2>{t('datasets.detail.versionsTitle')}</h2>
                      <span className='ml-meta'>
                        {t('datasets.headNotes.versions', {
                          count: data.versions.length,
                        })}
                      </span>
                    </div>
                    <DatasetVersionList versions={data.versions} />
                  </section>
                ) : null}

                <aside className='ml-detail-side'>
                  <DatasetInfoCard detail={data} />
                  <div className='ml-side-card'>
                    <h3>{t('datasets.detail.sideTags')}</h3>
                    <div className='ml-side-tags'>
                      {data.usage_boundary ? (
                        <span className='ml-usage-badge'>{data.usage_boundary}</span>
                      ) : null}
                      {data.tags.map((tag) => (
                        <span className='ml-tag' key={tag}>
                          {tag}
                        </span>
                      ))}
                      {!data.usage_boundary && data.tags.length === 0 ? (
                        <span className='ml-side-note'>
                          {t('datasets.fields.notAnnotated')}
                        </span>
                      ) : null}
                    </div>
                    <p className='ml-sub-label'>{t('datasets.fields.license')}</p>
                    <p className='ml-side-note'>
                      {data.license || t('datasets.fields.notAnnotated')}
                    </p>
                    <p className='ml-sub-label'>{t('datasets.fields.access')}</p>
                    <p className='ml-side-note'>
                      {data.access || t('datasets.fields.notAnnotated')}
                    </p>
                    <p className='ml-sub-label'>{t('datasets.fields.usage')}</p>
                    <p className='ml-side-note'>
                      {data.usage_boundary || t('datasets.fields.notAnnotated')}
                    </p>
                  </div>
                </aside>
              </div>
            </>
          ) : null}
        </div>
      </main>

      <footer className='ml-foot'>
        <div className='ml-container ml-foot-inner'>
          <span>© 2026 MedRouter</span>
          <span className='ml-mono'>{t('datasets.footer.source')}</span>
        </div>
      </footer>
    </div>
  )
}
