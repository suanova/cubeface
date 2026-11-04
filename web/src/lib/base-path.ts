declare global {
  interface Window {
    __BASE_PATH__?: string
  }
}

// The Go server injects window.__BASE_PATH__ into index.html so one build can
// be served from the site root or from a prefix such as /cubeface.
export const basePath =
  typeof window !== 'undefined' ? (window.__BASE_PATH__ ?? '').replace(/\/+$/, '') : ''

export function withBasePath(path: string): string {
  const normalized = path.startsWith('/') ? path : `/${path}`
  return `${basePath}${normalized}`
}
