/** Newest-first view of a catalog list, for the studio's recent panels. */
export function pickRecent<T>(
  items: readonly T[],
  updatedAt: (item: T) => string,
  limit = 4
): T[] {
  return [...items]
    .sort((a, b) => timestamp(updatedAt(b)) - timestamp(updatedAt(a)))
    .slice(0, limit)
}

function timestamp(value: string): number {
  const time = new Date(value).getTime()
  return Number.isNaN(time) ? 0 : time
}
