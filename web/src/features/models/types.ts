// Wire types for GET /api/models and GET /api/models/:name. They mirror the
// sanitized DTOs in internal/service/modelcatalog_dto.go: no raw storage paths,
// S3 URIs, credential names or profile environment values are exposed.

export type CatalogReason =
  | 'DISABLED'
  | 'NO_CLUSTER'
  | 'CRD_MISSING'
  | 'FORBIDDEN'
  | 'UNAUTHORIZED'
  | 'TIMEOUT'
  | 'API_ERROR'

export interface CatalogStatus {
  available: boolean
  reason?: CatalogReason | string
  message?: string
  retryable: boolean
  stale: boolean
  warning?: string
  profiles_available: boolean
  generated_at: string
}

export interface ModelSummary {
  model: string
  display_name: string
  publisher: string
  publisher_initials: string
  description: string
  architecture: string
  quantization: string
  task: string
  license: string
  tags: string[]
  usage_boundary: string
  storage_strategy: string
  latest_version: string
  version_count: number
  compatible_profile_count: number
  used_by_count: number
  featured: boolean
  created_at: string
  updated_at: string
}

export interface ModelList extends CatalogStatus {
  total: number
  version_total: number
  profile_total: number
  models: ModelSummary[]
}

export interface StorageSummary {
  strategy: string
  storage_class?: string
  capacity?: string
  path_configured: boolean
  uri_configured: boolean
  credentials_configured: boolean
}

export interface UsedByRef {
  namespace: string
  name: string
  version?: string
}

export interface EngineSummary {
  name: string
  version: string
}

export interface ProfileRoleSummary {
  name: string
  vendor: string
  accelerator_models: string[]
  gpu_per_pod: number | null
}

export interface RuntimeProfileSummary {
  name: string
  description: string
  engine: EngineSummary
  vendor: string
  accelerator_models: string[]
  architectures: string[]
  quantizations: string[]
  serving_mode?: 'standard' | 'pd-separation' | string
  gpu_per_pod: number | null
  roles: ProfileRoleSummary[]
  compatible_versions: string[]
  assets_resolved: boolean | null
  deprecated: boolean | null
  drifted: boolean | null
  used_by_count: number
}

export interface ModelVersionDetail {
  ref: string
  version: string
  is_latest: boolean
  architecture: string
  quantization: string
  description: string
  storage: StorageSummary
  storage_resolved: boolean | null
  in_use: boolean | null
  root_path_resolved: boolean
  used_by: UsedByRef[]
  compatible_profiles: string[]
  created_at: string
}

export interface ModelDetail extends CatalogStatus, ModelSummary {
  latest_ref: string
  versions: ModelVersionDetail[]
  compatible_profiles: RuntimeProfileSummary[]
  used_by: UsedByRef[]
}
