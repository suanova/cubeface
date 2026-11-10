import type { TFunction } from 'i18next'

import { formatNumber } from '@/features/models/format'

import type { DatasetSummary, StorageSummary } from './types'

export function datasetDescription(
  dataset: DatasetSummary,
  locale: string,
  t: TFunction
): string {
  if (dataset.description) return dataset.description
  const format = dataset.format || t('datasets.fields.unknown')
  if (dataset.record_count != null) {
    return t('datasets.fallbackDescription', {
      format,
      count: dataset.version_count,
      rows: formatNumber(dataset.record_count, locale),
    })
  }
  return t('datasets.fallbackDescriptionNoRows', {
    format,
    count: dataset.version_count,
  })
}

export function recordCountLabel(
  value: number | undefined,
  locale: string,
  t: TFunction
): string {
  if (value == null) return t('datasets.fields.unknown')
  return t('datasets.rows.records', { rows: formatNumber(value, locale) })
}

export function storageSummaryText(storage: StorageSummary, t: TFunction): string {
  const strategy = storage.strategy || t('datasets.fields.unknown')
  const details: string[] = []
  if (storage.storage_class) details.push(storage.storage_class)
  if (storage.capacity) details.push(storage.capacity)
  if (storage.path_configured) details.push(t('models.fields.pathConfigured'))
  if (storage.uri_configured) details.push(t('models.fields.uriConfigured'))
  if (storage.credentials_configured) {
    details.push(t('models.fields.credentialsConfigured'))
  }
  return details.length > 0 ? `${strategy} · ${details.join(' · ')}` : strategy
}
