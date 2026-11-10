import type { TFunction } from 'i18next'

import type { ModelSummary, StorageSummary } from './types'

export function modelDescription(model: ModelSummary, t: TFunction): string {
  if (model.description) return model.description
  return t('models.fallbackDescription', {
    architecture: model.architecture || t('models.fields.unknown'),
    quantization: model.quantization || t('models.fields.unknown'),
    count: model.version_count,
  })
}

export function storageSummaryText(storage: StorageSummary, t: TFunction): string {
  const strategy = storage.strategy || t('models.fields.unknown')
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

export function servingModeLabel(mode: string | undefined, t: TFunction): string {
  if (mode === 'pd-separation') return t('models.fields.pdSeparation')
  if (mode === 'standard') return t('models.fields.standard')
  return '—'
}

export function triStateLabel(value: boolean | null, t: TFunction): string {
  if (value === true) return t('models.fields.yes')
  if (value === false) return t('models.fields.no')
  return t('models.fields.unknown')
}

export function formatList(values: string[] | undefined): string {
  return values && values.length > 0 ? values.join(', ') : '—'
}
