const MINUTE = 60
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR
const WEEK = 7 * DAY
const MONTH = 30 * DAY
const YEAR = 365 * DAY

function safeLocale(locale: string): string {
  return locale === 'zh' ? 'zh-CN' : 'en'
}

export function formatDate(value: string | null | undefined, locale: string): string {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return new Intl.DateTimeFormat(safeLocale(locale), {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(date)
}

export function formatRelativeTime(
  value: string | null | undefined,
  locale: string,
  now: number = Date.now()
): string {
  if (!value) return '—'
  const date = new Date(value)
  const time = date.getTime()
  if (Number.isNaN(time)) return '—'

  const seconds = Math.round((time - now) / 1000)
  const absolute = Math.abs(seconds)
  if (absolute < 45) {
    return new Intl.RelativeTimeFormat(safeLocale(locale), { numeric: 'auto' }).format(
      0,
      'second'
    )
  }
  if (absolute > YEAR) {
    return formatDate(value, locale)
  }

  const formatter = new Intl.RelativeTimeFormat(safeLocale(locale), {
    numeric: 'auto',
  })
  const [amount, unit]: [number, Intl.RelativeTimeFormatUnit] =
    absolute < HOUR
      ? [Math.round(seconds / MINUTE), 'minute']
      : absolute < DAY
        ? [Math.round(seconds / HOUR), 'hour']
        : absolute < WEEK
          ? [Math.round(seconds / DAY), 'day']
          : absolute < MONTH
            ? [Math.round(seconds / WEEK), 'week']
            : absolute < YEAR
              ? [Math.round(seconds / MONTH), 'month']
              : [Math.round(seconds / YEAR), 'year']
  return formatter.format(amount, unit)
}

export function formatNumber(value: number, locale: string): string {
  return new Intl.NumberFormat(safeLocale(locale), {
    maximumFractionDigits: 0,
  }).format(value)
}

export function initials(publisher: string, fallback: string): string {
  const source = publisher.trim() || fallback.trim()
  if (!source) return '?'
  const fields = source.split(/\s+/).filter(Boolean)
  if (fields.length >= 2) {
    return (fields[0][0] + fields[1][0]).toUpperCase()
  }
  const runes = [...source]
  if (runes.length === 1) return runes[0].toUpperCase()
  return runes.slice(0, 2).join('').toUpperCase()
}
