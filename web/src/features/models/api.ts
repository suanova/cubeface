import axios from 'axios'

import { api, type ApiEnvelope } from '@/lib/api'

import type { ModelDetail, ModelList } from './types'

export class ModelNotFoundError extends Error {
  constructor() {
    super('Model not found')
    this.name = 'ModelNotFoundError'
  }
}

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  if (!envelope.success || envelope.data === undefined) {
    throw new Error(envelope.message || 'The request could not be completed.')
  }
  return envelope.data
}

export async function fetchModelList(): Promise<ModelList> {
  const response = await api.get<ApiEnvelope<ModelList>>('/models')
  return unwrap(response.data)
}

export async function fetchModelDetail(name: string): Promise<ModelDetail> {
  try {
    const response = await api.get<ApiEnvelope<ModelDetail>>(
      `/models/${encodeURIComponent(name)}`
    )
    return unwrap(response.data)
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 404) {
      throw new ModelNotFoundError()
    }
    throw error
  }
}
