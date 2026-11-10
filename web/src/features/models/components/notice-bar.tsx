import type { ReactNode } from 'react'

interface NoticeBarProps {
  label: string
  children: ReactNode
  warning?: boolean
}

export function NoticeBar({ label, children, warning }: NoticeBarProps) {
  return (
    <p className={warning ? 'ml-notice is-warning' : 'ml-notice'}>
      <b>{label}</b>
      <span>{children}</span>
    </p>
  )
}
