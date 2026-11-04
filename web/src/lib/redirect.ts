const FALLBACK = '/dashboard'

// safeRedirect only allows same-origin absolute paths, matching the server's
// OAuth redirect policy. Protocol-relative and backslash paths are rejected.
export function safeRedirect(raw: string | null | undefined): string {
  const value = raw?.trim()
  if (!value) return FALLBACK
  if (!value.startsWith('/') || value.startsWith('//')) return FALLBACK
  if (value.includes('\\') || /[\r\n\t]/.test(value)) return FALLBACK
  return value
}
