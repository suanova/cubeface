import { getRouteApi, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SiteHeader } from '@/components/site-header'

import { ModelNotFoundError } from '../api'
import { modelDetailQueryOptions } from '../queries'
import { DetailHeader } from './detail-header'
import { DetailTabs, type DetailTab } from './detail-tabs'
import { DetailLoading } from './loading-states'
import { ModelInfoCard } from './model-info-card'
import { OverviewPanel } from './overview-panel'
import { RuntimeProfileList } from './runtime-profile-list'
import { UnavailableState } from './unavailable-state'
import { VersionList } from './version-list'

const routeApi = getRouteApi('/models/$name')

export function ModelDetailPage() {
  const { t } = useTranslation()
  const { name } = routeApi.useParams()
  const { data, isPending, isError, error, refetch } = useQuery(
    modelDetailQueryOptions(name)
  )
  const [activeTab, setActiveTab] = useState<DetailTab>('overview')
  const [copied, setCopied] = useState(false)
  const timer = useRef<number | undefined>(undefined)

  useEffect(() => () => window.clearTimeout(timer.current), [])

  const copyName = () => {
    if (!data) return
    const done = () => {
      setCopied(true)
      toast.success(t('models.detail.copied'))
      window.clearTimeout(timer.current)
      timer.current = window.setTimeout(() => setCopied(false), 2000)
    }
    if (!navigator.clipboard) {
      toast.error(t('models.detail.copyFailed'))
      return
    }
    void navigator.clipboard.writeText(data.model).then(done, () => {
      toast.error(t('models.detail.copyFailed'))
    })
  }

  const notFound = isError && error instanceof ModelNotFoundError

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
              <DetailLoading />
            </>
          ) : null}

          {notFound ? (
            <div className='ml-state-card is-wide' role='status' style={{ marginTop: 40 }}>
              <p className='ml-state-kicker'>404</p>
              <h1 className='ml-state-title'>{t('models.states.notFound.title')}</h1>
              <p className='ml-state-text'>{t('models.states.notFound.text')}</p>
              <div className='ml-state-actions'>
                <Link to='/models' className='ml-btn ml-btn-secondary'>
                  <ArrowLeft aria-hidden='true' />
                  {t('models.states.notFound.back')}
                </Link>
              </div>
            </div>
          ) : null}

          {isError && !notFound ? (
            <div style={{ paddingTop: 40 }}>
              <UnavailableState
                reason='API_ERROR'
                retryable
                onRetry={() => void refetch()}
              />
            </div>
          ) : null}

          {data && !data.available ? (
            <div style={{ paddingTop: 40 }}>
              <UnavailableState
                reason={data.reason}
                retryable={data.retryable}
                onRetry={() => void refetch()}
              />
            </div>
          ) : null}

          {data?.available ? (
            <>
              <DetailHeader detail={data} copied={copied} onCopy={copyName} />

              <DetailTabs
                active={activeTab}
                onChange={setActiveTab}
                versionCount={data.versions.length}
                runtimeCount={data.compatible_profiles.length}
              />

              <div className='ml-detail-grid'>
                {activeTab === 'overview' ? <OverviewPanel detail={data} /> : null}

                {activeTab === 'runtimes' ? (
                  <section
                    className='ml-panel'
                    role='tabpanel'
                    id='ml-panel-runtimes'
                    aria-labelledby='ml-tab-runtimes'
                  >
                    <div className='ml-panel-head'>
                      <h2>{t('models.detail.runtimeTitle')}</h2>
                      <span className='ml-meta'>
                        {data.profiles_available
                          ? t('models.cards.runtimes', {
                              count: data.compatible_profiles.length,
                            })
                          : t('models.fields.unknown')}
                      </span>
                    </div>
                    <RuntimeProfileList
                      profiles={data.compatible_profiles}
                      profilesAvailable={data.profiles_available}
                    />
                  </section>
                ) : null}

                {activeTab === 'versions' ? (
                  <section
                    className='ml-panel'
                    role='tabpanel'
                    id='ml-panel-versions'
                    aria-labelledby='ml-tab-versions'
                  >
                    <div className='ml-panel-head'>
                      <h2>{t('models.detail.versionsTitle')}</h2>
                      <span className='ml-meta'>
                        {t('models.headNotes.versions', { count: data.versions.length })}
                      </span>
                    </div>
                    <VersionList
                      versions={data.versions}
                      profilesAvailable={data.profiles_available}
                    />
                  </section>
                ) : null}

                <aside className='ml-detail-side'>
                  <ModelInfoCard detail={data} />
                  <div className='ml-side-card'>
                    <h3>{t('models.detail.sideTags')}</h3>
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
                          {t('models.fields.notAnnotated')}
                        </span>
                      ) : null}
                    </div>
                    <p className='ml-sub-label'>{t('models.fields.license')}</p>
                    <p className='ml-side-note'>
                      {data.license || t('models.fields.notAnnotated')}
                    </p>
                    <p className='ml-sub-label'>{t('models.fields.usage')}</p>
                    <p className='ml-side-note'>
                      {data.usage_boundary || t('models.fields.notAnnotated')}
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
          <span className='ml-mono'>{t('models.footer.source')}</span>
        </div>
      </footer>
    </div>
  )
}
