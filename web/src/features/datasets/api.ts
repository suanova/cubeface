import axios from 'axios'

import { api, type ApiEnvelope } from '@/lib/api'

import type { DatasetDetail, DatasetList } from './types'

export class DatasetNotFoundError extends Error {
  constructor() {
    super('Dataset not found')
    this.name = 'DatasetNotFoundError'
  }
}

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  if (!envelope.success || envelope.data === undefined) {
    throw new Error(envelope.message || 'The request could not be completed.')
  }
  return envelope.data
}

export async function fetchDatasetList(): Promise<DatasetList> {
  const response = await api.get<ApiEnvelope<DatasetList>>('/datasets')
  return unwrap(response.data)
}

export async function fetchDatasetDetail(name: string): Promise<DatasetDetail> {
  try {
    const response = await api.get<ApiEnvelope<DatasetDetail>>(
      `/datasets/${encodeURIComponent(name)}`
    )
    return unwrap(response.data)
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 404) {
      throw new DatasetNotFoundError()
    }
    throw error
  }
}
