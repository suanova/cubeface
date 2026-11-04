export interface AuthUser {
  id: number
  username: string
  display_name: string
  email: string
  role: number
  status: number
  email_verified: boolean
  created_at: string
  last_login_at: string | null
}

export interface OAuthProviderInfo {
  id: string
  name: string
}

export interface StatusInfo {
  system_name: string
  version: string
  register_enabled: boolean
  email_verification_enabled: boolean
  oauth_providers: OAuthProviderInfo[]
  server_time: string
}

export interface LoginInput {
  username: string
  password: string
}

export interface RegisterInput {
  username: string
  email: string
  password: string
  verification_code: string
}
